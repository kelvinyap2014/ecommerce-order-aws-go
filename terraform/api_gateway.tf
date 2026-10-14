# 1. Create the API Gateway HTTP API
resource "aws_apigatewayv2_api" "order_api" {
  name          = "${var.project_name}-api-${var.environment}"
  protocol_type = "HTTP"
  description   = "E-commerce Order Orchestration API (${var.environment})"

  cors_configuration {
    allow_origins = ["*"]
    allow_methods = ["POST", "OPTIONS"]
    allow_headers = ["Content-Type", "Authorization"]
  }
}

# 2. Create a default Stage with auto-deploy enabled for rapid iteration
resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.order_api.id
  name        = "$default"
  auto_deploy = true

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_logs.arn
    format          = jsonencode({
      requestId      = "$context.requestId"
      ip             = "$context.identity.sourceIp"
      requestTime    = "$context.requestTime"
      httpMethod     = "$context.httpMethod"
      routeKey       = "$context.routeKey"
      status         = "$context.status"
      responseLength = "$context.responseLength"
      integrationError = "$context.integrationErrorMessage"
    })
  }
}

# 3. CloudWatch Log Group for API Gateway access logs
resource "aws_cloudwatch_log_group" "api_logs" {
  name              = "/aws/apigateway/${var.project_name}-api-${var.environment}"
  retention_in_days = 7
}

# 4. Integrate API Gateway with the Lambda Function
resource "aws_apigatewayv2_integration" "lambda_integration" {
  api_id                 = aws_apigatewayv2_api.order_api.id
  integration_type       = "AWS_PROXY"
  integration_method     = "POST"
  integration_uri        = aws_lambda_function.order_ingest.invoke_arn
  payload_format_version = "2.0"
}

# 5. Create a POST route for orders
resource "aws_apigatewayv2_route" "post_orders" {
  api_id    = aws_apigatewayv2_api.order_api.id
  route_key = "POST /orders"
  target    = "integrations/${aws_apigatewayv2_integration.lambda_integration.id}"
}

# 6. Allow API Gateway to invoke the Lambda function
resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.order_ingest.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.order_api.execution_arn}/*/*"
}

# 7. Output the public invoke URL of your API
output "api_endpoint" {
  description = "Base URL of the API Gateway HTTP API"
  value       = aws_apigatewayv2_api.order_api.api_endpoint
}
