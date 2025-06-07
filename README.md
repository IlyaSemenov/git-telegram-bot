# GitHub-Telegram Bot

A serverless bot that delivers GitHub webhook events to Telegram groups.

## Features

- Receive GitHub webhook events and forward them to Telegram groups
- Support for multiple GitHub events:
  - Push events (with branch filtering)
  - Workflow run events
- Secure chat ID encryption
- Easy deployment to AWS Lambda
- Custom domain support

## How It Works

1. Add the bot to your Telegram group
2. The bot provides a unique webhook URL for your GitHub repository
3. Add this URL to your GitHub repository's webhook settings
4. GitHub events are now delivered to your Telegram group

## Documentation

- [Deploying to AWS Lambda](docs/deploy-aws-lambda.md)
- [Running Locally](docs/run-local.md)

## Requirements

- Go 1.24 or higher
- AWS account (for Lambda deployment)
- Telegram Bot Token (from [@BotFather](https://t.me/BotFather))

## Quick Start

1. Clone this repository
2. Create a `.env` file based on `.env.example`
3. Run `go mod tidy` to download dependencies
4. Run `make run` to start the bot locally
5. Add the bot to your Telegram group
6. Follow the bot's instructions to set up GitHub webhooks

## License

MIT
