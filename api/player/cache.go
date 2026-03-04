package player

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

type DownloadTask struct {
	done chan struct{}
	err  error
}

// CacheManager handles background and synchronous downloading of audio tracks via yt-dlp.
type CacheManager struct {
	cacheDir   string
	mu         sync.Mutex
	inProgress map[string]*DownloadTask
	queue      chan string
}

var DefaultCacheManager *CacheManager

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error("failed to get user home directory for cache", "error", err)
		return
	}
	cacheDir := filepath.Join(home, ".ytmusic", "cache")
	DefaultCacheManager = NewCacheManager(cacheDir)
}

// NewCacheManager initializes a new CacheManager and starts background workers.
func NewCacheManager(cacheDir string) *CacheManager {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		slog.Error("failed to create cache directory", "error", err)
	}

	cm := &CacheManager{
		cacheDir:   cacheDir,
		inProgress: make(map[string]*DownloadTask),
		// buffered channel to hold download requests
		queue: make(chan string, 10000),
	}

	// start 3 concurrent download workers
	for i := 0; i < 3; i++ {
		go cm.worker()
	}

	return cm
}

// CacheSync blocks until the video is cached. If it is already being cached by a background worker, it waits for it.
func (cm *CacheManager) CacheSync(videoID string) error {
	if cm.IsCached(videoID) {
		return nil
	}

	cm.mu.Lock()
	task, exists := cm.inProgress[videoID]
	if !exists {
		task = &DownloadTask{done: make(chan struct{})}
		cm.inProgress[videoID] = task
		cm.mu.Unlock()

		// Run download synchronously
		err := cm.download(videoID)
		task.err = err
		close(task.done)

		cm.mu.Lock()
		delete(cm.inProgress, videoID)
		cm.mu.Unlock()

		return err
	}
	cm.mu.Unlock()

	// Wait for existing task to finish
	slog.Info("Track is already downloading, waiting for it to finish", "videoID", videoID)
	<-task.done
	if cm.IsCached(videoID) {
		return nil
	}
	return task.err
}

// QueueDownload adds a video ID to the background download queue if it's not already cached.
func (cm *CacheManager) QueueDownload(videoID string) {
	if cm.IsCached(videoID) {
		return
	}

	cm.mu.Lock()
	if _, exists := cm.inProgress[videoID]; exists {
		cm.mu.Unlock()
		return
	}
	task := &DownloadTask{done: make(chan struct{})}
	cm.inProgress[videoID] = task
	cm.mu.Unlock()

	// add to channel without blocking
	select {
	case cm.queue <- videoID:
	default:
		slog.Warn("cache queue is full, skipping", "videoID", videoID)
		cm.mu.Lock()
		delete(cm.inProgress, videoID)
		cm.mu.Unlock()
	}
}

// IsCached returns true if the audio file exists in the cache directory.
func (cm *CacheManager) IsCached(videoID string) bool {
	path := cm.GetCachedPath(videoID)
	_, err := os.Stat(path)
	return err == nil
}

// GetCachedPath returns the absolute path to the cached audio file.
func (cm *CacheManager) GetCachedPath(videoID string) string {
	return filepath.Join(cm.cacheDir, videoID+".m4a")
}

func (cm *CacheManager) download(videoID string) error {
	slog.Info("Caching track via yt-dlp", "videoID", videoID)

	path := cm.GetCachedPath(videoID)

	// Use yt-dlp to download and convert to m4a
	cmd := exec.Command("yt-dlp",
		"--quiet", "--no-warnings",
		"-x", "--audio-format", "m4a",
		"-o", path,
		"https://music.youtube.com/watch?v="+videoID,
	)

	if err := cmd.Run(); err != nil {
		slog.Error("Failed to cache track", "videoID", videoID, "error", err)
		// clean up any partial file
		_ = os.Remove(path)
		return err
	}

	slog.Info("Successfully cached track", "videoID", videoID)
	return nil
}

func (cm *CacheManager) worker() {
	for videoID := range cm.queue {
		cm.mu.Lock()
		task, exists := cm.inProgress[videoID]
		cm.mu.Unlock()

		if !exists {
			// Already downloaded by CacheSync, or some other edge case
			continue
		}

		err := cm.download(videoID)
		task.err = err
		close(task.done)

		cm.mu.Lock()
		delete(cm.inProgress, videoID)
		cm.mu.Unlock()
	}
}
