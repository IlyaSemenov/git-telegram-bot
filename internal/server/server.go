package server

import (
	"fmt"
	"log"
	"net/http"

	"git-telegram-bot/internal/config"
	"git-telegram-bot/internal/handlers"
	"git-telegram-bot/internal/services"
	"git-telegram-bot/internal/services/telegram"

	"github.com/gorilla/mux"
)

type Server struct {
	router *mux.Router
}

func New() (*Server, error) {
	// Initialize GitHub Telegram service
	githubTelegramSvc, err := telegram.NewGitHubTelegramService()
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize GitHub Telegram service: %w", err)
	}

	// Initialize GitLab Telegram service
	gitlabTelegramSvc, err := telegram.NewGitLabTelegramService()
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize GitLab Telegram service: %w", err)
	}

	githubSvc := services.NewGitHubService()
	gitlabSvc := services.NewGitLabService()

	// Initialize handlers
	githubHandler := handlers.NewGitHubHandler(githubTelegramSvc, githubSvc)
	gitlabHandler := handlers.NewGitLabHandler(gitlabTelegramSvc, gitlabSvc)
	githubTelegramHandler := handlers.NewTelegramHandler(githubTelegramSvc)
	gitlabTelegramHandler := handlers.NewTelegramHandler(gitlabTelegramSvc)

	// Set up router
	router := mux.NewRouter()

	// GitHub webhook endpoint
	router.HandleFunc("/github/{chatID}", githubHandler.HandleWebhook).Methods("POST")

	// GitLab webhook endpoint
	router.HandleFunc("/gitlab/{chatID}", gitlabHandler.HandleWebhook).Methods("POST")

	// Telegram webhook endpoints
	router.HandleFunc("/telegram/webhook/github", githubTelegramHandler.HandleWebhook).Methods("POST")
	router.HandleFunc("/telegram/webhook/gitlab", gitlabTelegramHandler.HandleWebhook).Methods("POST")

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("Failed to write response: %v", err)
		}
	}).Methods("GET")

	// Initialize function for Telegram bots
	initBots := func() error {
		if err := githubTelegramSvc.Init(); err != nil {
			return fmt.Errorf("Failed to init GitHub Telegram bot: %v", err)
		}
		if err := gitlabTelegramSvc.Init(); err != nil {
			return fmt.Errorf("Failed to init GitLab Telegram bot: %v", err)
		}
		return nil
	}

	if config.Global.IsLambda {
		router.HandleFunc("/init", func(w http.ResponseWriter, r *http.Request) {
			// Validate secret key from header
			secretKey := r.Header.Get("secret-key")
			if secretKey != config.Global.SecretKey {
				log.Printf("Invalid secret key provided for init endpoint")
				w.WriteHeader(http.StatusUnauthorized)
				if _, writeErr := w.Write([]byte("Unauthorized: Invalid secret key")); writeErr != nil {
					log.Printf("Failed to write response: %v", writeErr)
				}
				return
			}

			if err := initBots(); err != nil {
				log.Printf("%v", err)
				w.WriteHeader(http.StatusInternalServerError)
				if _, writeErr := w.Write([]byte(err.Error())); writeErr != nil {
					log.Printf("Failed to write response: %v", writeErr)
				}
				return
			}

			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("Telegram bots successfully initialized.")); err != nil {
				log.Printf("Failed to write response: %v", err)
			}
		}).Methods("GET")
	} else {
		if err := initBots(); err != nil {
			return nil, err
		}
	}

	return &Server{
		router: router,
	}, nil
}

func (s *Server) Router() *mux.Router {
	return s.router
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(":8080", s.router)
}
