provider "aws" {
  region = var.aws_region
}

# Provider for ACM certificate (must be in us-east-1 for CloudFront)
provider "aws" {
  alias  = "us_east_1"
  region = "us-east-1"
}

locals {
  function_name     = terraform.workspace == "default" ? "git-telegram-bot" : "git-telegram-bot-${terraform.workspace}"
  use_custom_domain = var.custom_domain != ""
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
      # Set BASE_URL to the custom domain if it's being used
      BASE_URL = local.use_custom_domain ? "https://${var.custom_domain}" : null
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

# ACM Certificate for custom domain (only created when using custom domain)
resource "aws_acm_certificate" "cert" {
  count             = local.use_custom_domain ? 1 : 0
  provider          = aws.us_east_1
  domain_name       = var.custom_domain
  validation_method = "DNS"

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_cloudfront_origin_request_policy" "lambda_webhook_headers" {
  name    = "lambda-webhook-headers-policy"
  comment = "Forwards webhook headers to Lambda"

  cookies_config {
    cookie_behavior = "none"
  }

  headers_config {
    header_behavior = "whitelist"
    headers {
      items = [
        "x-gitlab-event",
        "x-github-event",
        "content-type",
        "user-agent",
        "secret-key",
      ]
    }
  }

  query_strings_config {
    query_string_behavior = "all" # Forward all query params (if needed)
  }
}

# CloudFront distribution for custom domain
resource "aws_cloudfront_distribution" "distribution" {
  count = local.use_custom_domain ? 1 : 0

  origin {
    # Use the Lambda function URL as the origin
    domain_name = trimsuffix(trimprefix(aws_lambda_function_url.git_telegram_bot_url.function_url, "https://"), "/")
    origin_id   = "LambdaOrigin"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  enabled             = true
  is_ipv6_enabled     = true
  default_root_object = ""

  aliases = [var.custom_domain]

  default_cache_behavior {
    allowed_methods        = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "LambdaOrigin"
    viewer_protocol_policy = "redirect-to-https"

    cache_policy_id          = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad" # CachingDisabled
    origin_request_policy_id = aws_cloudfront_origin_request_policy.lambda_webhook_headers.id
  }

  price_class = "PriceClass_100"

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn      = aws_acm_certificate.cert[0].arn
    ssl_support_method       = "sni-only"
    minimum_protocol_version = "TLSv1.2_2021"
  }
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
