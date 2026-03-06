package api

import (
	"fmt"

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

func (c *Client) CycleRepeat() error {
	state, err := c.GetPlayerState()
	if err != nil {
		return err
	}

	nextRepeat := "off"
	switch state.Repeat {
	case "off":
		nextRepeat = "all"
	case "all":
		nextRepeat = "one"
	case "one":
		nextRepeat = "off"
	}

	resp, err := c.restClient.R().
		SetBody(map[string]string{"repeat": nextRepeat}).
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
