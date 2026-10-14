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

# 4. Output the public invoke URL of your API
output "api_endpoint" {
  description = "Base URL of the API Gateway HTTP API"
  value       = aws_apigatewayv2_api.order_api.api_endpoint
}
