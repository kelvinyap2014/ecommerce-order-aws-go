resource "aws_lambda_function" "order_api" {
  filename         = "../cmd/api/function.zip"
  function_name    = "${var.project_name}-api-${var.environment}"
  role             = aws_iam_role.lambda_exec.arn
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  source_code_hash = filebase64sha256("../cmd/api/function.zip")

  environment {
    variables = {
      ORDERS_TABLE_NAME = aws_dynamodb_table.orders.name
    }
  }
}

# Verification Lambda 1. Address & Blacklist Check Lambda
resource "aws_lambda_function" "check_address" {
  filename         = "../cmd/verifications/check-address/function.zip"
  function_name    = "${var.project_name}-check-address-${var.environment}"
  role             = aws_iam_role.lambda_exec.arn
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  source_code_hash = filebase64sha256("../cmd/verifications/check-address/function.zip")

  environment {
    variables = {
      ORDERS_TABLE_NAME = aws_dynamodb_table.orders.name
    }
  }
}

# Verification Lambda 2. Credit Check Lambda
resource "aws_lambda_function" "check_credit" {
  filename         = "../cmd/verifications/check-credit/function.zip"
  function_name    = "${var.project_name}-check-credit-${var.environment}"
  role             = aws_iam_role.lambda_exec.arn
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  source_code_hash = filebase64sha256("../cmd/verifications/check-credit/function.zip")

  environment {
    variables = {
      ORDERS_TABLE_NAME = aws_dynamodb_table.orders.name
    }
  }
}

# Verification Lambda 3. Payment Check Lambda
resource "aws_lambda_function" "check_payment" {
  filename         = "../cmd/verifications/check-payment/function.zip"
  function_name    = "${var.project_name}-check-payment-${var.environment}"
  role             = aws_iam_role.lambda_exec.arn
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  source_code_hash = filebase64sha256("../cmd/verifications/check-payment/function.zip")

  environment {
    variables = {
      ORDERS_TABLE_NAME = aws_dynamodb_table.orders.name
    }
  }
}

# Verification Lambda 4. Status Updater Lambda (Interacts with DynamoDB)
resource "aws_lambda_function" "update_status" {
  filename         = "../cmd/verifications/update-status/function.zip"
  function_name    = "${var.project_name}-update-status-${var.environment}"
  role             = aws_iam_role.lambda_exec.arn
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  source_code_hash = filebase64sha256("../cmd/verifications/update-status/function.zip")

  environment {
    variables = {
      ORDERS_TABLE_NAME = aws_dynamodb_table.orders.name
    }
  }
}
