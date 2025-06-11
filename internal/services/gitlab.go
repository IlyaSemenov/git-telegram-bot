package services

import (
	"encoding/json"
	"fmt"
	"strings"
)

type GitLabService struct{}

func NewGitLabService() *GitLabService {
	return &GitLabService{}
}

func (s *GitLabService) ParseEvent(eventType string, payload []byte) (string, error) {
	switch eventType {
	case "Push Hook":
		return s.handlePushEvent(payload)
	case "Pipeline Hook":
		return s.handlePipelineEvent(payload)
	case "Merge Request Hook":
		return s.handleMergeRequestEvent(payload)
	default:
		return "", fmt.Errorf("unsupported event type: %s", eventType)
	}
}

func (s *GitLabService) handlePushEvent(payload []byte) (string, error) {
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
		return "", err
	}

	// Extract branch name from ref
	branch := strings.TrimPrefix(event.Ref, "refs/heads/")

	// Build message
	var message strings.Builder

	message.WriteString(fmt.Sprintf(
		"🚀 *%s* pushed to [%s](%s) (branch `%s`)",
		event.UserName, event.Project.Name, event.Project.WebURL, branch,
	))

	// Add commit information
	if len(event.Commits) > 0 {
		message.WriteString(":\n\n")
		for _, commit := range event.Commits {
			message.WriteString(fmt.Sprintf("✅ *%s*: [%s](%s)\n",
				commit.Author.Name, strings.TrimSpace(commit.Message), commit.URL))
		}
	}

	return message.String(), nil
}

func (s *GitLabService) handlePipelineEvent(payload []byte) (string, error) {
	var event struct {
		ObjectAttributes struct {
			ID             int    `json:"id"`
			Ref            string `json:"ref"`
			Status         string `json:"status"`
			DetailedStatus string `json:"detailed_status"`
			WebURL         string `json:"web_url"`
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
			WebURL   string  `json:"web_url"`
		} `json:"builds"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		return "", err
	}

	// Only notify on completed pipelines
	if event.ObjectAttributes.Status != "success" &&
		event.ObjectAttributes.Status != "failed" &&
		event.ObjectAttributes.Status != "canceled" {
		return "", nil
	}

	// Build message
	var message strings.Builder

	// Add emoji based on status
	var emoji string
	switch event.ObjectAttributes.Status {
	case "success":
		emoji = "✅"
	case "failed":
		emoji = "❌"
	case "canceled":
		emoji = "⚠️"
	default:
		emoji = "ℹ️"
	}

	message.WriteString(fmt.Sprintf("%s Pipeline %s: [%s](%s) — [Pipeline #%d](%s) (branch `%s`)",
		emoji, event.ObjectAttributes.Status,
		event.Project.Name, event.Project.WebURL,
		event.ObjectAttributes.ID, event.ObjectAttributes.WebURL,
		event.ObjectAttributes.Ref,
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
			} else {
				durationStr = fmt.Sprintf("%.1f seconds", build.Duration)
			}

			message.WriteString(fmt.Sprintf("%s *%s* (%s)\n",
				buildEmoji, build.Name, durationStr))
		}
	}

	return message.String(), nil
}

func (s *GitLabService) handleMergeRequestEvent(payload []byte) (string, error) {
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
		return "", err
	}

	// Only notify on known MR actions
	if event.ObjectAttributes.Action != "open" &&
		event.ObjectAttributes.Action != "merge" &&
		event.ObjectAttributes.Action != "close" &&
		event.ObjectAttributes.Action != "reopen" &&
		event.ObjectAttributes.Action != "approved" &&
		event.ObjectAttributes.Action != "unapproved" {
		return "", nil
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

	message.WriteString(fmt.Sprintf("%s *%s* %s merge request: [%s](%s) — [%s](%s) (`%s` → `%s`).",
		emoji, event.User.Name, action,
		event.Project.Name, event.Project.WebURL,
		event.ObjectAttributes.Title, event.ObjectAttributes.URL,
		event.ObjectAttributes.SourceBranch, event.ObjectAttributes.TargetBranch,
	))

	return message.String(), nil
}
