package api

type Track struct {
	VideoID      string `json:"video_id"`
	Title        string `json:"title"`
	Artist       string `json:"artist"`
	Album        string `json:"album,omitempty"`
	Duration     string `json:"duration,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

type Artist struct {
	Name     string `json:"name"`
	BrowseID string `json:"browse_id,omitempty"`
}

type PlayerState struct {
	IsPlaying         bool   `json:"is_playing"`
	IsPaused          bool   `json:"is_paused"`
	CurrentTrack      *Track `json:"current_track,omitempty"`
	QueueLength       int    `json:"queue_length"`
	QueuePosition     int    `json:"queue_position"`
	Volume            int    `json:"volume"`
	Shuffle           bool   `json:"shuffle"`
	Repeat            string `json:"repeat"`
	CurrentPositionMs int64  `json:"current_position_ms"`
}

type LyricsLine struct {
	TimeMs int64  `json:"time_ms"`
	Text   string `json:"text"`
}

type LyricsResponse struct {
	TrackName    string       `json:"track_name"`
	ArtistName   string       `json:"artist_name"`
	AlbumName    string       `json:"album_name,omitempty"`
	Duration     float64      `json:"duration,omitempty"`
	PlainLyrics  string       `json:"plain_lyrics,omitempty"`
	SyncedLyrics string       `json:"synced_rics,omitempty"`
	ParsedLyrics []LyricsLine `json:"parsed_lyrics,omitempty"`
	Source       string       `json:"source"`
}

type SearchResult struct {
	ResultType string    `json:"result_type"`
	Track      *Track    `json:"track,omitempty"`
	Playlist   *Playlist `json:"playlist,omitempty"`
}

type Playlist struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	TrackCount   int    `json:"track_count,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

type PlaylistDetail struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	TrackCount   int     `json:"track_count"`
	Tracks       []Track `json:"tracks"`
	ThumbnailURL string  `json:"thumbnail_url,omitempty"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

type PlaylistListResponse struct {
	Playlists []Playlist `json:"playlists"`
}

type QueueResponse struct {
	Items []QueueItem `json:"items"`
}

type QueueItem struct {
	Position int   `json:"position"`
	Track    Track `json:"track"`
}

type AuthStatusResponse struct {
	Authenticated bool   `json:"authenticated"`
	AccountName   string `json:"account_name,omitempty"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
	Message string `json:"message,omitempty"`
}

type UserInfo struct {
	AccountName string `json:"account_name"`
	Email       string `json:"email,omitempty"`
}
