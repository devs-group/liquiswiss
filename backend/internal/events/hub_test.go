package events

import (
	"testing"
	"time"
)

func TestHubPublishReachesSubscriber(t *testing.T) {
	hub := NewHub()
	sub := hub.Subscribe(1)
	defer sub.Close()

	hub.Publish(Event{Entity: "transaction", Action: ActionCreated, ID: 42, OrganisationID: 7})

	select {
	case event := <-sub.Events:
		if event.Entity != "transaction" || event.ID != 42 || event.OrganisationID != 7 {
			t.Fatalf("unexpected event: %+v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("expected event, got none")
	}
}

func TestHubConnectionCapEvictsOldest(t *testing.T) {
	hub := NewHub()
	subs := make([]*Subscription, 0, MaxConnectionsPerUser)
	for range MaxConnectionsPerUser {
		subs = append(subs, hub.Subscribe(1))
	}

	extra := hub.Subscribe(1)
	if extra == nil {
		t.Fatal("subscription above cap must evict, not fail")
	}

	select {
	case <-subs[0].Done:
	case <-time.After(time.Second):
		t.Fatal("oldest subscription not evicted at cap")
	}
	for _, sub := range subs[1:] {
		select {
		case <-sub.Done:
			t.Fatal("newer subscription wrongly evicted")
		default:
		}
	}
	extra.Close()
}

func TestHubOverflowDoesNotBlockPublisher(t *testing.T) {
	hub := NewHub()
	sub := hub.Subscribe(1)
	defer sub.Close()

	done := make(chan struct{})
	go func() {
		for i := range subscriberBuffer * 2 {
			hub.Publish(Event{Entity: "transaction", Action: ActionUpdated, ID: int64(i)})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("publisher blocked on full subscriber buffer")
	}
}

func TestHubCloseUserSignalsDone(t *testing.T) {
	hub := NewHub()
	sub := hub.Subscribe(1)
	other := hub.Subscribe(2)
	defer other.Close()

	hub.CloseUser(1)

	select {
	case <-sub.Done:
	case <-time.After(time.Second):
		t.Fatal("expected Done to be closed")
	}
	select {
	case <-other.Done:
		t.Fatal("other user's subscription must stay open")
	default:
	}
}
