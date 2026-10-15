resource "aws_sfn_state_machine" "order_workflow" {
  name     = "${var.project_name}-workflow-${var.environment}"
  role_arn = aws_iam_role.sfn_exec.arn

  definition = templatefile("../statemachine/order_workflow.asl.json", {
    AddressLambdaArn      = aws_lambda_function.check_address.arn
    CreditLambdaArn       = aws_lambda_function.check_credit.arn
    PaymentLambdaArn      = aws_lambda_function.check_payment.arn
    UpdateStatusLambdaArn = aws_lambda_function.update_status.arn
  })
}

output "state_machine_arn" {
  description = "The ARN of the Step Functions order orchestration workflow"
  value       = aws_sfn_state_machine.order_workflow.arn
}
