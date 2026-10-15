# 1. Lambda Consumer Function
resource "aws_lambda_function" "stream_processor" {
  filename         = "../cmd/verifications/stream-processor/function.zip"
  function_name    = "${var.project_name}-stream-processor-${var.environment}"
  role             = aws_iam_role.lambda_exec.arn
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  source_code_hash = filebase64sha256("../cmd/verifications/stream-processor/function.zip")

  environment {
    variables = {
      STATE_MACHINE_ARN = aws_sfn_state_machine.order_workflow.arn
    }
  }
}

# 2. Event Source Mapping connecting DynamoDB Stream to Lambda with Filters
resource "aws_lambda_event_source_mapping" "order_stream_mapping" {
  event_source_arn  = aws_dynamodb_table.orders.stream_arn
  function_name     = aws_lambda_function.stream_processor.arn
  starting_position = "TRIM_HORIZON"
  batch_size         = 5 # Process orders in small batches

  # CRITICAL COST SAVING: Only fire for new PENDING records!
  filter_criteria {
    filter {
      pattern = jsonencode({
        eventName = ["INSERT"]
        dynamodb = {
          NewImage = {
            Status = {
              S = ["PENDING"]
            }
          }
        }
      })
    }
  }
}
