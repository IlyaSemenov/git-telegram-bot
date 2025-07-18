package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"git-telegram-bot/internal/config"
	"git-telegram-bot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BaseTelegramService provides common functionality for Telegram bots
type BaseTelegramService struct {
	bot         *tgbotapi.BotAPI
	botId       string
	chatStorage *storage.ChatStorage
}

// NewBaseTelegramService creates a new base Telegram service
func NewBaseTelegramService(botId string, token string, storageInstance *storage.Storage) (*BaseTelegramService, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("Failed to create %s bot API: %w", botId, err)
	}

	return &BaseTelegramService{
		bot:         bot,
		botId:       botId,
		chatStorage: storageInstance.ChatStorage,
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
func (s *BaseTelegramService) SendMessage(chatID int64, text string) error {
	_, err := s.SendMessageWithResult(chatID, text)
	return err
}

// SendMessageWithResult sends a message to a Telegram chat and returns the message info
func (s *BaseTelegramService) SendMessageWithResult(chatID int64, text string) (*tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.DisableWebPagePreview = true

	sentMsg, err := s.bot.Send(msg)

	// Update chat status based on delivery result
	ctx := context.Background()
	chat := &storage.Chat{
		ChatID:  chatID,
		BotType: s.botId,
	}

	if err == nil {
		// Save chat info (this will update existing or create new)
		if err := s.chatStorage.SaveChat(ctx, chat); err != nil {
			log.Printf("Failed to save chat info: %v", err)
		}
	} else if s.isBotBlockedError(err) {
		if err := s.chatStorage.DeleteChat(ctx, chat); err != nil {
			log.Printf("Failed to delete chat info: %v", err)
		}
	}

	return &sentMsg, err
}

// UpdateMessage updates an existing message in a Telegram chat
func (s *BaseTelegramService) UpdateMessage(chatID int64, messageID int, text string) error {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	msg.ParseMode = "HTML"
	msg.DisableWebPagePreview = true

	_, err := s.bot.Send(msg)
	return err
}

// isBotBlockedError checks if the error indicates the bot was blocked or removed
func (s *BaseTelegramService) isBotBlockedError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "bot was blocked") ||
		strings.Contains(errStr, "user is deactivated") ||
		strings.Contains(errStr, "chat not found") ||
		strings.Contains(errStr, "bot is not a member")
}

func (s *BaseTelegramService) GetChatWebhookURL(chatID int64) (string, error) {
	return fmt.Sprintf("%s/%s/%d", config.Global.BaseURL, s.botId, chatID), nil
}
