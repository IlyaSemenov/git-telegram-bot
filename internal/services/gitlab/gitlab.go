package gitlab

import (
	"fmt"

	telegram "git-telegram-bot/internal/services/telegram/gitlab"
)

type GitLabService struct {
	telegramSvc *telegram.GitLabTelegramService
}

func NewGitLabService(telegramSvc *telegram.GitLabTelegramService) *GitLabService {
	return &GitLabService{
		telegramSvc: telegramSvc,
	}
}

func (s *GitLabService) HandleEvent(chatID int64, eventType string, payload []byte) error {
	switch eventType {
	case "Push Hook":
		return s.handlePushEvent(chatID, payload)
	case "Pipeline Hook":
		return s.handlePipelineEvent(chatID, payload)
	case "Merge Request Hook":
		return s.handleMergeRequestEvent(chatID, payload)
	case "Issue Hook":
		return s.handleIssueEvent(chatID, payload)
	default:
		return fmt.Errorf("unsupported event type: %s", eventType)
	}
}
