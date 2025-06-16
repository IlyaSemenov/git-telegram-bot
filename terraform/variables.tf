variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-east-1"
}

variable "github_telegram_bot_token" {
  description = "GitHub Telegram Bot Token from BotFather"
  type        = string
  sensitive   = true
}

variable "gitlab_telegram_bot_token" {
  description = "GitLab Telegram Bot Token from BotFather"
  type        = string
  sensitive   = true
}
