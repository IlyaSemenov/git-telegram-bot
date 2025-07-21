package telegram

import (
	"context"
	"fmt"

	"git-telegram-bot/internal/config"
	"git-telegram-bot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// GitLabTelegramService is a Telegram service for GitLab notifications
type GitLabTelegramService struct {
	*BaseTelegramService
	storage *storage.Storage
}

// NewGitLabTelegramService creates a new GitLab Telegram service
func NewGitLabTelegramService(storageInstance *storage.Storage) (*GitLabTelegramService, error) {
	if config.Global.GitLabTelegramBotToken == "" {
		return nil, nil
	}

	base, err := NewBaseTelegramService("gitlab", config.Global.GitLabTelegramBotToken, storageInstance)
	if err != nil {
		return nil, err
	}

	return &GitLabTelegramService{
		BaseTelegramService: base,
		storage:             storageInstance,
	}, nil
}

// Init initializes the GitLab Telegram bot
func (s *GitLabTelegramService) Init() error {
	return s.BaseTelegramService.Init([]tgbotapi.BotCommand{
		{
			Command:     "start",
			Description: "Start the bot",
		},
		{
			Command:     "help",
			Description: "Show help information",
		},
		{
			Command:     "webhook",
			Description: "Get your unique GitLab webhook URL",
		},
	})
}

// ProcessUpdate processes a Telegram update
func (s *GitLabTelegramService) ProcessUpdate(updateJSON []byte) error {
	return s.BaseTelegramService.ProcessUpdate(updateJSON, s.handleUpdate)
}

// handleUpdate handles a Telegram update
func (s *GitLabTelegramService) handleUpdate(update *tgbotapi.Update) error {
	// Handle commands
	if update.Message != nil && update.Message.IsCommand() {
		return s.handleCommand(update.Message)
	}

	return nil
}

// handleCommand handles a Telegram command
func (s *GitLabTelegramService) handleCommand(message *tgbotapi.Message) error {
	switch message.Command() {
	case "start":
		return s.handleStartCommand(message)
	case "help":
		return s.handleHelpCommand(message)
	case "webhook":
		return s.handleWebhookCommand(message)
	default:
		return s.SendMessage(message.Chat.ID, "Unknown command. Type /help for available commands.")
	}
}

// handleStartCommand handles the /start command
func (s *GitLabTelegramService) handleStartCommand(message *tgbotapi.Message) error {
	text := "👋 Welcome to GitLab Watch Bot!\n\n" +
		"I can forward GitLab webhook events to this chat.\n\n" +
		"Use /webhook to get your unique webhook URL."

	return s.SendMessage(message.Chat.ID, text)
}

// SendOrUpdatePipelineMessage updates an existing pipeline message or creates a new one
func (s *GitLabTelegramService) SendOrUpdatePipelineMessage(chatID int64, pipelineURL string, text string) error {
	ctx := context.Background()
	pipelineUpdateKey := storage.CreatePipelineUpdateKey(pipelineURL, chatID)

	// Get existing pipeline mapping
	pipeline, err := s.storage.PipelineStorage.GetPipeline(ctx, pipelineUpdateKey)
	if err != nil {
		return err
	}

	if pipeline == nil {
		// Pipeline not found, send new message
		msg, err := s.SendMessageWithResult(chatID, text)
		if err != nil {
			return err
		}

		// Store pipeline mapping
		pipeline := &storage.Pipeline{
			PipelineUpdateKey: pipelineUpdateKey,
			MessageID:         msg.MessageID,
		}
		return s.storage.PipelineStorage.SavePipeline(ctx, pipeline)
	} else {
		// Update the existing message
		if err := s.UpdateMessage(chatID, pipeline.MessageID, text); err != nil {
			return err
		}
		// Update the mapping timestamp
		return s.storage.PipelineStorage.SavePipeline(ctx, pipeline)
	}
}

// handleHelpCommand handles the /help command
func (s *GitLabTelegramService) handleHelpCommand(message *tgbotapi.Message) error {
	text := "📚 <b>Available Commands</b>\n\n" +
		"• /start - Start the bot\n" +
		"• /help - Show this help message\n" +
		"• /webhook - Get your unique GitLab webhook URL\n\n" +
		"To set up webhooks, use the appropriate command and add the URL to your repository's webhook settings."

	return s.SendMessage(message.Chat.ID, text)
}

// handleGitLabCommand handles the /gitlab command
func (s *GitLabTelegramService) handleWebhookCommand(message *tgbotapi.Message) error {
	webhookURL, err := s.GetChatWebhookURL(message.Chat.ID)
	if err != nil {
		return err
	}

	// Create response message
	text := fmt.Sprintf("🔗 <b>Your GitLab Webhook URL</b>\n\n<code>%s</code>\n\n", webhookURL) +
		"<b>How to set up:</b>\n\n" +
		"1. Go to your GitLab project\n" +
		"2. Click on Settings > Webhooks\n" +
		"3. Paste the URL above in the 'URL' field\n" +
		"4. Select the events you want to receive\n" +
		"5. Click 'Add webhook'\n\n" +
		"You'll receive a confirmation message when the webhook is set up correctly."

	return s.SendMessage(message.Chat.ID, text)
}
