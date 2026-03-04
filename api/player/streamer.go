package player

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// Streamer wraps an ffmpeg subprocess that decodes an audio URL into raw PCM (s16le, 44100Hz, stereo).
type Streamer struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
	cancel context.CancelFunc

	mu    sync.Mutex
	isEOF bool
}

// NewStreamer spawns an ffmpeg process that reads from the given URL and outputs raw PCM to stdout.
// Output format: signed 16-bit little-endian, 44100 Hz sample rate, 2 channels (stereo).
func NewStreamer(audioURL string) (*Streamer, error) {
	ctx, cancel := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-reconnect", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "5",
		"-i", audioURL,
		"-f", "s16le", // raw PCM signed 16-bit LE
		"-ar", "44100", // sample rate
		"-ac", "2", // stereo
		"-loglevel", "error",
		"pipe:1", // output to stdout
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("creating stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("starting ffmpeg: %w", err)
	}

	return &Streamer{
		cmd:    cmd,
		stdout: stdout,
		cancel: cancel,
	}, nil
}

// Read implements io.Reader, reading decoded PCM data from ffmpeg's stdout.
func (s *Streamer) Read(p []byte) (int, error) {
	n, err := s.stdout.Read(p)
	if err != nil {
		s.mu.Lock()
		s.isEOF = true
		s.mu.Unlock()
	}
	return n, err
}

// IsEOF returns true if the ffmpeg stream has finished or encountered an error.
func (s *Streamer) IsEOF() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isEOF
}

// Close stops the ffmpeg process and releases resources.
func (s *Streamer) Close() error {
	s.cancel()
	// Close stdout pipe
	if s.stdout != nil {
		s.stdout.Close()
	}
	// Wait for process to exit (ignore error since we killed it)
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Wait()
	}
	return nil
}
