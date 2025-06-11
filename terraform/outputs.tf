output "lambda_function_url" {
  description = "URL of the Lambda function"
  value       = aws_lambda_function_url.git_telegram_bot_url.function_url
}

output "cloudwatch_log_group" {
  description = "CloudWatch Log Group for Lambda logs"
  value       = aws_cloudwatch_log_group.lambda_logs.name
}

output "encryption_key" {
  description = "Generated encryption key (sensitive)"
  value       = random_password.encryption_key.result
  sensitive   = true
}

output "github_actions_access_key" {
  description = "AWS Access Key ID for GitHub Actions"
  value       = aws_iam_access_key.github_actions.id
}

output "github_actions_secret_key" {
  description = "AWS Secret Access Key for GitHub Actions (sensitive)"
  value       = aws_iam_access_key.github_actions.secret
  sensitive   = true
}
