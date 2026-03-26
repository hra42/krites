package benchmark

import "sync"

// SSEEventType identifies the kind of SSE event.
type SSEEventType string

const (
	EventRunStarted      SSEEventType = "run_started"
	EventResultCompleted SSEEventType = "result_completed"
	EventJudgeScored     SSEEventType = "judge_scored"
	EventRunCompleted    SSEEventType = "run_completed"
	EventRunError        SSEEventType = "run_error"
)

// SSEEvent is a single event sent to subscribers.
type SSEEvent struct {
	Type SSEEventType `json:"type"`
	Data interface{}  `json:"data"`
}

// Broadcaster manages per-run subscriber channels for SSE streaming.
type Broadcaster struct {
	mu          sync.RWMutex
	subscribers map[string][]chan SSEEvent
}

// NewBroadcaster creates a new Broadcaster.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		subscribers: make(map[string][]chan SSEEvent),
	}
}

// Subscribe creates a buffered channel for the given run and returns it
// along with an unsubscribe function. The caller must call unsubscribe
// when done (e.g. when the HTTP client disconnects).
func (b *Broadcaster) Subscribe(runID string) (<-chan SSEEvent, func()) {
	ch := make(chan SSEEvent, 16)

	b.mu.Lock()
	b.subscribers[runID] = append(b.subscribers[runID], ch)
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		subs := b.subscribers[runID]
		for i, s := range subs {
			if s == ch {
				b.subscribers[runID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}

	return ch, unsubscribe
}

// Publish sends an event to all subscribers of the given run.
// Non-blocking: if a subscriber's channel is full, that event is dropped.
func (b *Broadcaster) Publish(runID string, event SSEEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subscribers[runID] {
		select {
		case ch <- event:
		default:
		}
	}
}

// CloseRun closes all subscriber channels for a run and removes the entry.
func (b *Broadcaster) CloseRun(runID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ch := range b.subscribers[runID] {
		close(ch)
	}
	delete(b.subscribers, runID)
}
