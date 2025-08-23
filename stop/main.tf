terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region  = "ap-northeast-1"
  profile = "terraform-sso-profile"
}

# CloudWatch Logs書き込み権限のみを持つIAMロール
resource "aws_iam_role" "minimal_lambda_role" {
  # terraform-* プレフィックスを忘れずに！
  name = "terraform-minimal-lambda-role"

  # Lambdaサービスがこのロールを引き受けるためのお決まりの設定
  assume_role_policy = jsonencode({
    Version   = "2012-10-17",
    Statement = [
      {
        Action    = "sts:AssumeRole",
        Effect    = "Allow",
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_policy" "lambda_ssm_policy" {
  name   = "terraform-lambda-ssm-policy"
  policy = jsonencode({
    Version   = "2012-10-17",
    Statement = [
      {
        Sid = "AllowLogging",
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ],
        Effect   = "Allow",
        Resource = "arn:aws:logs:*:*:*"
      },
      {
        Sid = "AllowSsmSendCommand",
        Action = [
          "ssm:SendCommand",
          "ssm:GetCommandInvocation"
        ],
        Effect = "Allow",
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_policy_attachment" {
  role       = aws_iam_role.minimal_lambda_role.name
  policy_arn = aws_iam_policy.lambda_ssm_policy.arn
}

resource "aws_lambda_function" "base_lambda" {
  filename         = "deployment_package.zip"
  function_name    = "SsmCommandExecutor"
  role             = aws_iam_role.minimal_lambda_role.arn
  handler          = "index.handler"

  # コードのZIPファイルが変更されたことをTerraformに伝えるために重要
  source_code_hash = filebase64sha256("deployment_package.zip")
  runtime          = "nodejs22.x"

  tags = {
    ManagedBy = "Terraform",
    Project   = "MinecraftServerAutomation"
  }
}