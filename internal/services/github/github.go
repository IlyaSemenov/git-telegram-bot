package github

import (
	"fmt"

	telegram "git-telegram-bot/internal/services/telegram/github"
)

type GitHubService struct {
	telegramSvc *telegram.GitHubTelegramService
}

func NewGitHubService(telegramSvc *telegram.GitHubTelegramService) *GitHubService {
	return &GitHubService{
		telegramSvc: telegramSvc,
	}
}

func (s *GitHubService) HandleEvent(chatID int64, eventType string, payload []byte, branchFilter string, includeProject bool) error {
	switch eventType {
	case "ping":
		return s.handlePingEvent(chatID, payload, includeProject)
	case "push":
		return s.handlePushEvent(chatID, payload, branchFilter, includeProject)
	case "workflow_run":
		return s.handleWorkflowRunEvent(chatID, payload, includeProject)
	default:
		return fmt.Errorf("unsupported event type: %s", eventType)
	}
}
