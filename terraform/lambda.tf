resource "aws_lambda_function" "order_ingest" {
  filename         = "../cmd/api/function.zip"
  function_name    = "${var.project_name}-ingest-${var.environment}"
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
