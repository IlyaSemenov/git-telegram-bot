package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"git-telegram-bot/internal/services"
)

type TelegramHandler struct {
	telegramSvc *services.TelegramService
}

func NewTelegramHandler(telegramSvc *services.TelegramService) *TelegramHandler {
	return &TelegramHandler{
		telegramSvc: telegramSvc,
	}
}

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
