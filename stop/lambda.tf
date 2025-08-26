# lambda関数用のIAMロール
resource "aws_iam_role" "minimal_lambda_role" {
  name = "terraform-minimal-lambda-role"
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

# lambda関数用のIAM Policy.
# Include: Logging, SSM
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

# lambda関数本体
resource "aws_lambda_function" "base_lambda" {
  filename         = "deployment_package.zip"
  function_name    = "SsmCommandExecutor"
  role             = aws_iam_role.minimal_lambda_role.arn
  handler          = "index.handler"

  # ハッシュによってファイルの変更を検知
  source_code_hash = filebase64sha256("deployment_package.zip")
  runtime          = "nodejs22.x"

  tags = {
    ManagedBy = "Terraform",
    Project   = "MinecraftServerAutomation"
  }
}
