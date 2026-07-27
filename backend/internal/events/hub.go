package events

import (
	"sync"
)

// Action constants for entity events
const (
	ActionCreated = "created"
	ActionUpdated = "updated"
	ActionDeleted = "deleted"
)

// Event is a minimal change notification. Clients refetch data through the
// REST API, so events never carry entity payloads.
type Event struct {
	Entity string `json:"entity"`
	Action string `json:"action"`
	ID     int64  `json:"id,omitempty"`
	// ParentID links sub-entities to their parent (e.g. salary cost → salary)
	// so clients can highlight the owning element
	ParentID int64 `json:"parentId,omitempty"`
	// OrganisationID is used server-side for delivery filtering and is not serialized
	OrganisationID int64 `json:"-"`
	// Origin identifies who caused the change (user + browser tab). Used
	// server-side to compute Own per connection; never serialized to clients.
	OriginUserID   int64  `json:"-"`
	OriginClientID string `json:"-"`
	// Own is set at delivery time per connection: true when this exact
	// connection (user + client id) caused the change
	Own bool `json:"own"`
}

// subscriberBuffer bounds the per-connection channel; slow consumers miss
// events instead of blocking publishers (clients recover via reconnect)
const subscriberBuffer = 32

// MaxConnectionsPerUser caps concurrent event streams per user
const MaxConnectionsPerUser = 5

type Subscription struct {
	// Events is never closed; consumers must also select on Done
	Events chan Event
	// Done is closed when the subscription is terminated (unsubscribe or forced close)
	Done   chan struct{}
	userID int64
	seq    uint64
	hub    *Hub
	once   sync.Once
}

// Close unsubscribes and signals Done. Safe to call multiple times.
func (s *Subscription) Close() {
	s.once.Do(func() {
		s.hub.remove(s)
		close(s.Done)
	})
}

// Hub is an in-process pub/sub for change events, keyed by userID.
// Kept intentionally narrow so an external broker could replace it later.
type Hub struct {
	mu     sync.RWMutex
	subs   map[int64]map[*Subscription]struct{}
	nextID uint64
}

func NewHub() *Hub {
	return &Hub{
		subs: make(map[int64]map[*Subscription]struct{}),
	}
}

// Subscribe registers a new event stream for the given user. When the user is
// at MaxConnectionsPerUser, the oldest stream is evicted (stale tabs and
// abandoned connections must never lock a user out of real-time updates).
func (h *Hub) Subscribe(userID int64) *Subscription {
	h.mu.Lock()
	var evict *Subscription
	if len(h.subs[userID]) >= MaxConnectionsPerUser {
		for sub := range h.subs[userID] {
			if evict == nil || sub.seq < evict.seq {
				evict = sub
			}
		}
	}
	h.nextID++
	sub := &Subscription{
		Events: make(chan Event, subscriberBuffer),
		Done:   make(chan struct{}),
		userID: userID,
		seq:    h.nextID,
		hub:    h,
	}
	if h.subs[userID] == nil {
		h.subs[userID] = make(map[*Subscription]struct{})
	}
	h.subs[userID][sub] = struct{}{}
	h.mu.Unlock()

	if evict != nil {
		evict.Close()
	}
	return sub
}

func (h *Hub) remove(sub *Subscription) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.subs[sub.userID]; ok {
		delete(set, sub)
		if len(set) == 0 {
			delete(h.subs, sub.userID)
		}
	}
}

// Publish fans the event out to every subscriber. Delivery-time filtering by
// organisation happens in the SSE handler; a full buffer skips the subscriber
// instead of blocking the mutation path.
func (h *Hub) Publish(event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, set := range h.subs {
		for sub := range set {
			select {
			case sub.Events <- event:
			default:
			}
		}
	}
}

// CloseUser terminates all streams of a user (logout, org switch, member removal)
func (h *Hub) CloseUser(userID int64) {
	h.mu.RLock()
	subs := make([]*Subscription, 0, len(h.subs[userID]))
	for sub := range h.subs[userID] {
		subs = append(subs, sub)
	}
	h.mu.RUnlock()
	for _, sub := range subs {
		sub.Close()
	}
}
