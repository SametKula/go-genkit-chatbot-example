package memory

import (
	"context"
	"sync"
)

// InMemorySessionStore is a thread-safe, in-memory implementation of SessionStore.
type InMemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string][]Message
}

// NewInMemorySessionStore creates a new InMemorySessionStore.
func NewInMemorySessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{
		sessions: make(map[string][]Message),
	}
}

// SaveMessage appends a message to the session's history.
func (s *InMemorySessionStore) SaveMessage(ctx context.Context, sessionID string, msg Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[sessionID] = append(s.sessions[sessionID], msg)
	return nil
}

// GetHistory retrieves the chat history for a session.
func (s *InMemorySessionStore) GetHistory(ctx context.Context, sessionID string) ([]Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history := s.sessions[sessionID]
	
	// Return a copy to avoid external modifications and race conditions
	historyCopy := make([]Message, len(history))
	copy(historyCopy, history)

	return historyCopy, nil
}
