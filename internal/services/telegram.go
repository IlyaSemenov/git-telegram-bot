package services

import (
	"encoding/json"
	"fmt"
	"net/url"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramService struct {
	bot       *tgbotapi.BotAPI
	cryptoSvc *CryptoService
	baseURL   string
}

func NewTelegramService(token string, cryptoSvc *CryptoService, baseURL string) (*TelegramService, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &TelegramService{
		bot:       bot,
		cryptoSvc: cryptoSvc,
		baseURL:   baseURL,
	}, nil
}

func (s *TelegramService) SetWebhook(webhookURL string) error {
	wh, _ := tgbotapi.NewWebhook(webhookURL)
	_, err := s.bot.Request(wh)
	return err
}

// SetCommands sets the list of available commands for the bot
func (s *TelegramService) SetCommands() error {
	commands := []tgbotapi.BotCommand{
		{
			Command:     "start",
			Description: "Start the bot",
		},
		{
			Command:     "help",
			Description: "Show help information",
		},
		{
			Command:     "github",
			Description: "Get your unique GitHub webhook URL",
		},
		{
			Command:     "gitlab",
			Description: "Get your unique GitLab webhook URL",
		},
	}

	config := tgbotapi.NewSetMyCommands(commands...)
	_, err := s.bot.Request(config)
	return err
}

func (s *TelegramService) ProcessUpdate(updateJSON []byte) error {
	var update tgbotapi.Update
	if err := json.Unmarshal(updateJSON, &update); err != nil {
		return err
	}

	return s.handleUpdate(&update)
}

func (s *TelegramService) handleUpdate(update *tgbotapi.Update) error {
	// Handle commands
	if update.Message != nil && update.Message.IsCommand() {
		return s.handleCommand(update.Message)
	}

	return nil
}

func (s *TelegramService) handleCommand(message *tgbotapi.Message) error {
	switch message.Command() {
	case "start":
		return s.handleStartCommand(message)
	case "help":
		return s.handleHelpCommand(message)
	case "github":
		return s.handleGitHubCommand(message)
	case "gitlab":
		return s.handleGitLabCommand(message)
	default:
		return s.SendMessage(message.Chat.ID, "Unknown command. Type /help for available commands.")
	}
}

func (s *TelegramService) handleStartCommand(message *tgbotapi.Message) error {
	text := "👋 Welcome to Git Pusher Bot!\n\n" +
		"I can forward GitHub and GitLab webhook events to this chat.\n\n" +
		"Use /github or /gitlab to get your unique webhook URLs."

	return s.SendMessage(message.Chat.ID, text)
}

func (s *TelegramService) handleHelpCommand(message *tgbotapi.Message) error {
	text := "📚 <b>Available Commands</b>\n\n" +
		"• /start - Start the bot\n" +
		"• /help - Show this help message\n" +
		"• /github - Get your unique GitHub webhook URL\n" +
		"• /gitlab - Get your unique GitLab webhook URL\n\n" +
		"To set up webhooks, use the appropriate command and add the URL to your repository's webhook settings."

	return s.SendMessage(message.Chat.ID, text)
}

func (s *TelegramService) handleGitHubCommand(message *tgbotapi.Message) error {
	// Encrypt chat ID
	chatIDStr := fmt.Sprintf("%d", message.Chat.ID)
	encryptedChatID, err := s.cryptoSvc.EncryptChatID(chatIDStr)
	if err != nil {
		return err
	}

	// Create GitHub webhook URL
	webhookURL := fmt.Sprintf("%s/github/%s", s.baseURL, url.PathEscape(encryptedChatID))

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

func (s *TelegramService) handleGitLabCommand(message *tgbotapi.Message) error {
	// Encrypt chat ID
	chatIDStr := fmt.Sprintf("%d", message.Chat.ID)
	encryptedChatID, err := s.cryptoSvc.EncryptChatID(chatIDStr)
	if err != nil {
		return err
	}

	// Create GitLab webhook URL
	webhookURL := fmt.Sprintf("%s/gitlab/%s", s.baseURL, url.PathEscape(encryptedChatID))

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

func (s *TelegramService) SendMessage(chatID interface{}, text string) error {
	var chatIDInt64 int64

	switch v := chatID.(type) {
	case int64:
		chatIDInt64 = v
	case string:
		// Try to parse string as int64
		if _, err := fmt.Sscanf(v, "%d", &chatIDInt64); err != nil {
			return fmt.Errorf("failed to parse chat ID: %w", err)
		}
	default:
		return fmt.Errorf("invalid chat ID type: %T", chatID)
	}

	msg := tgbotapi.NewMessage(chatIDInt64, text)
	msg.ParseMode = "HTML"
	msg.DisableWebPagePreview = true

	_, err := s.bot.Send(msg)
	return err
}
