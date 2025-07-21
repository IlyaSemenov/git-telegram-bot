package telegram

import (
	"fmt"

	"git-telegram-bot/internal/config"
	"git-telegram-bot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type GitHubTelegramService struct {
	*BaseTelegramService
}

func NewGitHubTelegramService(storageInstance *storage.Storage) (*GitHubTelegramService, error) {
	if config.Global.GitHubTelegramBotToken == "" {
		return nil, nil
	}

	base, err := NewBaseTelegramService("github", config.Global.GitHubTelegramBotToken, storageInstance)
	if err != nil {
		return nil, err
	}

	return &GitHubTelegramService{
		BaseTelegramService: base,
	}, nil
}

func (s *GitHubTelegramService) Init() error {
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
			Description: "Get your unique GitHub webhook URL",
		},
	})
}

// ProcessUpdate processes a Telegram update
func (s *GitHubTelegramService) ProcessUpdate(updateJSON []byte) error {
	return s.BaseTelegramService.ProcessUpdate(updateJSON, s.handleUpdate)
}

// handleUpdate handles a Telegram update
func (s *GitHubTelegramService) handleUpdate(update *tgbotapi.Update) error {
	// Handle commands
	if update.Message != nil && update.Message.IsCommand() {
		return s.handleCommand(update.Message)
	}

	return nil
}

// handleCommand handles a Telegram command
func (s *GitHubTelegramService) handleCommand(message *tgbotapi.Message) error {
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
func (s *GitHubTelegramService) handleStartCommand(message *tgbotapi.Message) error {
	text := "👋 Welcome to GitHub Watch Bot!\n\n" +
		"I can forward GitHub webhook events to this chat.\n\n" +
		"Use /webhook to get your unique webhook URL."

	return s.SendMessage(message.Chat.ID, text)
}

// handleHelpCommand handles the /help command
func (s *GitHubTelegramService) handleHelpCommand(message *tgbotapi.Message) error {
	text := "📚 <b>Available Commands</b>\n\n" +
		"• /start - Start the bot\n" +
		"• /help - Show this help message\n" +
		"• /webhook - Get your unique GitHub webhook URL\n\n" +
		"To set up webhooks, use the appropriate command and add the URL to your repository's webhook settings."

	return s.SendMessage(message.Chat.ID, text)
}

// handleGitHubCommand handles the /github command
func (s *GitHubTelegramService) handleWebhookCommand(message *tgbotapi.Message) error {
	webhookURL, err := s.GetChatWebhookURL(message.Chat.ID)
	if err != nil {
		return err
	}

	// Create response message
	text := fmt.Sprintf("🔗 <b>Your GitHub Webhook URL</b>\n\n<code>%s</code>\n\n", webhookURL) +
		"<b>How to set up:</b>\n\n" +
		"1. Go to your GitHub repository\n" +
		"2. Click on Settings > Webhooks > Add webhook\n" +
		"3. Paste the URL above in the 'Payload URL' field\n" +
		"4. Set Content type to 'application/json'\n" +
		"5. Select the events you want to receive\n" +
		"6. Click 'Add webhook'\n\n" +
		"You'll receive a confirmation message when the webhook is set up correctly."

	return s.SendMessage(message.Chat.ID, text)
}
