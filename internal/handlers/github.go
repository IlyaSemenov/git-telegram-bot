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

type GitHubHandler struct {
	telegramSvc *telegram.GitHubTelegramService
	githubSvc   *services.GitHubService
}

func NewGitHubHandler(telegramSvc *telegram.GitHubTelegramService, githubSvc *services.GitHubService) *GitHubHandler {
	return &GitHubHandler{
		telegramSvc: telegramSvc,
		githubSvc:   githubSvc,
	}
}

func (h *GitHubHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Get chat ID from URL
	vars := mux.Vars(r)
	chatID, err := telegram.ParseChatID(vars["chatID"])
	if err != nil {
		http.Error(w, "Failed to parse chatID from URL", http.StatusBadRequest)
		return
	}

	// Get branch filter from query parameters if present
	branchFilter := r.URL.Query().Get("branch")

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Get event type from GitHub headers
	eventType := r.Header.Get("X-GitHub-Event")
	if eventType == "" {
		log.Printf("Missing X-GitHub-Event header")
		http.Error(w, "Missing X-GitHub-Event header", http.StatusBadRequest)
		return
	}

	// Parse GitHub event
	message, err := h.githubSvc.ParseEvent(eventType, body, branchFilter)
	if err != nil {
		log.Printf("Failed to parse GitHub event: %v", err)
		http.Error(w, "Failed to parse GitHub event", http.StatusBadRequest)
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
