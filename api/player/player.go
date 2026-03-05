package player

import (
	"fmt"
	"log/slog"
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

	currentTrack *models.Track
	isPlaying    bool
	isPaused     bool
	volume       int // 0-100
	shuffle      bool
	repeat       string // "off", "all", "one"
}

// NewPlayer creates a new Player instance and initialises the audio context.
func NewPlayer() (*Player, error) {
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

	return &Player{
		otoCtx:   otoCtx,
		ytClient: &yt.Client{},
		Queue:    NewQueue(),
		volume:   100,
	}, nil
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
	otoPlayer.Play()

	p.streamer = streamer
	p.otoPlayer = otoPlayer
	p.currentTrack = track
	p.isPlaying = true
	p.isPaused = false

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
	} else {
		slog.Info("Pausing playback")
		p.otoPlayer.Pause()
		p.isPaused = true
		// we leave p.isPlaying as true or false? If paused, isPlaying is typically false in the API response.
		// The frontend expects: isPlaying = true, isPaused = false for playing.
		// For paused: isPlaying = true, isPaused = true or isPlaying = false, isPaused = true.
		p.isPlaying = false
	}
}

// Stop halts all playback and releases the streamer.
func (p *Player) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stopLocked()
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
		IsPlaying:     p.isPlaying,
		IsPaused:      p.isPaused,
		CurrentTrack:  p.currentTrack,
		QueueLength:   p.Queue.Len(),
		QueuePosition: p.Queue.Position(),
		Volume:        p.volume,
		Shuffle:       p.shuffle,
		Repeat:        p.repeat,
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

// monitorPlayback watches for the current track to finish playing, then auto-advances.
func (p *Player) monitorPlayback() {
	p.mu.Lock()
	currentPlayer := p.otoPlayer
	currentStreamer := p.streamer
	p.mu.Unlock()

	if currentPlayer == nil || currentStreamer == nil {
		return
	}

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
		p.mu.Unlock()

		if !currentPlayer.IsPlaying() {
			// If playback stopped because of a manual pause, keep waiting.
			if isPaused {
				time.Sleep(250 * time.Millisecond)
				continue
			}

			// If oto reports it's not playing, check if we've actually reached the end of the stream.
			// Sometimes otoPlayer.IsPlaying() can be briefly false if the buffer is starved or hasn't started yet.
			if currentStreamer.IsEOF() {
				break // Stream is finished and oto has drained its buffer
			}
		}
		time.Sleep(250 * time.Millisecond)
	}

	slog.Info("Track playback finished naturally", "track", p.currentTrack.Title)

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
