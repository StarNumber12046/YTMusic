package api

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	restClient *resty.Client
	BaseURL    string
	Token      string
}

func NewClient(baseURL string, token string) *Client {
	return &Client{
		restClient: resty.New().SetBaseURL(baseURL).SetHeader("X-Session-Token", token),
		BaseURL:    baseURL,
		Token:      token,
	}
}

func (c *Client) GetPlayerState() (*PlayerState, error) {
	var state PlayerState
	resp, err := c.restClient.R().
		SetResult(&state).
		Get("/player/state")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error: %s", resp.String())
	}
	return &state, nil
}

func (c *Client) PlayPause() error {
	resp, err := c.restClient.R().Post("/player/pause")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("api error: %s", resp.String())
	}
	return nil
}

func (c *Client) Next() error {
	resp, err := c.restClient.R().Post("/player/next")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("api error: %s", resp.String())
	}
	return nil
}

func (c *Client) Previous() error {
	resp, err := c.restClient.R().Post("/player/previous")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("api error: %s", resp.String())
	}
	return nil
}

func (c *Client) Stop() error {
	resp, err := c.restClient.R().Post("/player/stop")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("api error: %s", resp.String())
	}
	return nil
}

func (c *Client) Search(query string) ([]SearchResult, error) {
	var result SearchResponse
	resp, err := c.restClient.R().
		SetQueryParam("q", query).
		SetResult(&result).
		Get("/search")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error: %s", resp.String())
	}
	return result.Results, nil
}

func (c *Client) PlayTrack(videoID string) error {
	resp, err := c.restClient.R().
		SetBody(map[string]string{"video_id": videoID}).
		Post("/player/play")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("api error: %s", resp.String())
	}
	return nil
}

func (c *Client) GetPlaylists() ([]Playlist, error) {
	var result PlaylistListResponse
	resp, err := c.restClient.R().
		SetResult(&result).
		Get("/playlists")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error: %s", resp.String())
	}
	return result.Playlists, nil
}

func (c *Client) GetPlaylist(id string) (*PlaylistDetail, error) {
	var result PlaylistDetail
	resp, err := c.restClient.R().
		SetResult(&result).
		Get(fmt.Sprintf("/playlists/%s", id))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error: %s", resp.String())
	}
	return &result, nil
}

func (c *Client) PlayPlaylist(id string) error {
	resp, err := c.restClient.R().
		Post(fmt.Sprintf("/playlists/%s/play", id))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("api error: %s", resp.String())
	}
	return nil
}

func (c *Client) GetQueue() ([]QueueItem, error) {
	var result QueueResponse
	resp, err := c.restClient.R().
		SetResult(&result).
		Get("/queue")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error: %s", resp.String())
	}
	return result.Items, nil
}

func (c *Client) AddToQueue(videoID string) error {
	resp, err := c.restClient.R().
		SetBody(map[string]string{"video_id": videoID}).
		Post("/queue/add")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("api error: %s", resp.String())
	}
	return nil
}

func (c *Client) PlayNext(videoID string) error {
	resp, err := c.restClient.R().
		SetBody(map[string]string{"video_id": videoID}).
		Post("/queue/play-next")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("api error: %s", resp.String())
	}
	return nil
}

func (c *Client) ClearQueue() error {
	resp, err := c.restClient.R().Delete("/queue")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("api error: %s", resp.String())
	}
	return nil
}

func (c *Client) GetAuthStatus() (*AuthStatusResponse, error) {
	var result AuthStatusResponse
	resp, err := c.restClient.R().
		SetResult(&result).
		Get("/auth/status")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error: %s", resp.String())
	}
	return &result, nil
}

func (c *Client) Login(cookies string) (*LoginResponse, error) {
	var result LoginResponse
	resp, err := c.restClient.R().
		SetBody(map[string]string{"cookies": cookies}).
		SetResult(&result).
		Post("/auth/login")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error: %s", resp.String())
	}
	c.Token = result.Token
	c.restClient.SetHeader("X-Session-Token", result.Token)
	return &result, nil
}

func (c *Client) SetVolume(volume int) error {
	if volume < 0 {
		volume = 0
	}
	if volume > 100 {
		volume = 100
	}
	resp, err := c.restClient.R().
		SetBody(map[string]int{"volume": volume}).
		Post("/player/volume")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("api error: %s", resp.String())
	}
	return nil
}

func (c *Client) ToggleShuffle() error {
	resp, err := c.restClient.R().Post("/player/shuffle")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("api error: %s", resp.String())
	}
	return nil
}

func (c *Client) SetRepeat(repeat string) error {
	resp, err := c.restClient.R().
		SetBody(map[string]string{"repeat": repeat}).
		Post("/player/repeat")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("api error: %s", resp.String())
	}
	return nil
}

func (c *Client) GetLyrics() (*LyricsResponse, error) {
	var result LyricsResponse
	resp, err := c.restClient.R().
		SetResult(&result).
		Get("/lyrics")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error: %s", resp.String())
	}
	return &result, nil
}

func (c *Client) IsOnline() bool {
	resp, err := http.Get(c.BaseURL + "/auth/status")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized
}

func (c *Client) Logout() error {
	resp, err := c.restClient.R().Delete("/auth/logout")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("api error: %s", resp.String())
	}
	c.Token = ""
	c.restClient.SetHeader("X-Session-Token", "")
	return nil
}

func (c *Client) GetUserInfo() (*UserInfo, error) {
	var result UserInfo
	resp, err := c.restClient.R().
		SetResult(&result).
		Get("/user")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error: %s", resp.String())
	}
	return &result, nil
}
