package gitlab

import (
	"fmt"

	"git-telegram-bot/internal/services/telegram"
)

type GitLabService struct {
	telegramSvc telegram.PipelineMessageSender
}

func NewGitLabService(telegramSvc telegram.PipelineMessageSender) *GitLabService {
	return &GitLabService{
		telegramSvc: telegramSvc,
	}
}

func (s *GitLabService) HandleEvent(chatID int64, eventType string, payload []byte, includeProject bool) error {
	switch eventType {
	case "Push Hook":
		return s.handlePushEvent(chatID, payload, includeProject)
	case "Pipeline Hook":
		return s.handlePipelineEvent(chatID, payload, includeProject)
	case "Merge Request Hook":
		return s.handleMergeRequestEvent(chatID, payload, includeProject)
	case "Issue Hook":
		return s.handleIssueEvent(chatID, payload, includeProject)
	default:
		return fmt.Errorf("unsupported event type: %s", eventType)
	}
}
