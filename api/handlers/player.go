package handlers

import (
	"net/http"

	"ytmusic_api/middleware"
	"ytmusic_api/models"
	"ytmusic_api/player"
	"ytmusic_api/ytmusic"

	"github.com/gin-gonic/gin"
)

// PlayerHandler holds dependencies for playback control endpoints.
type PlayerHandler struct {
	Player *player.Player
	Client *ytmusic.Client
}

// NewPlayerHandler creates a new PlayerHandler.
func NewPlayerHandler(p *player.Player, client *ytmusic.Client) *PlayerHandler {
	return &PlayerHandler{
		Player: p,
		Client: client,
	}
}

// Play godoc
// @Summary      Play a track
// @Description  Resolves the audio stream for a YouTube Music video ID and begins playback.
// @Description  If the track is not already in the queue, it is added at the current position.
// @Tags         player
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        body body models.PlayRequest true "Track to play"
// @Success      200 {object} models.PlayerState
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      500 {object} models.ErrorResponse
// @Router       /player/play [post]
func (h *PlayerHandler) Play(c *gin.Context) {
	session := middleware.GetSession(c)

	var req models.PlayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid request body: " + err.Error(),
			Code:  http.StatusBadRequest,
		})
		return
	}

	// Try to get song info from Innertube for richer metadata
	track := &models.Track{VideoID: req.VideoID}
	if session != nil {
		if info, err := h.Client.GetSongInfo(session, req.VideoID); err == nil && info != nil {
			track = info
		}
	}

	// Add to queue if not already the current track
	current := h.Player.Queue.Current()
	if current == nil || current.VideoID != track.VideoID {
		h.Player.Queue.Add(*track)
		// Set position to the newly added track
		h.Player.Queue.SetPosition(h.Player.Queue.Len() - 1)
	}

	if err := h.Player.PlayTrack(track); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "playback failed: " + err.Error(),
			Code:  http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, h.Player.State())
}

// PauseToggle godoc
// @Summary      Toggle pause/resume
// @Description  Pauses the current track if playing, or resumes if paused.
// @Tags         player
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} models.PlayerState
// @Failure      401 {object} models.ErrorResponse
// @Router       /player/pause [post]
func (h *PlayerHandler) PauseToggle(c *gin.Context) {
	h.Player.Pause()
	c.JSON(http.StatusOK, h.Player.State())
}

// NextTrack godoc
// @Summary      Skip to next track
// @Description  Advances to the next track in the queue. Returns the new player state.
// @Tags         player
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} models.PlayerState
// @Failure      401 {object} models.ErrorResponse
// @Failure      500 {object} models.ErrorResponse
// @Router       /player/next [post]
func (h *PlayerHandler) NextTrack(c *gin.Context) {
	_, err := h.Player.Next()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to skip: " + err.Error(),
			Code:  http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, h.Player.State())
}

// PreviousTrack godoc
// @Summary      Go to previous track
// @Description  Goes back to the previous track in the queue. Returns the new player state.
// @Tags         player
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} models.PlayerState
// @Failure      401 {object} models.ErrorResponse
// @Failure      500 {object} models.ErrorResponse
// @Router       /player/previous [post]
func (h *PlayerHandler) PreviousTrack(c *gin.Context) {
	_, err := h.Player.Previous()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to go back: " + err.Error(),
			Code:  http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, h.Player.State())
}

// GetState godoc
// @Summary      Get player state
// @Description  Returns the current playback state including current track, pause status, and queue info.
// @Tags         player
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} models.PlayerState
// @Failure      401 {object} models.ErrorResponse
// @Router       /player/state [get]
func (h *PlayerHandler) GetState(c *gin.Context) {
	c.JSON(http.StatusOK, h.Player.State())
}

// Stop godoc
// @Summary      Stop playback
// @Description  Stops the current track and clears the playing state. Queue is preserved.
// @Tags         player
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} models.PlayerState
// @Failure      401 {object} models.ErrorResponse
// @Router       /player/stop [post]
func (h *PlayerHandler) Stop(c *gin.Context) {
	h.Player.Stop()
	c.JSON(http.StatusOK, h.Player.State())
}

// SetVolume godoc
// @Summary      Set playback volume
// @Description  Sets the playback volume (0-100).
// @Tags         player
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        body body models.VolumeRequest true "Volume level"
// @Success      200 {object} models.PlayerState
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Router       /player/volume [post]
func (h *PlayerHandler) SetVolume(c *gin.Context) {
	var req models.VolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid request body: " + err.Error(),
			Code:  http.StatusBadRequest,
		})
		return
	}

	h.Player.SetVolume(req.Volume)
	c.JSON(http.StatusOK, h.Player.State())
}

// ToggleShuffle godoc
// @Summary      Toggle shuffle mode
// @Description  Toggles shuffle mode on or off. When enabled, queue is shuffled.
// @Tags         player
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} models.PlayerState
// @Failure      401 {object} models.ErrorResponse
// @Router       /player/shuffle [post]
func (h *PlayerHandler) ToggleShuffle(c *gin.Context) {
	h.Player.ToggleShuffle()
	c.JSON(http.StatusOK, h.Player.State())
}

// SetRepeat godoc
// @Summary      Set repeat mode
// @Description  Sets repeat mode: "off", "all", "one", or "cycle" (cycles to next mode).
// @Tags         player
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        body body models.RepeatRequest true "Repeat mode"
// @Success      200 {object} models.PlayerState
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Router       /player/repeat [post]
func (h *PlayerHandler) SetRepeat(c *gin.Context) {
	var req models.RepeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid request body: " + err.Error(),
			Code:  http.StatusBadRequest,
		})
		return
	}

	// Handle cycle mode
	if req.Repeat == "cycle" {
		h.Player.CycleRepeat()
	} else {
		h.Player.SetRepeat(req.Repeat)
	}
	c.JSON(http.StatusOK, h.Player.State())
}
