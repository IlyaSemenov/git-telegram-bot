package telegram

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"git-telegram-bot/internal/config"
	"git-telegram-bot/internal/services"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BaseTelegramService provides common functionality for Telegram bots
type BaseTelegramService struct {
	bot       *tgbotapi.BotAPI
	botId     string
	cryptoSvc *services.CryptoService
}

// NewBaseTelegramService creates a new base Telegram service
func NewBaseTelegramService(botId string, token string, cryptoSvc *services.CryptoService) (*BaseTelegramService, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("Failed to create %s bot API: %w", botId, err)
	}

	return &BaseTelegramService{
		bot:       bot,
		botId:     botId,
		cryptoSvc: cryptoSvc,
	}, nil
}

// Init initializes a Telegram bot
func (s *BaseTelegramService) Init(commands []tgbotapi.BotCommand) error {
	log.Printf("Init %s bot @%s", s.botId, s.bot.Self.UserName)

	// Set up webhook
	webhookURL := config.Global.BaseURL + "/telegram/webhook/" + s.botId
	if err := s.SetWebhook(webhookURL); err != nil {
		return fmt.Errorf("Failed to set up %s bot webhook at %s: %w", s.botId, webhookURL, err)
	} else {
		log.Printf("Successfully set up %s bot webhook at %s", s.botId, webhookURL)
	}

	// Set up commands
	if err := s.SetCommands(commands); err != nil {
		return fmt.Errorf("Failed to set %s bot commands: %w", s.botId, err)
	}

	return nil
}

// SetWebhook sets up the webhook for the bot
func (s *BaseTelegramService) SetWebhook(webhookURL string) error {
	wh, _ := tgbotapi.NewWebhook(webhookURL)
	_, err := s.bot.Request(wh)
	return err
}

// SetCommands sets the list of available commands for the bot
func (s *BaseTelegramService) SetCommands(commands []tgbotapi.BotCommand) error {
	config := tgbotapi.NewSetMyCommands(commands...)
	_, err := s.bot.Request(config)
	return err
}

// ProcessUpdate processes a Telegram update
func (s *BaseTelegramService) ProcessUpdate(updateJSON []byte, handler func(*tgbotapi.Update) error) error {
	var update tgbotapi.Update
	if err := json.Unmarshal(updateJSON, &update); err != nil {
		return err
	}

	return handler(&update)
}

// SendMessage sends a message to a Telegram chat
func (s *BaseTelegramService) SendMessage(chatID interface{}, text string) error {
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

func (s *BaseTelegramService) GetChatWebhookURL(chatID int64) (string, error) {
	// Encrypt chat ID
	chatIDStr := fmt.Sprintf("%d", chatID)
	encryptedChatID, err := s.cryptoSvc.EncryptChatID(chatIDStr)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s/%s", config.Global.BaseURL, s.botId, url.PathEscape(encryptedChatID)), nil
}
