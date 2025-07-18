package storage

import (
	"context"
	"time"

	"gocloud.dev/docstore"
	_ "gocloud.dev/docstore/awsdynamodb"
	_ "gocloud.dev/docstore/memdocstore"
	"gocloud.dev/gcerrors"

	"git-telegram-bot/internal/config"
)

// Chat represents a Telegram chat where the bot has been added
type Chat struct {
	ChatID    int64     `docstore:"chat_id"`  // Partition Key (N)
	BotType   string    `docstore:"bot_type"` // Sort Key (S)
	CreatedAt time.Time `docstore:"created_at"`
	UpdatedAt time.Time `docstore:"updated_at"`
}

// ChatStorage handles chat persistence
type ChatStorage struct {
	collection *docstore.Collection
}

// NewChatStorage creates a new chat storage instance
func NewChatStorage(ctx context.Context) (*ChatStorage, error) {
	var connectionString string
	if config.Global.StorageConnectionStringBase != "" {
		connectionString = config.Global.StorageConnectionStringBase + "-chats?partition_key=chat_id&sort_key=bot_type"
	} else {
		connectionString = "mem://chats/chat_id"
	}

	collection, err := docstore.OpenCollection(ctx, connectionString)
	if err != nil {
		return nil, err
	}

	return &ChatStorage{
		collection: collection,
	}, nil
}

// SaveChat saves or updates a chat using update-or-create logic
func (s *ChatStorage) SaveChat(ctx context.Context, chat *Chat) error {
	now := time.Now()

	err := s.collection.Update(ctx, chat, docstore.Mods{"updated_at": now})
	if err == nil {
		return nil
	}

	code := gcerrors.Code(err)
	// In DynamoDB, this is FailedPrecondition instead of NotFound
	if code == gcerrors.NotFound || code == gcerrors.FailedPrecondition {
		chat.CreatedAt = now
		chat.UpdatedAt = now
		return s.collection.Put(ctx, chat)
	}

	return err
}

func (s *ChatStorage) DeleteChat(ctx context.Context, chat *Chat) error {
	return s.collection.Delete(ctx, chat)
}

// Close closes the storage connection
func (s *ChatStorage) Close() error {
	if s.collection != nil {
		return s.collection.Close()
	}
	return nil
}
