package player

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"ytmusic_api/models"

	"github.com/ebitengine/oto/v3"
	yt "github.com/kkdai/youtube/v2"
)

const (
	sampleRate   = 44100
	channelCount = 2
	bitDepth     = 2 // 16-bit = 2 bytes per sample
)

// Player is the core audio playback engine.
// It owns the oto context, the current streamer, and the playback queue.
type Player struct {
	mu sync.Mutex

	otoCtx    *oto.Context
	otoPlayer *oto.Player
	streamer  *Streamer
	ytClient  *yt.Client

	Queue *Queue

	discord         *DiscordPresence
	discordClientID string

	currentTrack   *models.Track
	isPlaying      bool
	isPaused       bool
	volume         int // 0-100
	shuffle        bool
	repeat         string // "off", "all", "one"
	playStartTime  time.Time
	pausedDuration time.Duration
}

// NewPlayer creates a new Player instance and initialises the audio context.
func NewPlayer(discordClientID string) (*Player, error) {
	op := &oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: channelCount,
		Format:       oto.FormatSignedInt16LE,
	}

	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		return nil, fmt.Errorf("initialising audio context: %w", err)
	}
	<-readyChan

	discordPresence, err := NewDiscordPresence(discordClientID)
	if err != nil {
		return nil, fmt.Errorf("initialising Discord presence: %w", err)
	}

	// Attempt to load saved volume from disk and apply it on startup
	vol := 100
	if saved, err := loadVolumeFromDisk(); err == nil {
		if saved >= 0 && saved <= 100 {
			vol = saved
		}
	}

	p := &Player{
		otoCtx:          otoCtx,
		ytClient:        &yt.Client{},
		Queue:           NewQueue(),
		volume:          vol,
		discord:         discordPresence,
		discordClientID: discordClientID,
	}
	// Start media-key/headphone event listener for playback controls
	go StartMediaEventDispatcher(p)
	StartMediaHotkeyListener()
	return p, nil
}

// PlayTrack resolves the audio stream for a videoId, stops any current playback,
// and begins playing the new track.
func (p *Player) PlayTrack(track *models.Track) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Stop any current playback
	p.stopLocked()

	// Always sync cache the track before playing
	if DefaultCacheManager != nil {
		if err := DefaultCacheManager.CacheSync(track.VideoID); err != nil {
			slog.Error("Failed to cache track synchronously", "videoID", track.VideoID, "error", err)
			return fmt.Errorf("failed to cache track for playing: %w", err)
		}
	}

	streamURL := DefaultCacheManager.GetCachedPath(track.VideoID)
	slog.Info("Playing track from local cache", "videoID", track.VideoID)

	// Enrich track info if title is missing
	if track.Title == "" {
		if info, infoErr := p.getVideoInfo(track.VideoID); infoErr == nil && info != nil {
			track.Title = info.Title
			track.Artist = info.Author
			track.ThumbnailURL = info.Thumbnails[0].URL
		}
	}

	// Start ffmpeg streamer
	streamer, err := NewStreamer(streamURL, true)
	if err != nil {
		return fmt.Errorf("creating audio streamer: %w", err)
	}

	// Create oto player from the PCM stream
	otoPlayer := p.otoCtx.NewPlayer(streamer)
	otoPlayer.SetVolume(float64(p.volume) / 100.0)
	otoPlayer.Play()

	p.streamer = streamer
	p.otoPlayer = otoPlayer
	p.currentTrack = track
	p.isPlaying = true
	p.isPaused = false
	p.playStartTime = time.Now()
	p.pausedDuration = 0

	// Reinitialize Discord presence if needed (e.g., after pause was closed)
	if p.discordClientID != "" && p.discord == nil {
		discordPresence, err := NewDiscordPresence(p.discordClientID)
		if err != nil {
			slog.Error("Failed to reinitialize Discord presence", "error", err)
		} else {
			p.discord = discordPresence
		}
	}

	// Update Discord presence
	if p.discord != nil {
		p.discord.UpdatePresence(track, true, 0)
	}

	// Monitor for track end in background
	go p.monitorPlayback()

	return nil
}

