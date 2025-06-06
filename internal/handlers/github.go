package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github-telegram-bot/internal/services"
	"github.com/gorilla/mux"
)

type GitHubHandler struct {
	cryptoSvc   *services.CryptoService
	telegramSvc *services.TelegramService
	githubSvc   *services.GitHubService
}

func NewGitHubHandler(cryptoSvc *services.CryptoService, telegramSvc *services.TelegramService, githubSvc *services.GitHubService) *GitHubHandler {
	return &GitHubHandler{
		cryptoSvc:   cryptoSvc,
		telegramSvc: telegramSvc,
		githubSvc:   githubSvc,
	}
}

func (h *GitHubHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Get encrypted chat ID from URL
	vars := mux.Vars(r)
	encryptedChatID := vars["chatID"]

	// Get branch filter from query parameters if present
	branchFilter := r.URL.Query().Get("branch")

	// Decrypt chat ID
	chatID, err := h.cryptoSvc.DecryptChatID(encryptedChatID)
	if err != nil {
		log.Printf("Failed to decrypt chat ID: %v", err)
		http.Error(w, "Invalid chat ID", http.StatusBadRequest)
		return
	}

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
