# Deploying to AWS Lambda

This guide explains how to deploy the GitHub-Telegram Bot to AWS Lambda using the AWS CLI.

## Prerequisites

1. AWS CLI installed and configured with appropriate credentials
2. A Telegram bot token (from [@BotFather](https://t.me/BotFather))
3. A custom domain name (optional)
4. Go installed on your local machine

## Step 1: Set Environment Variables

First, set the required environment variables:

```bash
# Set these values according to your setup
export TELEGRAM_BOT_TOKEN="your_telegram_bot_token"
export CUSTOM_DOMAIN="your-custom-domain.com"

# Generate a random encryption key
export ENCRYPTION_KEY=$(openssl rand -base64 32)
export BASE_URL="https://$CUSTOM_DOMAIN"

echo "Generated encryption key: $ENCRYPTION_KEY"
```

## Step 2: Create IAM Role

Create an IAM role that allows Lambda to access other AWS services:

```bash
LAMBDA_ROLE_ARN=$(aws iam create-role \
  --role-name github-telegram-bot-role \
  --assume-role-policy-document '{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Principal": {
          "Service": "lambda.amazonaws.com"
        },
        "Action": "sts:AssumeRole"
      }
    ]
  }' \
  --query 'Role.Arn' \
  --output text)

echo "Lambda Role ARN: $LAMBDA_ROLE_ARN"

# Attach the necessary policies
aws iam attach-role-policy \
  --role-name github-telegram-bot-role \
  --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole

# Create a custom policy to allow the Lambda function to discover its own URL
aws iam create-policy \
  --policy-name github-telegram-bot-url-policy \
  --policy-document '{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Action": "lambda:GetFunctionUrlConfig",
        "Resource": "arn:aws:lambda:*:*:function:github-telegram-bot"
      }
    ]
  }'

# Attach the custom policy to the role
aws iam attach-role-policy \
  --role-name github-telegram-bot-role \
  --policy-arn $(aws iam list-policies --query 'Policies[?PolicyName==`github-telegram-bot-url-policy`].Arn' --output text)
```

## Step 3: Build and Deploy the Lambda Function

First, ensure all dependencies are properly downloaded and the `go.sum` file is up to date:

```bash
go mod tidy
```

Build and deploy the function:

```bash
make deploy
```

## Step 4: Create a Function URL

Create a function URL to expose your Lambda function:

```bash
aws lambda create-function-url-config \
  --function-name github-telegram-bot \
  --auth-type NONE \
  --invoke-mode BUFFERED

# Add permission to allow public access to the function URL
aws lambda add-permission \
  --function-name github-telegram-bot \
  --action lambda:InvokeFunctionUrl \
  --principal "*" \
  --function-url-auth-type "NONE" \
  --statement-id "FunctionURLAllowPublicAccess"
```

Note the function URL from the output.

## Step 5: Test the Bot

1. Add your bot to a Telegram group
2. The bot will provide a webhook URL
3. Add this URL to your GitHub repository's webhook settings

## Updating the Bot

To update the bot after making changes:

```bash
make update
```
