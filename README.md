# Git-Telegram Bot

A Telegram bot that delivers GitHub and GitLab webhook events to Telegram chats.

It's an alternative to [notifine](https://github.com/mhkafadar/notifine) which is developed very slowly and lacked important features that I needed: notifications on GitHub/GitLab workflow runs, and branch filtering.

## Features

- Receive GitHub and GitLab webhook events and forward them to Telegram chats
- Support for multiple events:
  - Push events (with branch filtering)
  - GitHub workflow run events
  - GitLab pipeline events
  - GitLab merge request events
- Easy deployment to AWS Lambda with Terraform

## How It Works

1. Add [@github_watch_bot](https://t.me/github_watch_bot) or [@gitlab_watch_bot](https://t.me/gitlab_watch_bot) to your Telegram chat
2. Use `/webhook` command to get a unique webhook URL for your repository
3. Add this URL to your GitHub/GitLab repository's webhook settings
4. Events are now delivered to your Telegram chat

## Documentation

- [Running Locally](docs/run-local.md)
- [Deploying to AWS Lambda with Terraform](docs/deploy-aws-lambda.md)

## Requirements

- Go 1.24 or higher
- Telegram Bot Token (from [@BotFather](https://t.me/BotFather))
- AWS account (for Lambda deployment)
- Terraform 1.0+ (for infrastructure deployment)

## Quick Start

### Local Development

1. Clone this repository
2. Create a `.env` file based on `.env.example`
3. Run `go mod tidy` to download dependencies
4. Run `make run` to start the bot locally
5. Add the bot to your Telegram chat
6. Follow the bot's instructions to set up GitHub webhooks

### AWS Deployment

1. Clone this repository
2. Navigate to the `terraform` directory
3. Create a `terraform.tfvars` file with your configuration
4. Run `make terraform-init` and `make terraform-apply`
5. Run `make update` to deploy the updated code to Lambda
6. Add the bot to your Telegram chat

## License

MIT
