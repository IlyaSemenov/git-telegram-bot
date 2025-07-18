package gitlab

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
)

func (s *GitLabService) handleMergeRequestEvent(chatID int64, payload []byte) error {
	var event struct {
		ObjectAttributes struct {
			ID           int    `json:"id"`
			Title        string `json:"title"`
			Description  string `json:"description"`
			State        string `json:"state"`
			Action       string `json:"action"`
			URL          string `json:"url"`
			SourceBranch string `json:"source_branch"`
			TargetBranch string `json:"target_branch"`
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

	// Only notify on known MR actions
	if event.ObjectAttributes.Action != "open" &&
		event.ObjectAttributes.Action != "merge" &&
		event.ObjectAttributes.Action != "close" &&
		event.ObjectAttributes.Action != "reopen" &&
		event.ObjectAttributes.Action != "approved" &&
		event.ObjectAttributes.Action != "unapproved" {
		return nil
	}

	// Build message
	var message strings.Builder

	// Add emoji based on action
	var emoji string
	var action string
	switch event.ObjectAttributes.Action {
	case "open":
		emoji = "🔀"
		action = "opened"
	case "merge":
		emoji = "✅"
		action = "merged"
	case "close":
		emoji = "❌"
		action = "closed"
	case "reopen":
		emoji = "🔀"
		action = "reopened"
	case "approved":
		emoji = "✅"
		action = "approved"
	case "unapproved":
		emoji = "❌"
		action = "revoked approval for"
	default:
		emoji = "ℹ️"
		action = event.ObjectAttributes.Action
	}

	message.WriteString(fmt.Sprintf(
		"%s <b>%s</b> %s merge request: <a href=\"%s\">%s</a> — <a href=\"%s\">%s</a> (<code>%s</code> → <code>%s</code>).",
		emoji,
		html.EscapeString(event.User.Name),
		action,
		event.Project.WebURL,
		html.EscapeString(event.Project.Name),
		event.ObjectAttributes.URL,
		html.EscapeString(event.ObjectAttributes.Title),
		html.EscapeString(event.ObjectAttributes.SourceBranch),
		html.EscapeString(event.ObjectAttributes.TargetBranch),
	))

	return s.telegramSvc.SendMessage(chatID, message.String())
}
