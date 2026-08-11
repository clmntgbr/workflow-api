package registry

import (
	"context"
	"fmt"
	"log"
	"sync"
)

type EventHandler func(ctx context.Context, payload []byte) error

type HandlerRegistry struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
}

func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{handlers: make(map[string][]EventHandler)}
}

func (r *HandlerRegistry) Register(eventType string, h EventHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[eventType] = append(r.handlers[eventType], h)
}

func (r *HandlerRegistry) Dispatch(ctx context.Context, eventType string, payload []byte) error {
	r.mu.RLock()
	handlers := append([]EventHandler(nil), r.handlers[eventType]...)
	r.mu.RUnlock()

	if len(handlers) == 0 {
		log.Printf("no handler registered for event_type=%s", eventType)
		return nil
	}

	for _, h := range handlers {
		if err := h(ctx, payload); err != nil {
			return fmt.Errorf("handler failed for %s: %w", eventType, err)
		}
	}
	return nil
}
