package gitlab

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
)

// PipelineEventData represents the parsed pipeline event data
type PipelineEventData struct {
	ObjectAttributes struct {
		ID             int    `json:"id"`
		Ref            string `json:"ref"`
		Status         string `json:"status"`
		DetailedStatus string `json:"detailed_status"`
		URL            string `json:"url"`
		Duration       int    `json:"duration"`
	} `json:"object_attributes"`
	Project struct {
		Name              string `json:"name"`
		PathWithNamespace string `json:"path_with_namespace"`
		WebURL            string `json:"web_url"`
	} `json:"project"`
	User struct {
		Name string `json:"name"`
	} `json:"user"`
	Builds []struct {
		ID       int     `json:"id"`
		Stage    string  `json:"stage"`
		Name     string  `json:"name"`
		Status   string  `json:"status"`
		Duration float64 `json:"duration"`
	} `json:"builds"`
}

func (s *GitLabService) handlePipelineEvent(chatID int64, payload []byte) error {
	var event PipelineEventData

	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}

	var message strings.Builder

	// Add emoji based on status
	var emoji string
	switch event.ObjectAttributes.Status {
	case "success":
		emoji = "✅"
	case "failed":
		emoji = "❌"
	case "running":
		emoji = "🔄"
	case "pending":
		emoji = "⏳"
	case "canceled":
		emoji = "⚠️"
	case "skipped":
		emoji = "⏭️"
	default:
		emoji = "ℹ️"
	}

	message.WriteString(fmt.Sprintf(
		"%s Pipeline %s: <a href=\"%s\">%s</a> — <a href=\"%s\">Pipeline #%d</a> (branch <code>%s</code>)",
		emoji,
		html.EscapeString(event.ObjectAttributes.Status),
		event.Project.WebURL,
		html.EscapeString(event.Project.Name),
		event.ObjectAttributes.URL,
		event.ObjectAttributes.ID,
		html.EscapeString(event.ObjectAttributes.Ref),
	))

	// Add build information
	if len(event.Builds) > 0 {
		message.WriteString(":\n\n")
		for _, build := range event.Builds {
			// Add emoji based on build status
			var buildEmoji string
			switch build.Status {
			case "success":
				buildEmoji = "✅"
			case "failed":
				buildEmoji = "❌"
			case "running":
				buildEmoji = "🔄"
			case "pending":
				buildEmoji = "⏳"
			case "canceled":
				buildEmoji = "⚠️"
			case "skipped":
				buildEmoji = "⏭️"
			default:
				buildEmoji = "ℹ️"
			}

			// Format duration as string
			var durationStr string
			if build.Duration >= 1.0 {
				durationStr = fmt.Sprintf("%.0f seconds", build.Duration)
			} else if build.Duration > 0 {
				durationStr = fmt.Sprintf("%.1f seconds", build.Duration)
			} else {
				durationStr = ""
			}

			if durationStr != "" {
				message.WriteString(fmt.Sprintf(
					"%s <b>%s</b> (%s)\n",
					buildEmoji,
					html.EscapeString(build.Name),
					durationStr,
				))
			} else {
				message.WriteString(fmt.Sprintf(
					"%s <b>%s</b>\n",
					buildEmoji,
					html.EscapeString(build.Name),
				))
			}
		}
	}

	// Get pipeline URL for updating existing messages
	pipelineURL := event.ObjectAttributes.URL

	// Try to update existing message or create new one
	return s.telegramSvc.SendOrUpdatePipelineMessage(chatID, pipelineURL, message.String())
}
