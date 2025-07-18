package gitlab

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
)

func (s *GitLabService) handlePushEvent(chatID int64, payload []byte) error {
	var event struct {
		Ref      string `json:"ref"`
		Before   string `json:"before"`
		After    string `json:"after"`
		UserName string `json:"user_name"`
		Project  struct {
			Name              string `json:"name"`
			PathWithNamespace string `json:"path_with_namespace"`
			WebURL            string `json:"web_url"`
		} `json:"project"`
		Commits []struct {
			ID        string `json:"id"`
			Message   string `json:"message"`
			Timestamp string `json:"timestamp"`
			URL       string `json:"url"`
			Author    struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"author"`
		} `json:"commits"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}

	// Extract branch name from ref
	branch := strings.TrimPrefix(event.Ref, "refs/heads/")

	// Build message
	var message strings.Builder

	// Check if this is a branch deletion event (after hash is all zeros)
	if event.After == "0000000000000000000000000000000000000000" {
		message.WriteString(fmt.Sprintf(
			"🗑️ <b>%s</b> deleted branch <code>%s</code> from <a href=\"%s\">%s</a>",
			html.EscapeString(event.UserName),
			html.EscapeString(branch),
			event.Project.WebURL,
			html.EscapeString(event.Project.Name),
		))
	} else {
		// Regular push event
		message.WriteString(fmt.Sprintf(
			"🚀 <b>%s</b> pushed to <a href=\"%s\">%s</a> (branch <code>%s</code>)",
			html.EscapeString(event.UserName),
			event.Project.WebURL,
			html.EscapeString(event.Project.Name),
			html.EscapeString(branch),
		))

		// Add commit information
		if len(event.Commits) > 0 {
			message.WriteString(":\n\n")
			for _, commit := range event.Commits {
				message.WriteString(fmt.Sprintf(
					"👉 <b>%s</b>: <a href=\"%s\">%s</a>\n",
					html.EscapeString(commit.Author.Name),
					commit.URL,
					html.EscapeString(strings.TrimSpace(commit.Message)),
				))
			}
		}
	}

	return s.telegramSvc.SendMessage(chatID, message.String())
}
