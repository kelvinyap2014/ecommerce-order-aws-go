resource "aws_dynamodb_table" "orders" {
  name         = "${var.project_name}-orders-${var.environment}"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "OrderId"

  attribute {
    name = "OrderId"
    type = "S"
  }

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
