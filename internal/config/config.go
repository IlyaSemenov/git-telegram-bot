package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/joho/godotenv"
)

// Global configuration instance
var Global *Config

type Config struct {
	GitHubTelegramBotToken      string
	GitLabTelegramBotToken      string
	SecretKey                   string
	BaseURL                     string
	IsLambda                    bool
	StorageConnectionStringBase string
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
	var lambdaFunctionName = os.Getenv("AWS_LAMBDA_FUNCTION_NAME")

	// Check if we're running in Lambda
	isLambda := lambdaFunctionName != ""

	// Load .env file if it exists and we're not in Lambda
	if !isLambda {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: Error loading .env file: %v", err)
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

	githubTelegramBotToken := os.Getenv("GITHUB_TELEGRAM_BOT_TOKEN")
	gitlabTelegramBotToken := os.Getenv("GITLAB_TELEGRAM_BOT_TOKEN")

	if githubTelegramBotToken == "" && gitlabTelegramBotToken == "" {
		return fmt.Errorf("You must provide at least one of the environment variables: GITHUB_TELEGRAM_BOT_TOKEN or GITLAB_TELEGRAM_BOT_TOKEN.")
	}

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		return fmt.Errorf("SECRET_KEY environment variable is missing")
	}

	// Auto-detect storage connection string base
	var storageConnectionStringBase string
	if isLambda {
		// Running in AWS - use DynamoDB
		storageConnectionStringBase = fmt.Sprintf("dynamodb://%s", lambdaFunctionName)
	}

	Global = &Config{
		GitHubTelegramBotToken:      githubTelegramBotToken,
		GitLabTelegramBotToken:      gitlabTelegramBotToken,
		SecretKey:                   secretKey,
		BaseURL:                     baseURL,
		IsLambda:                    isLambda,
		StorageConnectionStringBase: storageConnectionStringBase,
	}

	return nil
}
