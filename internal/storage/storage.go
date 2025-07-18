package storage

import (
	"context"
	"fmt"
)

// Storage provides centralized access to all storage types
type Storage struct {
	ChatStorage *ChatStorage
}

// NewStorage creates a new centralized storage instance
func NewStorage(ctx context.Context) (*Storage, error) {
	chatStorage, err := NewChatStorage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize chat storage: %w", err)
	}

	return &Storage{
		ChatStorage: chatStorage,
	}, nil
}

// Close closes all storage connections
func (s *Storage) Close() error {
	if s.ChatStorage != nil {
		return s.ChatStorage.Close()
	}
	return nil
}
