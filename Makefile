.PHONY: build clean deploy lint build-lambda update update-env run dev terraform-init terraform-apply terraform-destroy

# Build the application
build:
	go build -o bin/bot ./cmd/bot

# Clean build artifacts
clean:
	rm -rf ./bin

# Build for AWS Lambda
build-lambda:
	GOOS=linux GOARCH=arm64 go build -o bin/bootstrap ./cmd/bot
	cd bin && zip function.zip bootstrap

# Initialize Terraform
terraform-init:
	cd terraform && terraform init

# Apply Terraform configuration
terraform-apply:
	cd terraform && terraform apply

# Destroy Terraform resources
terraform-destroy:
	cd terraform && terraform destroy

# Deploy to AWS Lambda using Terraform
deploy: build-lambda terraform-apply

# Update existing Lambda function
update: build-lambda
	aws lambda update-function-code \
		--function-name git-telegram-bot \
		--zip-file fileb://bin/function.zip \
		--architectures arm64

# Update environment variables (use after terraform apply)
update-env:
	cd terraform && terraform apply -target=aws_lambda_function.git_telegram_bot

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
