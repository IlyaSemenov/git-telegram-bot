package telegram

import (
	"context"
	"fmt"

	"git-telegram-bot/internal/config"
	"git-telegram-bot/internal/storage"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// GitLabTelegramService is a Telegram service for GitLab notifications
type GitLabTelegramService struct {
	*TelegramService
	pipelineStorage *storage.PipelineStorage
}

// NewGitLabTelegramService creates a new GitLab Telegram service
func NewGitLabTelegramService(storageInstance *storage.Storage) (*GitLabTelegramService, error) {
	s, err := NewTelegramService("gitlab", config.Global.GitLabTelegramBotToken, storageInstance)
	if s == nil || err != nil {
		return nil, err
	}

	gs := &GitLabTelegramService{
		TelegramService: s,
		pipelineStorage: storageInstance.PipelineStorage,
	}

	s.registerCommandHandler("start", gs.handleStartCommand)
	s.registerCommandHandler("help", gs.handleHelpCommand)
	s.registerCommandHandler("webhook", gs.handleWebhookCommand)

	return gs, nil
}

// InitBot initializes the GitLab Telegram bot
func (s *GitLabTelegramService) InitBot() error {
	return s.TelegramService.InitBot([]models.BotCommand{
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

// handleStartCommand handles the /start command
func (s *GitLabTelegramService) handleStartCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	text := "👋 Welcome to GitLab Watch Bot!\n\n" +
		"I can forward GitLab webhook events to this chat.\n\n" +
		"Use /webhook to get your unique webhook URL."

	s.SendMessageOrLogError(update.Message.Chat.ID, text)
}

// handleHelpCommand handles the /help command
func (s *GitLabTelegramService) handleHelpCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	text := "📚 <b>Available Commands</b>\n\n" +
		"• /start - Start the bot\n" +
		"• /help - Show this help message\n" +
		"• /webhook - Get your unique GitLab webhook URL\n\n" +
		"To set up webhooks, use the appropriate command and add the URL to your repository's webhook settings."

	s.SendMessageOrLogError(update.Message.Chat.ID, text)
}

// handleGitLabCommand handles the /gitlab command
func (s *GitLabTelegramService) handleWebhookCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	webhookURL := s.GetChatWebhookURL(update.Message.Chat.ID)

	// Create response message
	text := fmt.Sprintf("🔗 <b>Your GitLab Webhook URL</b>\n\n<code>%s</code>\n\n", webhookURL) +
		"<b>How to set up:</b>\n\n" +
		"1. Go to your GitLab project\n" +
		"2. Click on Settings > Webhooks\n" +
		"3. Paste the URL above in the 'URL' field\n" +
		"4. Select the events you want to receive\n" +
		"5. Click 'Add webhook'\n\n" +
		"You'll receive a confirmation message when the webhook is set up correctly."

	s.SendMessageOrLogError(update.Message.Chat.ID, text)
}

// SendOrUpdatePipelineMessage updates an existing pipeline message or creates a new one
func (s *GitLabTelegramService) SendOrUpdatePipelineMessage(chatID int64, pipelineURL string, text string) error {
	ctx := context.Background()
	pipelineUpdateKey := storage.CreatePipelineUpdateKey(pipelineURL, chatID)

	// Get existing pipeline mapping
	pipeline, err := s.pipelineStorage.GetPipeline(ctx, pipelineUpdateKey)
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
			MessageID:         msg.ID,
		}
		return s.pipelineStorage.SavePipeline(ctx, pipeline)
	} else {
		// Update the existing message
		if err := s.UpdateMessage(chatID, pipeline.MessageID, text); err != nil {
			return err
		}
		// Update the mapping timestamp
		return s.pipelineStorage.SavePipeline(ctx, pipeline)
	}
}
