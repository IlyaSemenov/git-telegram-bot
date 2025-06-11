.PHONY: build clean deploy lint build-lambda update update-env run dev

# Build the application
build:
	go build -o bin/bot ./cmd/bot

# Clean build artifacts
clean:
	rm -rf ./bin

# Build for AWS Lambda
build-lambda:
	GOOS=linux GOARCH=amd64 go build -o bin/bootstrap ./cmd/bot
	cd bin && zip function.zip bootstrap

# Deploy to AWS Lambda
deploy: build-lambda
	aws lambda create-function \
		--function-name git-telegram-bot \
		--runtime provided.al2 \
		--handler bootstrap \
		--zip-file fileb://bin/function.zip \
		--role $(LAMBDA_ROLE_ARN) \
		--environment Variables="{TELEGRAM_BOT_TOKEN=$(TELEGRAM_BOT_TOKEN),ENCRYPTION_KEY=$(ENCRYPTION_KEY)}" \
		--timeout 30 \
		--memory-size 128

# Update existing Lambda function
update: build-lambda
	aws lambda update-function-code \
		--function-name git-telegram-bot \
		--zip-file fileb://bin/function.zip

# Update environment variables
update-env:
	aws lambda update-function-configuration \
		--function-name git-telegram-bot \
		--environment Variables="{TELEGRAM_BOT_TOKEN=$(TELEGRAM_BOT_TOKEN),ENCRYPTION_KEY=$(ENCRYPTION_KEY)}"

# Run linter
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run ./...
	gofmt -s -w .

# Run locally
run:
	go run ./cmd/bot

# Run with hot reload
dev:
	go run github.com/air-verse/air@latest -c .air.toml
