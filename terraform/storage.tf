# DynamoDB table for storing chat information
resource "aws_dynamodb_table" "chats" {
  name         = "${local.function_name}-chats"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "chat_id"
  range_key    = "bot_type"

  attribute {
    name = "chat_id"
    type = "N" # Number (for Telegram chat_id)
  }

  attribute {
    name = "bot_type"
    type = "S" # String (e.g., "github", "gitlab")
  }

  tags = {
    Name        = "${local.function_name}-chats"
    Environment = terraform.workspace
  }
}

# IAM policy for DynamoDB access
resource "aws_iam_policy" "dynamodb_policy" {
  name        = "${local.function_name}-dynamodb-policy"
  description = "Allow Lambda to access DynamoDB tables"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:DescribeTable",
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem",
          "dynamodb:Query",
          "dynamodb:Scan"
        ]
        Resource = [
          aws_dynamodb_table.chats.arn,
          "${aws_dynamodb_table.chats.arn}/*"
        ]
      }
    ]
  })
}

# Attach DynamoDB policy to Lambda role
resource "aws_iam_role_policy_attachment" "lambda_dynamodb" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.dynamodb_policy.arn
}
