.PHONY: build clean deploy lint build-lambda update update-env run dev terraform-init terraform-apply terraform-destroy

# Build the application
build:
	go build -o bin/bot ./cmd/bot

# Clean build artifacts
clean:
	rm -rf ./bin

# Build for AWS Lambda
build-lambda:
	# For CGO, see https://github.com/aws/aws-lambda-go/issues/340
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bin/bootstrap ./cmd/bot
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
	# Get the Lambda function URL
	LAMBDA_URL=$$(aws lambda get-function-url-config --function-name git-telegram-bot --query "FunctionUrl" --output text) && \
	LAMBDA_INIT_URL=$${LAMBDA_URL}init && \
	echo "Initializing bot at $$LAMBDA_INIT_URL" && \
	curl -s "$$LAMBDA_INIT_URL"

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
