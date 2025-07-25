package gitlab

import (
	"cmp"
	"encoding/json"
	"fmt"
	"html"
	"slices"
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
	MergeRequest *struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	} `json:"merge_request"`
	User struct {
		Name string `json:"name"`
	} `json:"user"`
	Builds []Build `json:"builds"`
}

type Build = struct {
	ID       int     `json:"id"`
	Stage    string  `json:"stage"`
	Name     string  `json:"name"`
	Status   string  `json:"status"`
	Duration float64 `json:"duration"`
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
	case "created":
		emoji = "🛠️"
	case "waiting_for_resource":
		emoji = "🚦"
	case "preparing":
		emoji = "⚙️"
	case "manual":
		emoji = "✋"
	case "scheduled":
		emoji = "📅"
	default:
		emoji = "ℹ️"
	}

	// Replace underscores with spaces in the status
	statusDisplay := strings.ReplaceAll(event.ObjectAttributes.Status, "_", " ")

	// Format branch/MR reference
	var refInfo string
	if event.MergeRequest != nil {
		refInfo = fmt.Sprintf("merge request <a href=\"%s\">%s</a>",
			html.EscapeString(event.MergeRequest.URL),
			html.EscapeString(event.MergeRequest.Title),
		)
	} else {
		refInfo = fmt.Sprintf("branch <code>%s</code>",
			html.EscapeString(event.ObjectAttributes.Ref),
		)
	}

	message.WriteString(fmt.Sprintf(
		"%s Pipeline %s: <a href=\"%s\">%s</a> — <a href=\"%s\">Pipeline #%d</a> (%s)",
		emoji,
		html.EscapeString(statusDisplay),
		event.Project.WebURL,
		html.EscapeString(event.Project.Name),
		event.ObjectAttributes.URL,
		event.ObjectAttributes.ID,
		refInfo,
	))

	// Add build information
	if len(event.Builds) > 0 {
		slices.SortFunc(event.Builds, func(a, b Build) int {
			return cmp.Compare(a.ID, b.ID)
		})
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
			case "canceling":
				buildEmoji = "🛑"
			case "skipped":
				buildEmoji = "⏭️"
			case "created":
				buildEmoji = "🛠️"
			case "manual":
				buildEmoji = "✋"
			case "preparing":
				buildEmoji = "⚙️"
			case "scheduled":
				buildEmoji = "📅"
			case "waiting_for_resource":
				buildEmoji = "🚦"
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
