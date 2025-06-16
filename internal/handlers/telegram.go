package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// TelegramHandler handles Telegram webhook requests
type TelegramHandler struct {
	telegramSvc interface {
		ProcessUpdate(updateJSON []byte) error
	}
}

// NewTelegramHandler creates a new Telegram handler
func NewTelegramHandler(telegramSvc interface {
	ProcessUpdate(updateJSON []byte) error
}) *TelegramHandler {
	return &TelegramHandler{
		telegramSvc: telegramSvc,
	}
}

// HandleWebhook handles Telegram webhook requests
func (h *TelegramHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Process update
	if err := h.telegramSvc.ProcessUpdate(body); err != nil {
		log.Printf("Failed to process Telegram update: %v", err)
		http.Error(w, "Failed to process Telegram update", http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
