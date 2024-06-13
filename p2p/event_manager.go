package p2p

import (
	"context"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type EventManager struct {
	mu            sync.Mutex
	eventHandlers map[string]func(EventMessage)
	VectorClock   VectorClock
}

func NewEventManager() *EventManager {
	return &EventManager{
		eventHandlers: make(map[string]func(EventMessage)),
		VectorClock:   make(VectorClock),
	}
}

func (em *EventManager) RegisterEventHandler(eventType string, handler func(EventMessage)) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.eventHandlers[eventType] = handler
}

func (em *EventManager) Dispatch(event EventMessage) {
	em.mu.Lock()
	defer em.mu.Unlock()
	if handler, ok := em.eventHandlers[event.EventType]; ok {
		handler(event)
	}
}

func (em *EventManager) DispatchWithOrdering(event EventMessage) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.VectorClock.Increment(event.SenderID)
	em.VectorClock.Update(event.VectorClock)
	if handler, ok := em.eventHandlers[event.EventType]; ok {
		handler(event)
	}
	// publish the events to the room

}

func (em *EventManager) GossipEvent(ctx context.Context, event EventMessage, peers []peer.ID, host host.Host) {
	for _, peerID := range peers {
		if peerID == peer.ID(event.SenderID) {
			continue
		}
		go func(p peer.ID) {
			// Simulate gossiping event
			fmt.Printf("Gossiping event to peer %s\n", p.String())
			// Normally, send the event to the peer
			// Use the host to send the event
			// Example: host.SendMessage(p, event)
		}(peerID)
	}
}
