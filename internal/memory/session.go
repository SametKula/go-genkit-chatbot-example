package memory

import "context"

// Message represents a single chat message.
type Message struct {
	Role    string // e.g., "user", "model", "system"
	Content string
}

// SessionStore defines the interface for managing chat histories.
type SessionStore interface {
	SaveMessage(ctx context.Context, sessionID string, msg Message) error
	GetHistory(ctx context.Context, sessionID string) ([]Message, error)
}
