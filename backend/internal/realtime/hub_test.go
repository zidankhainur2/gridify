package realtime

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHub_SubscribeAndBroadcast(t *testing.T) {
	hub := NewHub()
	tournamentID := uuid.New()

	ch1 := hub.Subscribe(tournamentID)
	ch2 := hub.Subscribe(tournamentID)

	if count := hub.SubscriberCount(tournamentID); count != 2 {
		t.Fatalf("expected 2 subscribers, got %d", count)
	}

	payload := map[string]string{"message": "match finished"}
	hub.Broadcast(tournamentID, "match_updated", payload)

	select {
	case evt := <-ch1:
		if evt.Type != "match_updated" {
			t.Errorf("expected event type match_updated, got %s", evt.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("ch1 timed out waiting for event")
	}

	select {
	case evt := <-ch2:
		if evt.Type != "match_updated" {
			t.Errorf("expected event type match_updated, got %s", evt.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("ch2 timed out waiting for event")
	}

	hub.Unsubscribe(tournamentID, ch1)
	if count := hub.SubscriberCount(tournamentID); count != 1 {
		t.Fatalf("expected 1 subscriber, got %d", count)
	}

	hub.Unsubscribe(tournamentID, ch2)
	if count := hub.SubscriberCount(tournamentID); count != 0 {
		t.Fatalf("expected 0 subscribers, got %d", count)
	}
}

func TestHub_DifferentTournaments(t *testing.T) {
	hub := NewHub()
	t1 := uuid.New()
	t2 := uuid.New()

	ch1 := hub.Subscribe(t1)
	ch2 := hub.Subscribe(t2)
	defer hub.Unsubscribe(t1, ch1)
	defer hub.Unsubscribe(t2, ch2)

	hub.Broadcast(t1, "stage_updated", "stage 1")

	select {
	case evt := <-ch1:
		if evt.Type != "stage_updated" {
			t.Errorf("unexpected event: %v", evt)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("ch1 did not receive event")
	}

	select {
	case evt := <-ch2:
		t.Fatalf("ch2 should not receive event for t1, got: %v", evt)
	case <-time.After(50 * time.Millisecond):
		// Expected
	}
}

func TestHub_ConcurrentOperations(t *testing.T) {
	hub := NewHub()
	tournamentID := uuid.New()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch := hub.Subscribe(tournamentID)
			time.Sleep(10 * time.Millisecond)
			hub.Unsubscribe(tournamentID, ch)
		}()
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			hub.Broadcast(tournamentID, "ping", idx)
		}(i)
	}

	wg.Wait()
}
