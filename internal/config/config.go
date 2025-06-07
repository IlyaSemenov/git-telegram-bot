package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/joho/godotenv"
)

// Global configuration instance
var Global *Config

type Config struct {
	TelegramBotToken string
	EncryptionKey    string
	BaseURL          string
	IsLambda         bool
}

func GetLambdaURL() (string, error) {
	ctx := context.Background()

	// Load AWS config (uses Lambda's built-in IAM role)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("Failed to load AWS config: %v", err)
	}

	// Create Lambda client
	client := lambda.NewFromConfig(cfg)

	// Get function URL config
	functionName := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	resp, err := client.GetFunctionUrlConfig(ctx, &lambda.GetFunctionUrlConfigInput{
		FunctionName: &functionName,
	})
	if err != nil {
		return "", fmt.Errorf("Failed to get function URL: %v", err)
	}

	return *resp.FunctionUrl, nil
}

// Initialize loads the configuration and sets the Global variable
func Initialize() error {
	// Check if we're running in Lambda
	isLambda := os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != ""

	// Load .env file if it exists and we're not in Lambda
	if !isLambda {
		if err := godotenv.Load(); err != nil {
			fmt.Printf("Warning: Error loading .env file: %v\n", err)
		}
	}

	// Try to auto-discover BASE_URL if we're in Lambda
	baseURL := os.Getenv("BASE_URL")
	if isLambda && baseURL == "" {
		var err error
		baseURL, err = GetLambdaURL()
		if err != nil {
			return err
		}
		fmt.Printf("Auto-discovered BASE_URL: %s\n", baseURL)
	}

	// Ensure BaseURL doesn't end with a slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	Global = &Config{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		EncryptionKey:    os.Getenv("ENCRYPTION_KEY"),
		BaseURL:          baseURL,
		IsLambda:         isLambda,
	}

	return nil
}
