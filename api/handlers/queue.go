package handlers

import (
	"net/http"
	"strconv"

	"ytmusic_api/middleware"
	"ytmusic_api/models"
	"ytmusic_api/player"
	"ytmusic_api/ytmusic"

	"github.com/gin-gonic/gin"
)

// QueueHandler holds dependencies for queue management endpoints.
type QueueHandler struct {
	Player *player.Player
	Client *ytmusic.Client
}

// NewQueueHandler creates a new QueueHandler.
func NewQueueHandler(p *player.Player, client *ytmusic.Client) *QueueHandler {
	return &QueueHandler{
		Player: p,
		Client: client,
	}
}

// GetQueue godoc
// @Summary      List queue
// @Description  Returns the full playback queue with track details and current position.
// @Tags         queue
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} models.QueueResponse
// @Failure      401 {object} models.ErrorResponse
// @Router       /queue [get]
func (h *QueueHandler) GetQueue(c *gin.Context) {
	items := h.Player.Queue.Items()
	queueItems := make([]models.QueueItem, len(items))
	for i, track := range items {
		queueItems[i] = models.QueueItem{
			Position: i,
			Track:    track,
		}
	}

	c.JSON(http.StatusOK, models.QueueResponse{
		Items:           queueItems,
		Length:          len(items),
		CurrentPosition: h.Player.Queue.Position(),
	})
}

// AddToQueue godoc
// @Summary      Add track to queue
// @Description  Adds a track to the end of the playback queue by video ID.
// @Tags         queue
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        body body models.QueueAddRequest true "Track to add"
// @Success      200 {object} models.QueueResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      500 {object} models.ErrorResponse
// @Router       /queue/add [post]
func (h *QueueHandler) AddToQueue(c *gin.Context) {
	session := middleware.GetSession(c)

	var req models.QueueAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid request body: " + err.Error(),
			Code:  http.StatusBadRequest,
		})
		return
	}

	// Try to fetch track info
	track := models.Track{VideoID: req.VideoID}
	if session != nil {
		if info, err := h.Client.GetSongInfo(session, req.VideoID); err == nil && info != nil {
			track = *info
		}
	}

	h.Player.Queue.Add(track)

	// Return updated queue
	h.GetQueue(c)
}

// ClearQueue godoc
// @Summary      Clear queue
// @Description  Removes all tracks from the queue. Stops playback if active.
// @Tags         queue
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} models.MessageResponse
// @Failure      401 {object} models.ErrorResponse
// @Router       /queue [delete]
func (h *QueueHandler) ClearQueue(c *gin.Context) {
	h.Player.Stop()
	h.Player.Queue.Clear()

	c.JSON(http.StatusOK, models.MessageResponse{Message: "queue cleared"})
}

// RemoveFromQueue godoc
// @Summary      Remove track from queue
// @Description  Removes the track at the specified position (0-indexed) from the queue.
// @Tags         queue
// @Produce      json
// @Security     ApiKeyAuth
// @Param        position path int true "Queue position (0-indexed)"
// @Success      200 {object} models.QueueResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Router       /queue/{position} [delete]
func (h *QueueHandler) RemoveFromQueue(c *gin.Context) {
	posStr := c.Param("position")
	pos, err := strconv.Atoi(posStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid position: must be an integer",
			Code:  http.StatusBadRequest,
		})
		return
	}

	if !h.Player.Queue.Remove(pos) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "position out of range",
			Code:  http.StatusBadRequest,
		})
		return
	}

	// Return updated queue
	h.GetQueue(c)
}

// PlayNext godoc
// @Summary      Play track next
// @Description  Inserts a track to play immediately after the current track.
// @Tags         queue
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        body body models.QueueAddRequest true "Track to play next"
// @Success      200 {object} models.QueueResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Router       /queue/play-next [post]
func (h *QueueHandler) PlayNext(c *gin.Context) {
	session := middleware.GetSession(c)

	var req models.QueueAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid request body: " + err.Error(),
			Code:  http.StatusBadRequest,
		})
		return
	}

	// Try to fetch track info
	track := models.Track{VideoID: req.VideoID}
	if session != nil {
		if info, err := h.Client.GetSongInfo(session, req.VideoID); err == nil && info != nil {
			track = *info
		}
	}

	h.Player.PlayNext(&track)

	// Return updated queue
	h.GetQueue(c)
}
