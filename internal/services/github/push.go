package github

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
)

func (s *GitHubService) handlePushEvent(chatID int64, payload []byte, branchFilter string) error {
	var event struct {
		Ref        string `json:"ref"`
		Before     string `json:"before"`
		After      string `json:"after"`
		Repository struct {
			FullName string `json:"full_name"`
			HTMLURL  string `json:"html_url"`
		} `json:"repository"`
		Pusher struct {
			Name string `json:"name"`
		} `json:"pusher"`
		Forced  bool `json:"forced"`
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

	// If branch filter is specified and doesn't match the current branch, skip this event
	if branchFilter != "" && branchFilter != branch {
		return nil
	}

	// Build message
	var message strings.Builder

	// Check if this is a branch deletion event (after hash is all zeros)
	if event.After == "0000000000000000000000000000000000000000" {
		message.WriteString(fmt.Sprintf(
			"🗑️ <b>%s</b> deleted branch <code>%s</code> from <a href=\"%s\">%s</a>",
			html.EscapeString(event.Pusher.Name),
			html.EscapeString(branch),
			event.Repository.HTMLURL,
			html.EscapeString(event.Repository.FullName),
		))
	} else {
		// Use appropriate verb based on whether it's a force push
		pushVerb := "pushed"
		if event.Forced {
			pushVerb = "force-pushed"
		}

		message.WriteString(fmt.Sprintf(
			"🚀 <b>%s</b> %s to <a href=\"%s\">%s</a> (branch <code>%s</code>)",
			html.EscapeString(event.Pusher.Name),
			pushVerb,
			event.Repository.HTMLURL,
			html.EscapeString(event.Repository.FullName),
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
