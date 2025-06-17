provider "aws" {
  region = var.aws_region
}

locals {
  function_name = terraform.workspace == "default" ? "git-telegram-bot" : "git-telegram-bot-${terraform.workspace}"
}

# Random secret key
resource "random_password" "secret_key" {
  length  = 32
  special = true
}

# Lambda IAM Role
resource "aws_iam_role" "lambda_role" {
  name = "${local.function_name}-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

# Attach basic Lambda execution policy
resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# Custom policy to allow Lambda to discover its own URL
resource "aws_iam_policy" "lambda_url_policy" {
  name        = "${local.function_name}-url-policy"
  description = "Allow Lambda to get its function URL configuration"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "lambda:GetFunctionUrlConfig"
        Resource = "arn:aws:lambda:${var.aws_region}:${data.aws_caller_identity.current.account_id}:function:${local.function_name}"
      }
    ]
  })
}

# Attach URL policy to role
resource "aws_iam_role_policy_attachment" "lambda_url" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.lambda_url_policy.arn
}

# Get current AWS account ID
data "aws_caller_identity" "current" {}

# Lambda function
resource "aws_lambda_function" "git_telegram_bot" {
  function_name = local.function_name
  role          = aws_iam_role.lambda_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  architectures = ["arm64"]

  # Use the actual function zip file
  filename         = "../bin/function.zip"
  source_code_hash = filebase64sha256("../bin/function.zip")

  environment {
    variables = {
      GITHUB_TELEGRAM_BOT_TOKEN = var.github_telegram_bot_token
      GITLAB_TELEGRAM_BOT_TOKEN = var.gitlab_telegram_bot_token
      SECRET_KEY                = random_password.secret_key.result
    }
  }

  depends_on = [
    aws_cloudwatch_log_group.lambda_logs,
  ]
}


# Lambda function URL
resource "aws_lambda_function_url" "git_telegram_bot_url" {
  function_name      = aws_lambda_function.git_telegram_bot.function_name
  authorization_type = "NONE"
  invoke_mode        = "BUFFERED"
}

# Permission for function URL
resource "aws_lambda_permission" "function_url_permission" {
  statement_id  = "FunctionURLAllowPublicAccess"
  action        = "lambda:InvokeFunctionUrl"
  function_name = aws_lambda_function.git_telegram_bot.function_name
  principal     = "*"
  source_arn    = null

  function_url_auth_type = "NONE"
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "lambda_logs" {
  name              = "/aws/lambda/${local.function_name}"
  retention_in_days = 14
}

# IAM User for GitHub Actions (only for production)
resource "aws_iam_user" "github_actions" {
  count = terraform.workspace == "default" ? 1 : 0
  name  = "github-actions-git-telegram-bot"
}

# IAM Policy for GitHub Actions (only for production)
resource "aws_iam_policy" "github_actions_policy" {
  count       = terraform.workspace == "default" ? 1 : 0
  name        = "github-actions-git-telegram-bot-policy"
  description = "Policy for GitHub Actions to update Lambda code"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "lambda:UpdateFunctionCode",
          "lambda:GetFunction",
          "lambda:GetFunctionUrlConfig"
        ]
        Resource = aws_lambda_function.git_telegram_bot.arn
      }
    ]
  })
}

# Attach policy to IAM user (only for production)
resource "aws_iam_user_policy_attachment" "github_actions_policy_attachment" {
  count      = terraform.workspace == "default" ? 1 : 0
  user       = aws_iam_user.github_actions[0].name
  policy_arn = aws_iam_policy.github_actions_policy[0].arn
}

# Access key for GitHub Actions (only for production)
resource "aws_iam_access_key" "github_actions" {
  count = terraform.workspace == "default" ? 1 : 0
  user  = aws_iam_user.github_actions[0].name
}
