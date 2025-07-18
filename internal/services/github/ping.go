package github

import (
	"encoding/json"
	"fmt"
	"html"
)

func (s *GitHubService) handlePingEvent(chatID int64, payload []byte) error {
	var event struct {
		Zen        string `json:"zen"`
		HookID     int    `json:"hook_id"`
		Repository struct {
			FullName string `json:"full_name"`
			HTMLURL  string `json:"html_url"`
		} `json:"repository"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}

	message := fmt.Sprintf(
		"✅ GitHub webhook configured successfully for <a href=\"%s\">%s</a>.",
		event.Repository.HTMLURL,
		html.EscapeString(event.Repository.FullName),
	)

	return s.telegramSvc.SendMessage(chatID, message)
}
