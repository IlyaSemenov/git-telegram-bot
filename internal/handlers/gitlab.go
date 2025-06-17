package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"git-telegram-bot/internal/services"
	"git-telegram-bot/internal/services/telegram"

	"github.com/gorilla/mux"
)

type GitLabHandler struct {
	telegramSvc *telegram.GitLabTelegramService
	gitlabSvc   *services.GitLabService
}

func NewGitLabHandler(telegramSvc *telegram.GitLabTelegramService, gitlabSvc *services.GitLabService) *GitLabHandler {
	return &GitLabHandler{
		telegramSvc: telegramSvc,
		gitlabSvc:   gitlabSvc,
	}
}

func (h *GitLabHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Get chat ID from URL
	vars := mux.Vars(r)
	chatID, err := telegram.ParseChatID(vars["chatID"])
	if err != nil {
		http.Error(w, "Failed to parse chatID from URL", http.StatusBadRequest)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Get event type from GitLab headers
	eventType := r.Header.Get("X-Gitlab-Event")
	if eventType == "" {
		log.Printf("Missing X-Gitlab-Event header")
		http.Error(w, "Missing X-Gitlab-Event header", http.StatusBadRequest)
		return
	}

	// Parse GitLab event
	message, err := h.gitlabSvc.ParseEvent(eventType, body)
	if err != nil {
		log.Printf("Failed to parse GitLab event: %v", err)
		http.Error(w, "Failed to parse GitLab event", http.StatusBadRequest)
		return
	}

	// Skip empty messages
	if message == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Send message to Telegram
	if err := h.telegramSvc.SendMessage(chatID, message); err != nil {
		log.Printf("Failed to send message to Telegram: %v", err)
		http.Error(w, "Failed to send message to Telegram", http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
