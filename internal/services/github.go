package services

import (
	"encoding/json"
	"fmt"
	"strings"
)

type GitHubService struct{}

func NewGitHubService() *GitHubService {
	return &GitHubService{}
}

func (s *GitHubService) ParseEvent(eventType string, payload []byte, branchFilter string) (string, error) {
	switch eventType {
	case "ping":
		return s.handlePingEvent(payload)
	case "push":
		return s.handlePushEvent(payload, branchFilter)
	case "workflow_run":
		return s.handleWorkflowRunEvent(payload)
	default:
		return "", fmt.Errorf("unsupported event type: %s", eventType)
	}
}

func (s *GitHubService) handlePingEvent(payload []byte) (string, error) {
	var event struct {
		Zen        string `json:"zen"`
		HookID     int    `json:"hook_id"`
		Repository struct {
			FullName string `json:"full_name"`
			HTMLURL  string `json:"html_url"`
		} `json:"repository"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		return "", err
	}

	return fmt.Sprintf("✅ GitHub webhook configured successfully for [%s](%s).",
		event.Repository.FullName, event.Repository.HTMLURL), nil
}

func (s *GitHubService) handlePushEvent(payload []byte, branchFilter string) (string, error) {
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
		return "", err
	}

	// Extract branch name from ref
	branch := strings.TrimPrefix(event.Ref, "refs/heads/")

	// If branch filter is specified and doesn't match the current branch, skip this event
	if branchFilter != "" && branchFilter != branch {
		return "", nil
	}

	// Build message
	var message strings.Builder

	// Use appropriate verb based on whether it's a force push
	pushVerb := "pushed"
	if event.Forced {
		pushVerb = "force-pushed"
	}

	message.WriteString(fmt.Sprintf(
		"🚀 *%s* %s to [%s](%s) (branch `%s`)",
		event.Pusher.Name, pushVerb, event.Repository.FullName, event.Repository.HTMLURL, branch,
	))

	// Add commit information
	if len(event.Commits) > 0 {
		message.WriteString(":\n\n")
		for _, commit := range event.Commits {
			message.WriteString(fmt.Sprintf("*%s*: [%s](%s)\n",
				commit.Author.Name, commit.Message, commit.URL))
		}
	}

	return message.String(), nil
}

func (s *GitHubService) handleWorkflowRunEvent(payload []byte) (string, error) {
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
		return "", err
	}

	// Only notify on completed workflow runs
	if event.Action != "completed" {
		return "", nil
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

	message.WriteString(fmt.Sprintf("%s Workflow %s: [%s](%s) — [%s](%s).",
		emoji, event.WorkflowRun.Conclusion,
		event.Repository.FullName, event.Repository.HTMLURL,
		event.WorkflowRun.Name, event.WorkflowRun.HTMLURL,
	))

	return message.String(), nil
}
