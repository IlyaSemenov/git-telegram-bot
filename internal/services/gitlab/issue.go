package gitlab

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
)

func (s *GitLabService) handleIssueEvent(chatID int64, payload []byte) error {
	var event struct {
		ObjectAttributes struct {
			ID          int    `json:"id"`
			Title       string `json:"title"`
			Description string `json:"description"`
			State       string `json:"state"`
			Action      string `json:"action"`
			URL         string `json:"url"`
		} `json:"object_attributes"`
		Project struct {
			Name              string `json:"name"`
			PathWithNamespace string `json:"path_with_namespace"`
			WebURL            string `json:"web_url"`
		} `json:"project"`
		User struct {
			Name string `json:"name"`
		} `json:"user"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}

	// Only notify on known issue actions
	if event.ObjectAttributes.Action != "open" &&
		event.ObjectAttributes.Action != "close" &&
		event.ObjectAttributes.Action != "reopen" {
		return nil
	}

	// Build message
	var message strings.Builder

	// Add emoji based on action
	var emoji string
	var action string
	switch event.ObjectAttributes.Action {
	case "open":
		emoji = "🆕"
		action = "opened"
	case "close":
		emoji = "✅"
		action = "closed"
	case "reopen":
		emoji = "🔄"
		action = "reopened"
	default:
		emoji = "ℹ️"
		action = event.ObjectAttributes.Action
	}

	message.WriteString(fmt.Sprintf(
		"%s <b>%s</b> %s issue: <a href=\"%s\">%s</a> — <a href=\"%s\">%s</a>.",
		emoji,
		html.EscapeString(event.User.Name),
		action,
		event.Project.WebURL,
		html.EscapeString(event.Project.Name),
		event.ObjectAttributes.URL,
		html.EscapeString(event.ObjectAttributes.Title),
	))

	return s.telegramSvc.SendMessage(chatID, message.String())
}
