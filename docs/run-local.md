# Running Locally

This guide explains how to run the GitHub-Telegram Bot locally for development and testing.

## Prerequisites

1. Go installed on your local machine
2. A Telegram bot token (from [@BotFather](https://t.me/BotFather))
3. A tunneling tool to expose your local server to the internet (Microsoft Dev Tunnels, ngrok, etc.)

## Step 1: Install Dependencies

Before running the bot, you need to download all required dependencies:

```bash
go mod tidy
```

This command will download all dependencies listed in the `go.mod` file and create/update the `go.sum` file with the correct checksums. This step is essential when you first clone the repository or when dependencies change.

## Step 2: Set Up a Tunnel

Since the bot operates exclusively in webhook mode, you need to expose your local server to the internet.

### Using ngrok

1. Install ngrok if you haven't already:

```bash
brew install ngrok
```

2. Start ngrok:

```bash
ngrok http 8080
```

3. Note the HTTPS URL provided by ngrok (e.g., `https://a1b2c3d4.ngrok.io`)

## Step 3: Set Up Environment Variables

Create a `.env` file in the root directory of the project based on the `.env.example` file:

```
TELEGRAM_BOT_TOKEN=your_telegram_bot_token
ENCRYPTION_KEY=your_random_encryption_key
BASE_URL=https://your-tunnel-url  # Use the URL from your tunnel
```

To generate a random encryption key:

```bash
openssl rand -base64 32
```

## Step 4: Run the Bot

Start the bot in local mode with hot reload:

```bash
make dev
```

This will start the bot with hot reload enabled, automatically restarting the server when you make code changes.

Alternatively, you can run without hot reload:

```bash
make run
```

The bot will start and automatically set up a webhook with Telegram using your tunnel URL.

If you encounter errors about missing dependencies or `go.sum` entries, run `go mod tidy` again to resolve them.

## Step 5: Test the Bot

1. Add your bot to a Telegram group or send a direct message to it
2. The bot will provide a unique webhook URL for GitHub
3. Add this URL to your GitHub repository's webhook settings for testing

## Debugging

The bot logs information to the console, which can help with debugging.

If you're having issues with GitHub webhooks, check:

1. The tunnel console for incoming requests
2. The bot's console output for any errors
3. GitHub's webhook delivery page for request/response details

## Local Development Tips

- Changes to the code require restarting the bot
- You can use `go run ./cmd/bot` with the Go debugger for step-by-step debugging
- Keep your tunnel running in a separate terminal window while developing
