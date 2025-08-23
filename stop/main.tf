# 1. TerraformとAWSプロバイダーの設定
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "ap-northeast-1"
  profile = "terraform-sso-profile"
}

# 2. Lambda関数用のIAMロールとポリシーを定義
resource "aws_iam_role" "lambda_exec_role" {
  name = "terraform-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole",
        Effect = "Allow",
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_policy" "lambda_logging_policy" {
  name        = "terraform-lambda-logging"
  description = "IAM policy for logging from a Lambda function"

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ],
        Effect   = "Allow",
        Resource = "arn:aws:logs:*:*:*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_logs_attachment" {
  role       = aws_iam_role.lambda_exec_role.name
  policy_arn = aws_iam_policy.lambda_logging_policy.arn
}

# 3. Lambda関数を定義
resource "aws_lambda_function" "hello_world_lambda" {
  filename      = "deployment_package.zip"
  function_name = "TerraformHelloWorldLambda"

  role          = aws_iam_role.lambda_exec_role.arn
  handler       = "index.handler"
  source_code_hash = filebase64sha256("deployment_package.zip")
  runtime = "nodejs22.x"

  tags = {
    ManagedBy = "Terraform"
    Language  = "Node.js"
  }
}

# 4. (任意) 作成されたリソースの情報を出力
output "lambda_function_arn" {
  description = "The ARN of the Lambda function"
  value       = aws_lambda_function.hello_world_lambda.arn
}
