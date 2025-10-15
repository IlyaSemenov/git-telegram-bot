# Git-Telegram Bot

A Telegram bot that delivers GitHub and GitLab webhook events to Telegram chats.

It's an alternative to [notifine](https://github.com/mhkafadar/notifine) which is developed very slowly and lacked important features that I needed: notifications on GitHub/GitLab workflow runs, and branch filtering.

## Features

- Receive GitHub and GitLab webhook events and forward them to Telegram chats
- Support for multiple events:
  - Push events (with branch filtering)
  - GitHub workflow run events
  - GitLab pipeline events with real-time updates
  - GitLab merge request events

## How It Works

1. Add [@github_watch_bot](https://t.me/github_watch_bot) or [@gitlab_watch_bot](https://t.me/gitlab_watch_bot) to your Telegram chat
2. Use `/webhook` command to get a unique webhook URL for your repository
3. Add this URL to your GitHub/GitLab repository's webhook settings
4. Events are now delivered to your Telegram chat

### Branch Filtering (GitHub)

You can filter GitHub webhook events by branch by adding a `?branch=<branch-name>` query parameter to your webhook URL.

**Example:**

```url
https://webhook.git-watch.mobicom.dev/github/123456789?branch=main
```

This will only send notifications for events that occur on the `main` branch.

## Privacy Policy

This bot is designed with privacy as a core principle. Here’s how data is handled:

**Stored data:**

- **Chat identifiers**:
  - Telegram chat IDs (numeric only) and timestamp of the last handled event
  - Automatically removed if the bot is blocked by the chat
- **GitLab pipeline tracking**:
  - SHA-256 hashes of pipeline identifiers (irreversible, cannot reveal original URLs)
  - Associated Telegram message IDs (for updating status messages)
  - Automatically purged after 24 hours of pipeline inactivity

**Explicitly NOT stored:**

- Repository/pipeline URLs (only hashes)
- Names of users, organizations, or repositories
- Commit messages, code content, or file changes
- Personally identifiable information (PII)
- Data that could identify individuals or organizations

**Data flow**:

1. Webhook events are processed in real-time (never persisted)
2. Pipeline URLs are instantly hashed for status updates
3. Only necessary notification content is forwarded to Telegram
4. No message content remains in the system after delivery

**Retention rules**:

- Most data: Purged immediately after processing
- Pipeline tracking: Purged after 24 hours of inactivity
- All chat data: Removed when the bot is blocked

This is a privacy-focused relay bot that retains only the minimal data required for functionality.

## Developer Documentation

- [Running Locally](docs/run-local.md)
- [Deploying to AWS Lambda with Terraform](docs/deploy-aws-lambda.md)
- [Deploying to Dokku](docs/deploy-dokku.md)

## Requirements

- Go 1.25 or higher
- Telegram Bot Token (from [@BotFather](https://t.me/BotFather))
- AWS account (for Lambda deployment)
- Terraform 1.0+ (for infrastructure deployment, if using AWS Lambda)

## License

MIT
