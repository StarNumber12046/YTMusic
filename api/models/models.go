package models

// LoginRequest represents the request body for POST /auth/login.
type LoginRequest struct {
	// Raw Cookie header string copied from browser DevTools
	Cookies string `json:"cookies" binding:"required" example:"SAPISID=abc123/...; __Secure-3PSID=..."`
}

// LoginResponse is returned after a successful login.
type LoginResponse struct {
	Token     string `json:"token" example:"550e8400-e29b-41d4-a716-446655440000"`
	ExpiresAt string `json:"expires_at" example:"2026-03-05T00:00:00Z"`
}

// AuthStatusResponse returns the current session status.
type AuthStatusResponse struct {
	Authenticated bool   `json:"authenticated" example:"true"`
	Token         string `json:"token,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	ExpiresAt     string `json:"expires_at,omitempty" example:"2026-03-05T00:00:00Z"`
	AccountName   string `json:"account_name,omitempty" example:"user@gmail.com"`
}

// Track represents a single music track.
type Track struct {
	VideoID      string     `json:"video_id" example:"dQw4w9WgXcQ"`
	Title        string     `json:"title" example:"Never Gonna Give You Up"`
	Artist       string     `json:"artist" example:"Rick Astley"`
	Album        string     `json:"album,omitempty" example:"Whenever You Need Somebody"`
	DurationMs   int64      `json:"duration_ms,omitempty" example:"213000"`
	Duration     string     `json:"duration,omitempty" example:"3:33"`
	ThumbnailURL string     `json:"thumbnail_url,omitempty" example:"https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg"`
	Artists      []Artist   `json:"artists,omitempty"`
	AlbumInfo    *AlbumInfo `json:"album_info,omitempty"`
}

// Artist represents a music artist.
type Artist struct {
	Name     string `json:"name" example:"Rick Astley"`
	BrowseID string `json:"browse_id,omitempty" example:"UCuAXFkgsw1L7xaCfnd5JJOw"`
}

// AlbumInfo holds album details for a track.
type AlbumInfo struct {
	Name     string `json:"name" example:"Whenever You Need Somebody"`
	BrowseID string `json:"browse_id,omitempty" example:"MPREb_..."`
}

// PlayRequest is the request body to play a specific track.
type PlayRequest struct {
	VideoID string `json:"video_id" binding:"required" example:"dQw4w9WgXcQ"`
}

// PlayerState represents the current playback state.
type PlayerState struct {
	IsPlaying     bool   `json:"is_playing" example:"true"`
	IsPaused      bool   `json:"is_paused" example:"false"`
	CurrentTrack  *Track `json:"current_track,omitempty"`
	QueueLength   int    `json:"queue_length" example:"5"`
	QueuePosition int    `json:"queue_position" example:"0"`
	Volume        int    `json:"volume" example:"100"`
	Shuffle       bool   `json:"shuffle" example:"false"`
	Repeat        string `json:"repeat" example:"off"` // "off", "all", "one"
}

// QueueItem represents a single item in the playback queue.
type QueueItem struct {
	Position int   `json:"position" example:"0"`
	Track    Track `json:"track"`
}

// QueueAddRequest is the request body to add a track to the queue.
type QueueAddRequest struct {
	VideoID string `json:"video_id" binding:"required" example:"dQw4w9WgXcQ"`
}

// QueueResponse returns the full queue.
type QueueResponse struct {
	Items           []QueueItem `json:"items"`
	Length          int         `json:"length" example:"5"`
	CurrentPosition int         `json:"current_position" example:"0"`
}

// Playlist represents a user playlist.
type Playlist struct {
	ID           string `json:"id" example:"PLxxxxxxxxxxxxxxxx"`
	Title        string `json:"title" example:"My Favorites"`
	Description  string `json:"description,omitempty" example:"My top picks"`
	TrackCount   int    `json:"track_count,omitempty" example:"42"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	Author       string `json:"author,omitempty" example:"user@gmail.com"`
}

// PlaylistDetailResponse includes the playlist metadata and its tracks.
type PlaylistDetailResponse struct {
	Playlist Playlist `json:"playlist"`
	Tracks   []Track  `json:"tracks"`
}

// SearchRequest holds query parameters for search.
type SearchRequest struct {
	Query  string `form:"q" binding:"required" example:"never gonna give you up"`
	Filter string `form:"filter" example:"songs"` // songs, playlists, albums, artists
	Limit  int    `form:"limit" example:"20"`
}

// SearchResult holds a single result from a search.
type SearchResult struct {
	ResultType string    `json:"result_type" example:"song"` // song, playlist, album, artist
	Track      *Track    `json:"track,omitempty"`
	Playlist   *Playlist `json:"playlist,omitempty"`
	Artist     *Artist   `json:"artist,omitempty"`
	AlbumRef   *AlbumRef `json:"album,omitempty"`
}

// AlbumRef is a lightweight album reference from search results.
type AlbumRef struct {
	BrowseID     string `json:"browse_id" example:"MPREb_..."`
	Title        string `json:"title" example:"Whenever You Need Somebody"`
	Artist       string `json:"artist,omitempty" example:"Rick Astley"`
	Year         string `json:"year,omitempty" example:"1987"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

// SearchResponse wraps the list of search results.
type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Query   string         `json:"query" example:"never gonna give you up"`
}

// PlaylistListResponse wraps a list of playlists.
type PlaylistListResponse struct {
	Playlists []Playlist `json:"playlists"`
}

// ErrorResponse is the standard error payload.
type ErrorResponse struct {
	Error string `json:"error" example:"unauthorized"`
	Code  int    `json:"code" example:"401"`
}

// MessageResponse is a simple success message.
type MessageResponse struct {
	Message string `json:"message" example:"ok"`
}

// VolumeRequest is the request body to set volume.
type VolumeRequest struct {
	Volume int `json:"volume" binding:"required" example:"80"`
}

// RepeatRequest is the request body to set repeat mode.
type RepeatRequest struct {
	Repeat string `json:"repeat" binding:"required" example:"all"` // "off", "all", "one"
}

// UserInfoResponse holds account information.
type UserInfoResponse struct {
	AccountName  string `json:"account_name,omitempty"`
	ChannelID    string `json:"channel_id,omitempty"`
	ChannelTitle string `json:"channel_title,omitempty"`
}
