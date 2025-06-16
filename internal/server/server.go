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
	cryptoSvc := services.NewCryptoService(config.Global.SecretKey)

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

			if err := telegramSvc.Init(); err != nil {
				log.Printf("Failed to init Telegram bot: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				if _, writeErr := w.Write(fmt.Appendf(nil, "Error: %v", err)); writeErr != nil {
					log.Printf("Failed to write response: %v", writeErr)
				}
				return
			}
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("Telegram bot successfully initialized.")); err != nil {
				log.Printf("Failed to write response: %v", err)
			}
		}).Methods("GET")
	} else {
		if err := telegramSvc.Init(); err != nil {
			return nil, fmt.Errorf("Failed to init Telegram bot: %v", err)
		}
	}

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
