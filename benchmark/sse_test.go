package benchmark

import (
	"testing"
	"time"
)

func TestBroadcaster_SubscribeAndPublish(t *testing.T) {
	b := NewBroadcaster()
	ch, unsub := b.Subscribe("run1")
	defer unsub()

	event := SSEEvent{Type: EventRunStarted, Data: "test"}
	b.Publish("run1", event)

	select {
	case got := <-ch:
		if got.Type != EventRunStarted {
			t.Errorf("type = %v, want %v", got.Type, EventRunStarted)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestBroadcaster_MultipleSubscribers(t *testing.T) {
	b := NewBroadcaster()
	ch1, unsub1 := b.Subscribe("run1")
	defer unsub1()
	ch2, unsub2 := b.Subscribe("run1")
	defer unsub2()

	event := SSEEvent{Type: EventResultCompleted, Data: "result"}
	b.Publish("run1", event)

	for i, ch := range []<-chan SSEEvent{ch1, ch2} {
		select {
		case got := <-ch:
			if got.Type != EventResultCompleted {
				t.Errorf("subscriber %d: type = %v, want %v", i, got.Type, EventResultCompleted)
			}
		case <-time.After(time.Second):
			t.Fatalf("subscriber %d: timed out", i)
		}
	}
}

func TestBroadcaster_CloseRun(t *testing.T) {
	b := NewBroadcaster()
	ch, _ := b.Subscribe("run1")

	b.CloseRun("run1")

	// Channel should be closed, range should end
	_, ok := <-ch
	if ok {
		t.Error("expected channel to be closed")
	}
}

func TestBroadcaster_Unsubscribe(t *testing.T) {
	b := NewBroadcaster()
	ch1, unsub1 := b.Subscribe("run1")
	ch2, unsub2 := b.Subscribe("run1")
	defer unsub2()

	unsub1()

	event := SSEEvent{Type: EventRunCompleted, Data: "done"}
	b.Publish("run1", event)

	// ch2 should receive the event
	select {
	case got := <-ch2:
		if got.Type != EventRunCompleted {
			t.Errorf("type = %v, want %v", got.Type, EventRunCompleted)
		}
	case <-time.After(time.Second):
		t.Fatal("ch2 timed out")
	}

	// ch1 should not receive anything (unsubscribed)
	select {
	case <-ch1:
		t.Error("ch1 should not receive events after unsubscribe")
	case <-time.After(50 * time.Millisecond):
		// expected
	}
}

func TestBroadcaster_PublishToNonexistentRun(t *testing.T) {
	b := NewBroadcaster()
	// Should not panic
	b.Publish("nonexistent", SSEEvent{Type: EventRunError, Data: "err"})
}
