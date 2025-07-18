package storage

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"gocloud.dev/docstore"
	_ "gocloud.dev/docstore/awsdynamodb"
	_ "gocloud.dev/docstore/memdocstore"
	"gocloud.dev/gcerrors"

	"git-telegram-bot/internal/config"
)

// Pipeline represents a GitLab pipeline with its associated Telegram message
type Pipeline struct {
	PipelineUpdateKey string    `docstore:"pipeline_update_key"` // Partition Key (S) - hash of pipeline URL + chat ID
	MessageID         int       `docstore:"message_id"`          // Telegram message ID
	CreatedAt         time.Time `docstore:"created_at"`
	UpdatedAt         time.Time `docstore:"updated_at"`
	ExpiresAt         int64     `docstore:"expires_at"` // TTL timestamp in epoch seconds
}

// PipelineStorage handles pipeline persistence
type PipelineStorage struct {
	collection *docstore.Collection
}

// NewPipelineStorage creates a new pipeline storage instance
func NewPipelineStorage(ctx context.Context) (*PipelineStorage, error) {
	var connectionString string
	if config.Global.StorageConnectionStringBase != "" {
		connectionString = config.Global.StorageConnectionStringBase + "-pipelines?partition_key=pipeline_update_key"
	} else {
		connectionString = "mem://pipelines/pipeline_update_key"
	}

	collection, err := docstore.OpenCollection(ctx, connectionString)
	if err != nil {
		return nil, err
	}

	return &PipelineStorage{
		collection: collection,
	}, nil
}

// CreatePipelineUpdateKey creates a composite hash from pipeline URL and chat ID
func CreatePipelineUpdateKey(pipelineURL string, chatID int64) string {
	data := fmt.Sprintf("%s:%d", pipelineURL, chatID)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// SavePipeline saves or replaces a pipeline (Put operation only)
func (s *PipelineStorage) SavePipeline(ctx context.Context, pipeline *Pipeline) error {
	now := time.Now()
	if pipeline.CreatedAt.IsZero() {
		pipeline.CreatedAt = now
	}
	pipeline.UpdatedAt = now
	pipeline.ExpiresAt = now.Add(time.Hour * 24).Unix()
	return s.collection.Put(ctx, pipeline)
}

// GetPipeline retrieves a pipeline by its update key
func (s *PipelineStorage) GetPipeline(ctx context.Context, pipelineUpdateKey string) (*Pipeline, error) {
	pipeline := &Pipeline{PipelineUpdateKey: pipelineUpdateKey}
	err := s.collection.Get(ctx, pipeline)

	if err == nil {
		return pipeline, nil
	} else if gcerrors.Code(err) == gcerrors.NotFound {
		return nil, nil
	} else {
		return nil, err
	}
}

// Close closes the storage connection
func (s *PipelineStorage) Close() error {
	if s.collection != nil {
		return s.collection.Close()
	}
	return nil
}
