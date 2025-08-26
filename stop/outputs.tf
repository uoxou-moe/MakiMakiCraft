output "lambda_function_arn" {
    description = "The ARN of the SsmCommandExecutor Lambda function"
    value       = aws_lambda_function.base_lambda.arn
}
