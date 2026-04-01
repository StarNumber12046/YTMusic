package player

import (
	"math/rand"
	"sync"

	"ytmusic_api/models"
)

// Queue manages an ordered list of tracks for playback.
type Queue struct {
	mu            sync.RWMutex
	items         []models.Track
	position      int            // index of the currently active track
	originalOrder []models.Track // stored for unshuffle
}

// NewQueue creates an empty queue.
func NewQueue() *Queue {
	return &Queue{
		items:    make([]models.Track, 0),
		position: -1,
	}
}

// Add appends a track to the end of the queue.
func (q *Queue) Add(track models.Track) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, track)
	// If queue was empty, set position to the first item
	if len(q.items) == 1 {
		q.position = 0
	}
	if DefaultCacheManager != nil {
		DefaultCacheManager.QueueDownload(track.VideoID)
	}
}

// AddAll appends multiple tracks to the queue.
func (q *Queue) AddAll(tracks []models.Track) {
	q.mu.Lock()
	defer q.mu.Unlock()
	wasEmpty := len(q.items) == 0
	q.items = append(q.items, tracks...)
	if wasEmpty && len(q.items) > 0 {
		q.position = 0
	}
	if DefaultCacheManager != nil {
		for _, track := range tracks {
			DefaultCacheManager.QueueDownload(track.VideoID)
		}
	}
}

// Remove removes the track at the given position (0-indexed).
func (q *Queue) Remove(position int) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if position < 0 || position >= len(q.items) {
		return false
	}

	q.items = append(q.items[:position], q.items[position+1:]...)

	// Adjust current position
	if len(q.items) == 0 {
		q.position = -1
	} else if position < q.position {
		q.position--
	} else if q.position >= len(q.items) {
		q.position = len(q.items) - 1
	}

	return true
}

// Clear removes all tracks and resets position.
func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = make([]models.Track, 0)
	q.position = -1
}

// Current returns the track at the current position, or nil if empty.
func (q *Queue) Current() *models.Track {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.position < 0 || q.position >= len(q.items) {
		return nil
	}
	track := q.items[q.position]
	return &track
}

// Next advances to the next track and returns it. Returns nil if at end.
func (q *Queue) Next() *models.Track {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.position+1 >= len(q.items) {
		return nil
	}
	q.position++
	track := q.items[q.position]
	return &track
}

// Previous goes back to the previous track and returns it. Returns nil if at start.
func (q *Queue) Previous() *models.Track {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.position <= 0 {
		return nil
	}
	q.position--
	track := q.items[q.position]
	return &track
}

// SetPosition sets the queue position to a specific index.
func (q *Queue) SetPosition(pos int) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if pos < 0 || pos >= len(q.items) {
		return false
	}
	q.position = pos
	return true
}

// Position returns the current queue position.
func (q *Queue) Position() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.position
}

// Len returns the number of items in the queue.
func (q *Queue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.items)
}

// Items returns a copy of all queue items.
func (q *Queue) Items() []models.Track {
	q.mu.RLock()
	defer q.mu.RUnlock()
	result := make([]models.Track, len(q.items))
	copy(result, q.items)
	return result
}

// ReplaceAll clears the queue and replaces it with the given tracks, setting position to 0.
func (q *Queue) ReplaceAll(tracks []models.Track) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = make([]models.Track, len(tracks))
	copy(q.items, tracks)
	if len(q.items) > 0 {
		q.position = 0
	} else {
		q.position = -1
	}
	if DefaultCacheManager != nil {
		for _, track := range tracks {
			DefaultCacheManager.QueueDownload(track.VideoID)
		}
	}
}

// Shuffle shuffles the queue items (keeping current track at position).
func (q *Queue) Shuffle() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) <= 1 {
		return
	}

	// Store original order before shuffling
	q.originalOrder = make([]models.Track, len(q.items))
	copy(q.originalOrder, q.items)

	// Shuffle items after current position
	currentPos := q.position
	if currentPos < 0 {
		currentPos = 0
	}
	if currentPos < len(q.items)-1 {
		// Fisher-Yates shuffle for items after current
		for i := len(q.items) - 1; i > currentPos; i-- {
			j := rand.Intn(i - currentPos + 1)
			q.items[i], q.items[currentPos+j] = q.items[currentPos+j], q.items[i]
		}
	}
}

// Unshuffle restores the original queue order before shuffle.
func (q *Queue) Unshuffle() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.originalOrder == nil {
		return
	}

	// Find current track in original order to restore position
	currentTrack := ""
	if q.position >= 0 && q.position < len(q.items) {
		currentTrack = q.items[q.position].VideoID
	}

	q.items = make([]models.Track, len(q.originalOrder))
	copy(q.items, q.originalOrder)
	q.originalOrder = nil

	// Restore position based on current track
	if currentTrack != "" {
		for i, track := range q.items {
			if track.VideoID == currentTrack {
				q.position = i
				break
			}
		}
	}
}

// PlayNext inserts a track right after the current position (or at the end if queue is empty).
// The track will become the next track to play.
func (q *Queue) PlayNext(track models.Track) {
	q.mu.Lock()
	defer q.mu.Unlock()

	insertPos := q.position + 1
	if insertPos > len(q.items) {
		insertPos = len(q.items)
	}

	// Insert track at the specified position
	items := make([]models.Track, len(q.items)+1)
	copy(items[:insertPos], q.items[:insertPos])
	items[insertPos] = track
	copy(items[insertPos+1:], q.items[insertPos:])
	q.items = items

	// If queue was empty, set position to 0
	if len(q.items) == 1 {
		q.position = 0
	} else if q.position >= insertPos {
		// If we inserted before or at current position, advance position
		q.position++
	}

	if DefaultCacheManager != nil {
		DefaultCacheManager.QueueDownload(track.VideoID)
	}
}
