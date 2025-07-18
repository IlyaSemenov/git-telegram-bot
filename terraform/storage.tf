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

# DynamoDB table for storing pipeline information
resource "aws_dynamodb_table" "pipelines" {
  name         = "${local.function_name}-pipelines"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "pipeline_update_key"

  attribute {
    name = "pipeline_update_key"
    type = "S" # String (hash of pipeline URL + chat ID)
  }

  ttl {
    attribute_name = "expires_at"
    enabled        = true
  }

  tags = {
    Name        = "${local.function_name}-pipelines"
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
          "dynamodb:BatchGetItem", // docstore .Get() needs this even for key access
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem",
          "dynamodb:Query",
          "dynamodb:Scan"
        ]
        Resource = [
          aws_dynamodb_table.chats.arn,
          "${aws_dynamodb_table.chats.arn}/*",
          aws_dynamodb_table.pipelines.arn,
          "${aws_dynamodb_table.pipelines.arn}/*"
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
