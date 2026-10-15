resource "aws_dynamodb_table" "orders" {
  name         = "${var.project_name}-orders-${var.environment}"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "OrderId"

  attribute {
    name = "OrderId"
    type = "S"
  }

  attribute {
    name = "CustomerId"
    type = "S"
  }

  attribute {
    name = "CreatedAt"
    type = "S"
  }

  # Global Secondary Index for customer order history queries
  # Enables: "Get all orders for customer X" without table scan
  # Query: SELECT * WHERE CustomerId = 'cust_123' ORDER BY CreatedAt DESC
  global_secondary_index {
    name            = "CustomerIndex"
    hash_key        = "CustomerId"
    range_key       = "CreatedAt"
    projection_type = "ALL" # Project all attributes (no additional lookups needed)
    # Note: billing_mode inherited from table; no separate throughput config for PAY_PER_REQUEST
  }

  # Enable DynamoDB Streams
  stream_enabled   = true
  stream_view_type = "NEW_IMAGE" # We only need the newly inserted order record

  point_in_time_recovery {
    enabled = var.environment == "prod" # Enable PITR automatically only in prod

    # How to verify: Run `aws dynamodb describe-table --table-name <table-name>` and check PointInTimeRecoveryDescription.
  }

  tags = {
    Name = "${var.project_name}-orders-${var.environment}"
  }
}

output "dynamodb_table_name" {
  description = "The name of the DynamoDB table"
  value       = aws_dynamodb_table.orders.name
}

output "dynamodb_table_arn" {
  description = "The ARN of the DynamoDB table"
  value       = aws_dynamodb_table.orders.arn
}
