resource "aws_iam_role" "lambda_exec" {
  name = "${var.project_name}-lambda-exec-${var.environment}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_policy" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "lambda_dynamodb_policy" {
  name = "${var.project_name}-lambda-dynamodb-${var.environment}"
  role = aws_iam_role.lambda_exec.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:PutItem",
          "dynamodb:GetItem",
          "dynamodb:Scan",
          "dynamodb:UpdateItem"
        ]
        Resource = aws_dynamodb_table.orders.arn
      },
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetRecords",
          "dynamodb:GetShardIterator",
          "dynamodb:DescribeStream",
          "dynamodb:ListStreams"
        ]
        Resource = aws_dynamodb_table.orders.stream_arn
      }
    ]
  })
}

# IAM Policy allowing Lambda to invoke Step Functions
resource "aws_iam_role_policy" "lambda_sfn_policy" {
  name = "${var.project_name}-lambda-sfn-${var.environment}"
  role = aws_iam_role.lambda_exec.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "states:StartExecution"
        ]
        Resource = aws_sfn_state_machine.order_workflow.arn
      }
    ]
  })
}

# IAM Role for Step Functions Execution
resource "aws_iam_role" "sfn_exec" {
  name = "${var.project_name}-sfn-exec-${var.environment}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "states.amazonaws.com"
      }
    }]
  })
}

# Policy allowing Step Functions to invoke the verification Lambdas
resource "aws_iam_role_policy" "sfn_lambda_policy" {
  name = "${var.project_name}-sfn-lambda-${var.environment}"
  role = aws_iam_role.sfn_exec.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "lambda:InvokeFunction"
      ]
      Resource = [
        aws_lambda_function.check_address.arn,
        aws_lambda_function.check_credit.arn,
        aws_lambda_function.check_payment.arn,
        aws_lambda_function.update_status.arn
      ]
    }]
  })
}
