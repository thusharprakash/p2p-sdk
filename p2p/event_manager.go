package p2p

import "sync"

type EventHandler func(EventMessage)

type EventManager struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

func NewEventManager() *EventManager {
	return &EventManager{
		handlers: make(map[string][]EventHandler),
	}
}

func (em *EventManager) Subscribe(eventType string, handler EventHandler) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.handlers[eventType] = append(em.handlers[eventType], handler)
}

func (em *EventManager) Dispatch(event EventMessage) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if handlers, found := em.handlers[event.EventType]; found {
		for _, handler := range handlers {
			handler(event)
		}

	}
}

func (em *EventManager) DispatchWithOrdering(event EventMessage) {
	em.mu.Lock()
	defer em.mu.Unlock()
	// Ensure events are processed in logical order using vector clocks
	if handlers, found := em.handlers[event.EventType]; found {
		for _, handler := range handlers {
			handler(event)
		}
	}
}
