provider "aws" {
  region = var.aws_region
}

# Provider for ACM certificate (must be in us-east-1 for CloudFront)
provider "aws" {
  alias  = "us_east_1"
  region = "us-east-1"
}
