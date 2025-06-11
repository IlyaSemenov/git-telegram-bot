package server

import (
	"fmt"
	"log"
	"net/http"

	"git-telegram-bot/internal/config"
	"git-telegram-bot/internal/handlers"
	"git-telegram-bot/internal/services"

	"github.com/gorilla/mux"
)

type Server struct {
	router      *mux.Router
	telegramSvc *services.TelegramService
}

func New() (*Server, error) {
	// Initialize services
	cryptoSvc := services.NewCryptoService(config.Global.EncryptionKey)

	telegramSvc, err := services.NewTelegramService(config.Global.TelegramBotToken, cryptoSvc, config.Global.BaseURL)
	if err != nil {
		return nil, err
	}

	githubSvc := services.NewGitHubService()
	gitlabSvc := services.NewGitLabService()

	// Initialize handlers
	githubHandler := handlers.NewGitHubHandler(cryptoSvc, telegramSvc, githubSvc)
	gitlabHandler := handlers.NewGitLabHandler(cryptoSvc, telegramSvc, gitlabSvc)
	telegramHandler := handlers.NewTelegramHandler(telegramSvc)

	// Set up router
	router := mux.NewRouter()

	// GitHub webhook endpoint
	router.HandleFunc("/github/{chatID}", githubHandler.HandleWebhook).Methods("POST")

	// GitLab webhook endpoint
	router.HandleFunc("/gitlab/{chatID}", gitlabHandler.HandleWebhook).Methods("POST")

	// Telegram webhook endpoint
	router.HandleFunc("/telegram/webhook", telegramHandler.HandleWebhook).Methods("POST")

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("Failed to write response: %v", err)
		}
	}).Methods("GET")

	return &Server{
		router:      router,
		telegramSvc: telegramSvc,
	}, nil
}

func (s *Server) Router() *mux.Router {
	return s.router
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(":8080", s.router)
}

func (s *Server) SetupTelegramBot() error {
	// Set up webhook
	webhookURL := config.Global.BaseURL + "/telegram/webhook"
	if err := s.telegramSvc.SetWebhook(webhookURL); err != nil {
		return fmt.Errorf("Failed to set up Telegram webhook at %s: %w", webhookURL, err)
	} else {
		log.Printf("Successfully set up Telegram webhook at %s", webhookURL)
	}

	// Set up commands
	if err := s.telegramSvc.SetCommands(); err != nil {
		return fmt.Errorf("Failed to set Telegram commands: %w", err)
	}

	return nil
}