// Pause toggles pause/resume on the current track.
func (p *Player) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()

	slog.Info("Pause/Toggle requested", "isPaused", p.isPaused, "isPlaying", p.isPlaying, "hasOtoPlayer", p.otoPlayer != nil)

	if p.otoPlayer == nil {
		slog.Warn("Pause/Toggle ignored: otoPlayer is nil")
		return
	}

	if p.isPaused {
		slog.Info("Resuming playback")
		p.otoPlayer.Play()
		p.isPaused = false
		p.isPlaying = true
		// Adjust play start time to account for paused duration
		p.playStartTime = time.Now().Add(-p.pausedDuration)

		// Reinitialize Discord presence on resume
		if p.discordClientID != "" {
			if p.discord != nil {
				p.discord.Close()
			}
			discordPresence, err := NewDiscordPresence(p.discordClientID)
			if err != nil {
				slog.Error("Failed to reinitialize Discord presence on resume", "error", err)
			} else {
				p.discord = discordPresence
			}
		}
	} else {
		slog.Info("Pausing playback")
		p.otoPlayer.Pause()
		p.isPaused = true
		p.isPlaying = false

		// Close Discord presence completely on pause
		if p.discord != nil {
			p.discord.Close()
			p.discord = nil
		}
	}

	// Update Discord presence (or it will be reinitialized on resume)
	if p.discord != nil {
		positionMs := p.currentPositionMs()
		p.discord.UpdatePresence(p.currentTrack, !p.isPaused, positionMs)
	}
}

// Stop halts all playback and releases the streamer.
func (p *Player) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stopLocked()

	// Clear Discord presence
	if p.discord != nil {
		p.discord.Clear()
	}
}

// stopLocked stops playback (caller must hold p.mu).
func (p *Player) stopLocked() {
	if p.otoPlayer != nil {
		p.otoPlayer.Pause()
		p.otoPlayer = nil
	}
	if p.streamer != nil {
		_ = p.streamer.Close()
		p.streamer = nil
	}
	p.isPlaying = false
	p.isPaused = false
}

// Next skips to the next track in the queue.
// Returns the new track or nil if queue is exhausted.
func (p *Player) Next() (*models.Track, error) {
	p.mu.Lock()
	repeat := p.repeat
	currentTrack := p.currentTrack
	p.mu.Unlock()

	// Handle repeat "one" - replay current track
	if repeat == "one" && currentTrack != nil {
		if err := p.PlayTrack(currentTrack); err != nil {
			return nil, err
		}
		return currentTrack, nil
	}

	next := p.Queue.Next()
	if next == nil {
		// Handle repeat "all" - wrap to beginning
		if repeat == "all" && p.Queue.Len() > 0 {
			p.Queue.SetPosition(0)
			next = p.Queue.Next()
		}
		if next == nil {
			p.Stop()
			return nil, nil
		}
	}
	if err := p.PlayTrack(next); err != nil {
		return nil, err
	}
	return next, nil
}

// Previous goes back to the previous track in the queue.
func (p *Player) Previous() (*models.Track, error) {
	prev := p.Queue.Previous()
	if prev == nil {
		return nil, nil
	}
	if err := p.PlayTrack(prev); err != nil {
		return nil, err
	}
	return prev, nil
}

// State returns the current player state.
func (p *Player) State() models.PlayerState {
	p.mu.Lock()
	defer p.mu.Unlock()

	return models.PlayerState{
		IsPlaying:         p.isPlaying,
		IsPaused:          p.isPaused,
		CurrentTrack:      p.currentTrack,
		QueueLength:       p.Queue.Len(),
		QueuePosition:     p.Queue.Position(),
		Volume:            p.volume,
		Shuffle:           p.shuffle,
		Repeat:            p.repeat,
		CurrentPositionMs: p.currentPositionMs(),
	}
}

// SetVolume sets the playback volume (0-100).
func (p *Player) SetVolume(vol int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if vol < 0 {
		vol = 0
	}
	if vol > 100 {
		vol = 100
	}
	p.volume = vol

	if p.otoPlayer != nil {
		// oto volume is a float64 where 1.0 = 100%
		p.otoPlayer.SetVolume(float64(vol) / 100.0)
	}

	// Persist volume to disk for restoration after restart
	if err := saveVolumeToDisk(vol); err != nil {
		slog.Error("failed to persist volume", "volume", vol, "error", err)
	}
}

// saveVolumeToDisk writes the current volume to disk so it can be restored on restart.
func saveVolumeToDisk(vol int) error {
	// Ensure directory exists: ~/.ytmusic
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".ytmusic")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, "volume.json")
	payload, err := json.Marshal(struct {
		Volume int `json:"volume"`
	}{Volume: vol})
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0644)
}

