package player

import (
	"fmt"
	"sync"
	"time"

	"ytmusic_api/models"

	"github.com/777vibecoder/discord-rpc"
)

type DiscordPresence struct {
	client *discordrpc.Client
	mu     sync.Mutex
}

func NewDiscordPresence(clientID string) (*DiscordPresence, error) {
	if clientID == "" {
		return nil, nil
	}

	client, err := discordrpc.New(clientID)
	if err != nil {
		return nil, fmt.Errorf("creating Discord RPC client: %w", err)
	}

	return &DiscordPresence{client: client}, nil
}

func (d *DiscordPresence) UpdatePresence(track *models.Track, isPlaying bool, positionMs int64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.client == nil || track == nil {
		return
	}

	if !isPlaying {
		d.clearLocked()
		return
	}

	activity := d.buildActivity(track, positionMs)
	_ = d.client.SetActivity(activity)
}

func (d *DiscordPresence) buildActivity(track *models.Track, positionMs int64) discordrpc.Activity {
	activity := discordrpc.Activity{
		Type:    2, // Listening to
		Details: track.Title,
		State:   fmt.Sprintf("by %s", track.Artist),
		Assets: &discordrpc.Assets{
			LargeImage: track.ThumbnailURL,
			LargeText:  track.Title,
		},
	}

	if track.DurationMs > 0 {
		startTime := discordrpc.Epoch{Time: time.Now().Add(-time.Duration(positionMs) * time.Millisecond)}
		endTime := discordrpc.Epoch{Time: time.Now().Add(time.Duration(track.DurationMs-positionMs) * time.Millisecond)}
		activity.Timestamps = &discordrpc.Timestamps{
			Start: &startTime,
			End:   &endTime,
		}
	}

	return activity
}

func (d *DiscordPresence) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.clearLocked()
}

func (d *DiscordPresence) clearLocked() {
	if d.client == nil {
		return
	}
	_ = d.client.SetActivity(discordrpc.Activity{})
}

func (d *DiscordPresence) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.client == nil {
		return
	}
	_ = d.client.Socket.Close()
	d.client = nil
}
