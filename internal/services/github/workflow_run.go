package github

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
)

func (s *GitHubService) handleWorkflowRunEvent(chatID int64, payload []byte) error {
	var event struct {
		Action      string `json:"action"`
		WorkflowRun struct {
			Name       string `json:"name"`
			HTMLURL    string `json:"html_url"`
			Status     string `json:"status"`
			Conclusion string `json:"conclusion"`
		} `json:"workflow_run"`
		Repository struct {
			FullName string `json:"full_name"`
			HTMLURL  string `json:"html_url"`
		} `json:"repository"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}

	// Only notify on completed workflow runs
	if event.Action != "completed" {
		return nil
	}

	// Build message
	var message strings.Builder

	// Add emoji based on conclusion
	var emoji string
	switch event.WorkflowRun.Conclusion {
	case "success":
		emoji = "✅"
	case "failure":
		emoji = "❌"
	case "cancelled":
		emoji = "⚠️"
	default:
		emoji = "ℹ️"
	}

	message.WriteString(fmt.Sprintf(
		"%s Workflow %s: <a href=\"%s\">%s</a> — <a href=\"%s\">%s</a>.",
		emoji,
		html.EscapeString(event.WorkflowRun.Conclusion),
		event.Repository.HTMLURL,
		html.EscapeString(event.Repository.FullName),
		event.WorkflowRun.HTMLURL,
		html.EscapeString(event.WorkflowRun.Name),
	))

	return s.telegramSvc.SendMessage(chatID, message.String())
}