// loadVolumeFromDisk reads the saved volume from disk, if present.
func loadVolumeFromDisk() (int, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, err
	}
	path := filepath.Join(home, ".ytmusic", "volume.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	var payload struct {
		Volume int `json:"volume"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return 0, err
	}
	return payload.Volume, nil
}

// ToggleShuffle toggles shuffle mode on/off.
func (p *Player) ToggleShuffle() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.shuffle = !p.shuffle
	if p.shuffle {
		p.Queue.Shuffle()
	} else {
		p.Queue.Unshuffle()
	}
}

// SetRepeat sets the repeat mode: "off", "all", or "one".
func (p *Player) SetRepeat(mode string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if mode != "off" && mode != "all" && mode != "one" {
		return
	}
	p.repeat = mode
}

// CycleRepeat cycles through repeat modes: off -> all -> one -> off.
func (p *Player) CycleRepeat() {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch p.repeat {
	case "off":
		p.repeat = "all"
	case "all":
		p.repeat = "one"
	case "one":
		p.repeat = "off"
	default:
		p.repeat = "off"
	}
}

// resolveAudioURL uses kkdai/youtube to get the best audio-only stream URL for a video.
func (p *Player) resolveAudioURL(videoID string) (string, error) {
	video, err := p.ytClient.GetVideo(videoID)
	if err != nil {
		return "", fmt.Errorf("fetching video info: %w", err)
	}

	// Prefer audio-only formats, sorted by bitrate
	formats := video.Formats.Type("audio")
	if len(formats) == 0 {
		return "", fmt.Errorf("no audio streams found for video %s", videoID)
	}

	// Sort by audio quality (bitrate) descending
	formats.Sort()

	streamURL, err := p.ytClient.GetStreamURL(video, &formats[0])
	if err != nil {
		return "", fmt.Errorf("getting stream URL: %w", err)
	}

	return streamURL, nil
}

// getVideoInfo retrieves video metadata via kkdai/youtube.
func (p *Player) getVideoInfo(videoID string) (*yt.Video, error) {
	return p.ytClient.GetVideo(videoID)
}

// currentPositionMs returns the current playback position in milliseconds.
func (p *Player) currentPositionMs() int64 {
	if p.playStartTime.IsZero() || p.currentTrack == nil {
		return 0
	}
	elapsed := time.Since(p.playStartTime) - p.pausedDuration
	return elapsed.Milliseconds()
}

// monitorPlayback watches for the current track to finish playing, then auto-advances.
func (p *Player) monitorPlayback() {
	p.mu.Lock()
	currentPlayer := p.otoPlayer
	currentStreamer := p.streamer
	currentTrack := p.currentTrack
	discord := p.discord
	p.mu.Unlock()

	if currentPlayer == nil || currentStreamer == nil {
		return
	}

	var lastPauseTime time.Time

	// Poll until oto player reports it is no longer playing.
	// oto has no callback mechanism, so we poll with a sleep to avoid CPU spin.
	for {
		p.mu.Lock()
		// Check if player was replaced/stopped manually
		if p.otoPlayer != currentPlayer {
			p.mu.Unlock()
			return
		}
		isPaused := p.isPaused
		playStartTime := p.playStartTime
		pausedDuration := p.pausedDuration
		p.mu.Unlock()

		if !currentPlayer.IsPlaying() {
			// If playback stopped because of a manual pause, keep waiting.
			if isPaused {
				// Track paused duration
				if lastPauseTime.IsZero() {
					lastPauseTime = time.Now()
				} else {
					p.mu.Lock()
					p.pausedDuration = pausedDuration + time.Since(lastPauseTime)
					p.mu.Unlock()
				}
				// Update Discord presence while paused
				if discord != nil && currentTrack != nil {
					positionMs := p.currentPositionMs()
					discord.UpdatePresence(currentTrack, false, positionMs)
				}
				time.Sleep(250 * time.Millisecond)
				continue
			}

			lastPauseTime = time.Time{}

			// If oto reports it's not playing, check if we've actually reached the end of the stream.
			// Sometimes otoPlayer.IsPlaying() can be briefly false if the buffer is starved or hasn't started yet.
			if currentStreamer.IsEOF() {
				break // Stream is finished and oto has drained its buffer
			}
		} else {
			// Track when we resume from pause
			if lastPauseTime.IsZero() == false && !isPaused {
				lastPauseTime = time.Time{}
			}
			// Update Discord presence periodically while playing
			if discord != nil && currentTrack != nil && !isPaused && !playStartTime.IsZero() {
				positionMs := p.currentPositionMs()
				discord.UpdatePresence(currentTrack, true, positionMs)
			}
		}
		time.Sleep(250 * time.Millisecond)
	}

	slog.Info("Track playback finished naturally", "track", currentTrack.Title)

	// Clear Discord presence when track ends
	if discord != nil {
		discord.Clear()
	}

	// Check if this is still the active player (not replaced by skip/new play)
	p.mu.Lock()
	if p.otoPlayer == currentPlayer {
		slog.Info("Auto-advancing to next track")
		p.stopLocked()
		p.mu.Unlock()

		// Auto-advance to next track
		next := p.Queue.Next()
		if next != nil {
			if err := p.PlayTrack(next); err != nil {
				slog.Error("auto-advance failed", "error", err)
			}
		} else {
			slog.Info("Queue finished")
		}
	} else {
		p.mu.Unlock()
	}
}

// Close shuts down the player and releases all resources.
func (p *Player) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.stopLocked()

	if p.discord != nil {
		p.discord.Close()
	}
}
