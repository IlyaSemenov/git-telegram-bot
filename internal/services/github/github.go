package github

import (
	"fmt"

	"git-telegram-bot/internal/services/telegram"
)

type GitHubService struct {
	telegramSvc *telegram.GitHubTelegramService
}

func NewGitHubService(telegramSvc *telegram.GitHubTelegramService) *GitHubService {
	return &GitHubService{
		telegramSvc: telegramSvc,
	}
}

func (s *GitHubService) HandleEvent(chatID int64, eventType string, payload []byte, branchFilter string) error {
	switch eventType {
	case "ping":
		return s.handlePingEvent(chatID, payload)
	case "push":
		return s.handlePushEvent(chatID, payload, branchFilter)
	case "workflow_run":
		return s.handleWorkflowRunEvent(chatID, payload)
	default:
		return fmt.Errorf("unsupported event type: %s", eventType)
	}
}
