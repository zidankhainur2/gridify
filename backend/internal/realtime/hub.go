package realtime

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Event represents a realtime event sent to tournament spectators/organizers.
type Event struct {
	Type         string    `json:"type"`
	TournamentID uuid.UUID `json:"tournament_id"`
	Data         any       `json:"data"`
	Timestamp    time.Time `json:"timestamp"`
}

// Hub manages active SSE subscribers per tournament.
type Hub struct {
	mu          sync.RWMutex
	subscribers map[uuid.UUID]map[chan Event]struct{}
}

// NewHub creates a new in-memory SSE broker hub.
func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[uuid.UUID]map[chan Event]struct{}),
	}
}

// Subscribe registers a new subscriber channel for a tournament.
func (h *Hub) Subscribe(tournamentID uuid.UUID) chan Event {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan Event, 32)
	if _, ok := h.subscribers[tournamentID]; !ok {
		h.subscribers[tournamentID] = make(map[chan Event]struct{})
	}
	h.subscribers[tournamentID][ch] = struct{}{}
	return ch
}

// Unsubscribe removes a subscriber channel for a tournament and closes it.
func (h *Hub) Unsubscribe(tournamentID uuid.UUID, ch chan Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if subs, ok := h.subscribers[tournamentID]; ok {
		if _, exists := subs[ch]; exists {
			delete(subs, ch)
			close(ch)
		}
		if len(subs) == 0 {
			delete(h.subscribers, tournamentID)
		}
	}
}

// Broadcast sends an event to all subscribers of a specific tournament.
func (h *Hub) Broadcast(tournamentID uuid.UUID, eventType string, data any) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	subs, ok := h.subscribers[tournamentID]
	if !ok || len(subs) == 0 {
		return
	}

	evt := Event{
		Type:         eventType,
		TournamentID: tournamentID,
		Data:         data,
		Timestamp:    time.Now().UTC(),
	}

	for ch := range subs {
		select {
		case ch <- evt:
		default:
			// Buffer full, skip to avoid blocking the broadcaster
		}
	}
}

// SubscriberCount returns the number of active listeners for a tournament.
func (h *Hub) SubscriberCount(tournamentID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if subs, ok := h.subscribers[tournamentID]; ok {
		return len(subs)
	}
	return 0
}
