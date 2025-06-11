variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-east-1"
}

variable "telegram_bot_token" {
  description = "Telegram Bot Token from BotFather"
  type        = string
  sensitive   = true
}
