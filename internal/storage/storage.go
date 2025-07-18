package storage

import (
	"context"
	"fmt"
	"io"
)

// Storage provides centralized access to all storage types
type Storage struct {
	ChatStorage     *ChatStorage
	PipelineStorage *PipelineStorage
}

// NewStorage creates a new centralized storage instance
func NewStorage(ctx context.Context) (*Storage, error) {
	chatStorage, err := NewChatStorage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize chat storage: %w", err)
	}

	pipelineStorage, err := NewPipelineStorage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize pipeline storage: %w", err)
	}

	return &Storage{
		ChatStorage:     chatStorage,
		PipelineStorage: pipelineStorage,
	}, nil
}

// Close closes all storage connections and returns the first error encountered.
func (s *Storage) Close() error {
	var firstErr error

	// Define all closable storages in a slice
	closers := []io.Closer{
		s.ChatStorage,
		s.PipelineStorage,
		// Add more storages here as needed
	}

	// Close each and track the first error
	for _, closer := range closers {
		if closer != nil {
			if err := closer.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}
