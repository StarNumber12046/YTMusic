package ytmusic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"ytmusic_api/models"
)

const (
	baseURL       = "https://music.youtube.com/youtubei/v1"
	userAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	apiKey        = "AIzaSyC9XL3ZjWddXya6X74dJoCTL-WEYFDNX30" // public Innertube key for WEB_REMIX
	clientName    = "WEB_REMIX"
	clientVersion = "1.20241127.01.00"
)

// Client talks to the YouTube Music Innertube API.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new Innertube client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// innertubeContext is the standard context payload required by every Innertube request.
func innertubeContext() map[string]interface{} {
	return map[string]interface{}{
		"client": map[string]interface{}{
			"clientName":    clientName,
			"clientVersion": clientVersion,
			"hl":            "en",
			"gl":            "US",
		},
	}
}

// doRequest performs an authenticated Innertube POST request.
func (c *Client) doRequest(session *Session, endpoint string, body map[string]interface{}) (map[string]interface{}, error) {
	// Inject the context into the body
	body["context"] = innertubeContext()

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	url := fmt.Sprintf("%s/%s?key=%s&prettyPrint=false", baseURL, endpoint, apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "https://music.youtube.com/")
	req.Header.Set("Origin", ytMusicOrigin)
	req.Header.Set("X-Goog-AuthUser", "0")
	req.Header.Set("X-Origin", ytMusicOrigin)

	if session != nil && session.SAPISID != "" {
		req.Header.Set("Authorization", GetAuthorizationHeader(session.SAPISID))
		req.Header.Set("Cookie", session.Cookies)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("innertube returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result, nil
}

// Search performs a search query on YouTube Music.
func (c *Client) Search(session *Session, query string, filter string) ([]models.SearchResult, error) {
	body := map[string]interface{}{
		"query": query,
	}

	// Set search params based on filter type
	if filter != "" {
		params := getSearchParams(filter)
		if params != "" {
			body["params"] = params
		}
	}

	raw, err := c.doRequest(session, "search", body)
	if err != nil {
		return nil, err
	}

	return parseSearchResults(raw, filter), nil
}

// GetPlaylist retrieves a playlist's tracks.
func (c *Client) GetPlaylist(session *Session, playlistID string) (*models.PlaylistDetailResponse, error) {
	browseID := playlistID
	if len(playlistID) > 0 && playlistID[:2] != "VL" {
		browseID = "VL" + playlistID
	}

	body := map[string]interface{}{
		"browseId": browseID,
	}

	raw, err := c.doRequest(session, "browse", body)
	if err != nil {
		return nil, err
	}

	// Debug: log the raw response
	if jsonBytes, marshalErr := json.MarshalIndent(raw, "", "  "); marshalErr == nil {
		slog.Info("===== PLAYLIST DETAIL RAW RESPONSE =====", "playlistID", playlistID, "json", string(jsonBytes))
	}

	return parsePlaylistResponse(raw, playlistID), nil
}

// GetLibraryPlaylists retrieves the user's library playlists.
func (c *Client) GetLibraryPlaylists(session *Session) ([]models.Playlist, error) {
	body := map[string]interface{}{
		"browseId": "FEmusic_liked_playlists",
	}

	raw, err := c.doRequest(session, "browse", body)
	if err != nil {
		return nil, err
	}

	// Debug: log the raw response as JSON to see structure
	if jsonBytes, marshalErr := json.MarshalIndent(raw, "", "  "); marshalErr == nil {
		slog.Info("===== PLAYLISTS RAW RESPONSE =====", "json", string(jsonBytes))
	}

	playlists := parseLibraryPlaylists(raw)
	slog.Info("Got playlists", "count", len(playlists))
	return playlists, nil
}

// GetUserInfo retrieves account information.
func (c *Client) GetUserInfo(session *Session) (*models.UserInfoResponse, error) {
	body := map[string]interface{}{}

	raw, err := c.doRequest(session, "account/list", body)
	if err != nil {
		return nil, err
	}

	return parseUserInfo(raw), nil
}

// GetWatchPlaylist retrieves the up-next / queue tracks for a given video.
func (c *Client) GetWatchPlaylist(session *Session, videoID string, playlistID string) ([]models.Track, error) {
	body := map[string]interface{}{
		"enablePersistentPlaylistPanel": true,
		"isAudioOnly":                   true,
		"videoId":                       videoID,
	}
	if playlistID != "" {
		body["playlistId"] = playlistID
	}

	raw, err := c.doRequest(session, "next", body)
	if err != nil {
		return nil, err
	}

	return parseWatchPlaylist(raw), nil
}

// GetSongInfo retrieves basic info about a song (title, artist, etc.) from the player endpoint.
func (c *Client) GetSongInfo(session *Session, videoID string) (*models.Track, error) {
	body := map[string]interface{}{
		"videoId": videoID,
	}

	raw, err := c.doRequest(session, "player", body)
	if err != nil {
		return nil, err
	}

	return parseSongInfo(raw, videoID), nil
}

// --- Search params ---

// getSearchParams returns the encoded params string for filtering search results.
func getSearchParams(filter string) string {
	// These are the protobuf-encoded filter params used by YouTube Music.
	switch filter {
	case "songs":
		return "EgWKAQIIAWoKEAkQBRAKEAMQBA%3D%3D"
	case "videos":
		return "EgWKAQIQAWoKEAkQChAFEAMQBA%3D%3D"
	case "albums":
		return "EgWKAQIYAWoKEAkQChAFEAMQBA%3D%3D"
	case "artists":
		return "EgWKAQIgAWoKEAkQChAFEAMQBA%3D%3D"
	case "playlists":
		return "EgeKAQQoAEABagoQAxAEEAkQChAF"
	default:
		return ""
	}
}

// --- Response Parsers ---

// parseSearchResults extracts search results from an Innertube search response.
func parseSearchResults(raw map[string]interface{}, filter string) []models.SearchResult {
	var results []models.SearchResult

	contents := navigatePath(raw, "contents", "tabbedSearchResultsRenderer", "tabs")
	tabs, ok := contents.([]interface{})
	if !ok || len(tabs) == 0 {
		return results
	}

	tabContent := navigatePath(tabs[0], "tabRenderer", "content", "sectionListRenderer", "contents")
	sections, ok := tabContent.([]interface{})
	if !ok {
		return results
	}

	for _, section := range sections {
		shelfContents := navigatePath(section, "musicShelfRenderer", "contents")
		items, ok := shelfContents.([]interface{})
		if !ok {
			continue
		}

		for _, item := range items {
			result := parseSearchResultItem(item)
			if result != nil {
				results = append(results, *result)
			}
		}
	}

	return results
}

// parseSearchResultItem parses a single musicResponsiveListItemRenderer into a SearchResult.
func parseSearchResultItem(item interface{}) *models.SearchResult {
	renderer := navigatePath(item, "musicResponsiveListItemRenderer")
	if renderer == nil {
		return nil
	}
	rendererMap, ok := renderer.(map[string]interface{})
	if !ok {
		return nil
	}

	// Extract flex columns text
	columns := navigatePath(rendererMap, "flexColumns")
	colArr, ok := columns.([]interface{})
	if !ok || len(colArr) == 0 {
		return nil
	}

	title := extractFlexColumnText(colArr, 0)
	subtitle := extractFlexColumnText(colArr, 1)

	// Try to determine type from the item
	videoID := extractVideoID(rendererMap)
	browseID := extractBrowseID(rendererMap)
	thumbnail := extractThumbnail(rendererMap)

	if videoID != "" {
		// It's a song or video
		return &models.SearchResult{
			ResultType: "song",
			Track: &models.Track{
				VideoID:      videoID,
				Title:        title,
				Artist:       subtitle,
				ThumbnailURL: thumbnail,
			},
		}
	}

	if browseID != "" {
		// Could be album, artist, or playlist
		navType := extractNavigationType(rendererMap)
		switch {
		case containsStr(navType, "ARTIST") || containsStr(browseID, "UC"):
			return &models.SearchResult{
				ResultType: "artist",
				Artist: &models.Artist{
					Name:     title,
					BrowseID: browseID,
				},
			}
		case containsStr(navType, "ALBUM") || containsStr(browseID, "MPRE"):
			return &models.SearchResult{
				ResultType: "album",
				AlbumRef: &models.AlbumRef{
					BrowseID:     browseID,
					Title:        title,
					Artist:       subtitle,
					ThumbnailURL: thumbnail,
				},
			}
		default:
			return &models.SearchResult{
				ResultType: "playlist",
				Playlist: &models.Playlist{
					ID:           browseID,
					Title:        title,
					Author:       subtitle,
					ThumbnailURL: thumbnail,
				},
			}
		}
	}

	// Fallback: treat as song with just title
	return &models.SearchResult{
		ResultType: "song",
		Track: &models.Track{
			Title:  title,
			Artist: subtitle,
		},
	}
}

// parsePlaylistResponse parses an Innertube browse response for a playlist.
func parsePlaylistResponse(raw map[string]interface{}, playlistID string) *models.PlaylistDetailResponse {
	resp := &models.PlaylistDetailResponse{
		Playlist: models.Playlist{ID: playlistID},
	}

	// Extract header
	headerRenderer := navigatePath(raw, "header", "musicImmersiveHeaderRenderer")
	if headerRenderer == nil {
		headerRenderer = navigatePath(raw, "header", "musicDetailHeaderRenderer")
	}
	if headerRenderer == nil {
		slog.Debug("No header renderer found")
	}
	if headerMap, ok := headerRenderer.(map[string]interface{}); ok {
		resp.Playlist.Title = extractText(navigatePath(headerMap, "title"))
		resp.Playlist.Description = extractText(navigatePath(headerMap, "description"))
		resp.Playlist.ThumbnailURL = extractThumbnailFromObj(headerMap)
		slog.Debug("Parsed header", "title", resp.Playlist.Title)
	}

	// Extract tracks
	sectionContents := navigatePath(raw, "contents", "singleColumnBrowseResultsRenderer", "tabs")
	tabs, ok := sectionContents.([]interface{})
	if !ok || len(tabs) == 0 {
		slog.Debug("No tabs found in contents")
		return resp
	}

	tabContents := navigatePath(tabs[0], "tabRenderer", "content", "sectionListRenderer", "contents")
	sections, ok := tabContents.([]interface{})
	if !ok || len(sections) == 0 {
		slog.Debug("No sections in tab")
		return resp
	}
	slog.Debug("Found sections", "count", len(sections))

	for i, section := range sections {
		slog.Debug("Checking section", "i", i)
		musicContents := navigatePath(section, "musicShelfRenderer", "contents")
		if musicContents == nil {
			musicContents = navigatePath(section, "musicPlaylistShelfRenderer", "contents")
		}
		if musicContents == nil {
			musicContents = navigatePath(section, "gridRenderer", "items")
		}

		items, ok := musicContents.([]interface{})
		if !ok || items == nil {
			slog.Debug("No items found in section", "i", i)
			continue
		}

		slog.Debug("Found items", "count", len(items))

		for _, item := range items {
			track := parsePlaylistItem(item)
			if track != nil {
				resp.Tracks = append(resp.Tracks, *track)
				slog.Debug("Parsed track", "title", track.Title)
			}
		}
	}

	resp.Playlist.TrackCount = len(resp.Tracks)
	slog.Info("Parsed playlist", "title", resp.Playlist.Title, "trackCount", len(resp.Tracks))

	return resp
}

// parsePlaylistItem parses a single playlist item into a Track.
func parsePlaylistItem(item interface{}) *models.Track {
	renderer := navigatePath(item, "musicResponsiveListItemRenderer")
	if renderer == nil {
		return nil
	}
	rendererMap, ok := renderer.(map[string]interface{})
	if !ok {
		return nil
	}

	columns := navigatePath(rendererMap, "flexColumns")
	colArr, ok := columns.([]interface{})
	if !ok || len(colArr) == 0 {
		return nil
	}

	title := extractFlexColumnText(colArr, 0)
	artist := extractFlexColumnText(colArr, 1)
	album := extractFlexColumnText(colArr, 2)
	videoID := extractVideoID(rendererMap)
	thumbnail := extractThumbnail(rendererMap)

	// Try to get duration from fixedColumns
	duration := ""
	fixedCols := navigatePath(rendererMap, "fixedColumns")
	if fixedArr, ok := fixedCols.([]interface{}); ok && len(fixedArr) > 0 {
		duration = extractFlexColumnText(fixedArr, 0)
	}

	return &models.Track{
		VideoID:      videoID,
		Title:        title,
		Artist:       artist,
		Album:        album,
		Duration:     duration,
		ThumbnailURL: thumbnail,
	}
}

// parseLibraryPlaylists parses the browse response for library playlists.
func parseLibraryPlaylists(raw map[string]interface{}) []models.Playlist {
	var playlists []models.Playlist

	sectionContents := navigatePath(raw, "contents", "singleColumnBrowseResultsRenderer", "tabs")
	tabs, ok := sectionContents.([]interface{})
	if !ok || len(tabs) == 0 {
		return playlists
	}

	tabContents := navigatePath(tabs[0], "tabRenderer", "content", "sectionListRenderer", "contents")
	sections, ok := tabContents.([]interface{})
	if !ok {
		return playlists
	}

	for _, section := range sections {
		gridContents := navigatePath(section, "gridRenderer", "items")
		if gridContents == nil {
			gridContents = navigatePath(section, "musicShelfRenderer", "contents")
		}
		items, ok := gridContents.([]interface{})
		if !ok {
			continue
		}

		for _, item := range items {
			pl := parseLibraryPlaylistItem(item)
			if pl != nil {
				playlists = append(playlists, *pl)
			}
		}
	}

	return playlists
}

// parseLibraryPlaylistItem parses a single library playlist item.
func parseLibraryPlaylistItem(item interface{}) *models.Playlist {
	renderer := navigatePath(item, "musicTwoRowItemRenderer")
	if renderer == nil {
		return nil
	}
	rendererMap, ok := renderer.(map[string]interface{})
	if !ok {
		return nil
	}

	// Get browse ID - must have a browseEndpoint to be a real playlist
	browseID := ""
	navEndpoint := navigatePath(rendererMap, "navigationEndpoint", "browseEndpoint", "browseId")
	if id, ok := navEndpoint.(string); ok {
		browseID = id
	}
	// Skip "New playlist" which has createPlaylistEndpoint instead
	if browseID == "" {
		if createEndpoint := navigatePath(rendererMap, "navigationEndpoint", "createPlaylistEndpoint"); createEndpoint != nil {
			return nil // Skip "New playlist" item
		}
		return nil
	}

	// Extract title - may be nested in runs
	title := extractText(navigatePath(rendererMap, "title"))

	// Extract subtitle (author + track count)
	subtitleRuns := navigatePath(rendererMap, "subtitle", "runs")
	var subtitleParts []string
	if runs, ok := subtitleRuns.([]interface{}); ok {
		for _, run := range runs {
			if runMap, ok := run.(map[string]interface{}); ok {
				if text, ok := runMap["text"].(string); ok {
					subtitleParts = append(subtitleParts, text)
				}
			}
		}
	}
	subtitle := ""
	trackCount := 0
	for _, part := range subtitleParts {
		subtitle += part
		// Try to extract track count from text like "16 tracks"
		if len(part) > 7 && part[len(part)-7:] == " tracks" {
			// Extract number
			for i := 0; i < len(part); i++ {
				if part[i] >= '0' && part[i] <= '9' {
					j := i
					for j < len(part) && part[j] >= '0' && part[j] <= '9' {
						j++
					}
					if j > i {
						fmt.Sscanf(part[i:j], "%d", &trackCount)
						break
					}
				}
			}
		}
	}

	thumbnail := extractThumbnailFromObj(rendererMap)

	return &models.Playlist{
		ID:           browseID,
		Title:        title,
		Author:       subtitle,
		TrackCount:   trackCount,
		ThumbnailURL: thumbnail,
	}
}

// parseUserInfo parses the account/list response for user information.
func parseUserInfo(raw map[string]interface{}) *models.UserInfoResponse {
	response := &models.UserInfoResponse{}

	// Try to get account info from the response
	if details, ok := raw["responseContext"].(map[string]interface{}); ok {
		if mainApp, ok := details["mainAppWebResponseContext"].(map[string]interface{}); ok {
			_ = mainApp // Used for debugging if needed
		}
	}

	// Try different path: response.account
	if account, ok := raw["accounts"].([]interface{}); ok && len(account) > 0 {
		if acc, ok := account[0].(map[string]interface{}); ok {
			if id, ok := acc["id"].(string); ok {
				response.ChannelID = id
			}
			if email, ok := acc["email"].(string); ok {
				response.AccountName = email
			}
		}
	}

	// Try: response.header
	if header, ok := raw["header"].(map[string]interface{}); ok {
		if cloudMiner, ok := header["cloudMinerResponseRenderer"].(map[string]interface{}); ok {
			if title, ok := cloudMiner["title"].(map[string]interface{}); ok {
				if text, ok := title["text"].(string); ok {
					response.ChannelTitle = text
				}
			}
		}
	}

	return response
}

// parseWatchPlaylist parses the next endpoint response to get queued tracks.
func parseWatchPlaylist(raw map[string]interface{}) []models.Track {
	var tracks []models.Track

	// Navigate to the playlist panel
	panelContents := navigatePath(raw, "contents", "singleColumnMusicWatchNextResultsRenderer",
		"tabbedRenderer", "watchNextTabbedResultsRenderer", "tabs")
	tabs, ok := panelContents.([]interface{})
	if !ok || len(tabs) == 0 {
		return tracks
	}

	tabContent := navigatePath(tabs[0], "tabRenderer", "content", "musicQueueRenderer",
		"content", "playlistPanelRenderer", "contents")
	items, ok := tabContent.([]interface{})
	if !ok {
		return tracks
	}

	for _, item := range items {
		track := parseWatchTrack(item)
		if track != nil {
			tracks = append(tracks, *track)
		}
	}

	return tracks
}

// parseWatchTrack parses a single watch playlist panel item.
func parseWatchTrack(item interface{}) *models.Track {
	renderer := navigatePath(item, "playlistPanelVideoRenderer")
	if renderer == nil {
		return nil
	}
	rendererMap, ok := renderer.(map[string]interface{})
	if !ok {
		return nil
	}

	title := extractText(navigatePath(rendererMap, "title"))
	videoID := ""
	if vid, ok := rendererMap["videoId"].(string); ok {
		videoID = vid
	}
	thumbnail := extractThumbnailFromObj(rendererMap)

	// Artist from longBylineText
	artist := extractText(navigatePath(rendererMap, "longBylineText"))

	// Duration
	duration := extractText(navigatePath(rendererMap, "lengthText"))

	return &models.Track{
		VideoID:      videoID,
		Title:        title,
		Artist:       artist,
		Duration:     duration,
		ThumbnailURL: thumbnail,
	}
}

// parseSongInfo extracts track info from a player endpoint response.
func parseSongInfo(raw map[string]interface{}, videoID string) *models.Track {
	track := &models.Track{VideoID: videoID}

	details := navigatePath(raw, "videoDetails")
	if detailsMap, ok := details.(map[string]interface{}); ok {
		if t, ok := detailsMap["title"].(string); ok {
			track.Title = t
		}
		if a, ok := detailsMap["author"].(string); ok {
			track.Artist = a
		}
		if lenStr, ok := detailsMap["lengthSeconds"].(string); ok {
			track.Duration = lenStr + "s"
		}

		// Thumbnail
		thumbs := navigatePath(detailsMap, "thumbnail", "thumbnails")
		if thumbArr, ok := thumbs.([]interface{}); ok && len(thumbArr) > 0 {
			if last, ok := thumbArr[len(thumbArr)-1].(map[string]interface{}); ok {
				if u, ok := last["url"].(string); ok {
					track.ThumbnailURL = u
				}
			}
		}
	}

	return track
}

// --- Navigation helpers ---

// navigatePath walks a nested map[string]interface{} by key names.
func navigatePath(obj interface{}, keys ...string) interface{} {
	current := obj
	for _, key := range keys {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[key]
		default:
			return nil
		}
	}
	return current
}

// extractText extracts the text from a runs-based or simpleText-based text object.
func extractText(obj interface{}) string {
	if obj == nil {
		return ""
	}
	objMap, ok := obj.(map[string]interface{})
	if !ok {
		return ""
	}

	if simple, ok := objMap["simpleText"].(string); ok {
		return simple
	}

	if runs, ok := objMap["runs"].([]interface{}); ok {
		var text string
		for _, run := range runs {
			if runMap, ok := run.(map[string]interface{}); ok {
				if t, ok := runMap["text"].(string); ok {
					text += t
				}
			}
		}
		return text
	}

	return ""
}

// extractFlexColumnText extracts the text from a specific flex column index.
func extractFlexColumnText(columns []interface{}, idx int) string {
	if idx >= len(columns) {
		return ""
	}
	textRuns := navigatePath(columns[idx],
		"musicResponsiveListItemFlexColumnRenderer", "text")
	return extractText(textRuns)
}

// extractVideoID tries to extract a videoId from a renderer's overlay or playback endpoint.
func extractVideoID(renderer map[string]interface{}) string {
	// From overlay
	vid := navigatePath(renderer, "overlay", "musicItemThumbnailOverlayRenderer",
		"content", "musicPlayButtonRenderer", "playNavigationEndpoint",
		"watchEndpoint", "videoId")
	if id, ok := vid.(string); ok {
		return id
	}

	// From playlistItemData
	vid = navigatePath(renderer, "playlistItemData", "videoId")
	if id, ok := vid.(string); ok {
		return id
	}

	return ""
}

// extractBrowseID extracts a browseId from a navigation endpoint.
func extractBrowseID(renderer map[string]interface{}) string {
	bid := navigatePath(renderer, "navigationEndpoint", "browseEndpoint", "browseId")
	if id, ok := bid.(string); ok {
		return id
	}
	return ""
}

// extractNavigationType extracts the navigation endpoint type.
func extractNavigationType(renderer map[string]interface{}) string {
	pageType := navigatePath(renderer, "navigationEndpoint", "browseEndpoint",
		"browseEndpointContextSupportedConfigs", "browseEndpointContextMusicConfig",
		"pageType")
	if pt, ok := pageType.(string); ok {
		return pt
	}
	return ""
}

// extractThumbnail extracts the highest-resolution thumbnail URL.
func extractThumbnail(renderer map[string]interface{}) string {
	return extractThumbnailFromObj(renderer)
}

// extractThumbnailFromObj extracts thumbnail URL from a renderer map.
func extractThumbnailFromObj(obj map[string]interface{}) string {
	thumbs := navigatePath(obj, "thumbnail", "musicThumbnailRenderer", "thumbnail", "thumbnails")
	if thumbs == nil {
		thumbs = navigatePath(obj, "thumbnail", "thumbnails")
	}
	thumbArr, ok := thumbs.([]interface{})
	if !ok || len(thumbArr) == 0 {
		return ""
	}
	// Pick last (highest resolution)
	if last, ok := thumbArr[len(thumbArr)-1].(map[string]interface{}); ok {
		if u, ok := last["url"].(string); ok {
			return u
		}
	}
	return ""
}

// containsStr checks if a string contains a substring.
func containsStr(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && bytes.Contains([]byte(s), []byte(substr))
}
