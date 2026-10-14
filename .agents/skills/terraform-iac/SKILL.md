# Skill: Terraform & Infrastructure as Code for E-Commerce Order Orchestration

This skill provides step-by-step guidance for developing, testing, and deploying the serverless e-commerce order orchestration infrastructure using Terraform and AWS.

---

## 1. Architecture Overview

The infrastructure consists of five core AWS services coordinated via Terraform:

| Service | File | Purpose |
|---------|------|---------|
| **DynamoDB** | `dynamodb.tf` | Stores order state (`PENDING`, `VERIFIED`, `CONFIRMED`, `REJECTED`); GSI enables customer order history queries |
| **Lambda** | `lambda.tf` | Packages and deploys Go binaries (ingestion API + 5 verification/dispatch functions) |
| **API Gateway** | `api_gateway.tf` | HTTP API endpoint for order submission (`POST /orders`); routes to Ingestion Lambda |
| **Step Functions** | `step_functions.tf` | Deploys ASL state machine orchestrating address checks, credit checks, payment auth, WMS dispatch |
| **IAM** | Embedded in each `*.tf` | Manages execution roles with least-privilege permissions for each Lambda and Step Functions |

All resources are tagged automatically with `Project` and `Environment` from `variables.tf`.

---

## 2. Project File Structure

```
terraform/
├── main.tf                   ← AWS provider config, Terraform version constraints
├── variables.tf              ← Common variables (environment, project_name, etc.)
├── dynamodb.tf               ← Orders table, GSI definitions, billing mode
├── lambda.tf                 ← Lambda function definitions, IAM execution roles, Go binary packaging
├── step_functions.tf         ← ASL state machine definition, Lambda invocation permissions
├── api_gateway.tf            ← HTTP API routes, stage config, CloudWatch logging
├── terraform-dev.tfvars      ← Variable overrides for DEV (e.g., environment = "dev")
├── terraform-prod.tfvars     ← Variable overrides for PROD (e.g., environment = "prod")
```

**Key points:**
- Region is defined in `variables.tf` with default `ap-southeast-2` and can be overridden in `terraform-*tfvars` files.
- All environment-specific settings (region, instance counts, billing mode, log retention) are in `terraform-*tfvars` files.
- `lambda.tf` handles Go binary compilation and packaging; assumes binaries are pre-built in `../cmd/` directories.

---

## 3. Prerequisites & Setup

### 3.1 AWS Credentials & Configuration

Before running any Terraform command, verify AWS credentials:

```bash
# Check if AWS CLI is configured
aws sts get-caller-identity
# Output should show your account ID, user ARN, etc.

# If not configured yet:
aws configure
# You'll be prompted for Access Key ID, Secret Access Key, region (use ap-southeast-2), output format (json)
```

### 3.2 Go Build Prerequisites

Lambda functions are written in Go and must be compiled to Linux binaries before Terraform packages them:

```bash
# Verify Go version (1.18+ required)
go version

# Compile all Lambda functions for Linux/amd64
# (This is typically done in the project Makefile or manually per function)
GOOS=linux GOARCH=amd64 go build -o cmd/api/bootstrap cmd/api/main.go
GOOS=linux GOARCH=amd64 go build -o cmd/verifications/address-check/bootstrap cmd/verifications/address-check/main.go
# ... (repeat for other verifications and wms-integration)
```

Many projects automate this in a `Makefile` — check the repo's Makefile for a `build` or `build-lambda` target.

### 3.3 Terraform Initialization

```bash
cd terraform

# Initialize Terraform (downloads AWS provider plugin v5.0+, sets up .terraform/ folder)
terraform init

# Verify initialization
ls -la .terraform/
# Should see plugins/ and modules/ directories
```

---

## 4. Development Workflow

### 4.1 Plan Changes (DEV Environment)

After editing any `.tf` file or changing Go source code (and rebuilding binaries), review planned changes:

```bash
cd terraform

# Display all planned resource creations, updates, deletions
terraform plan -var-file="terraform-dev.tfvars"

# Save the plan to a file for review or apply later
terraform plan -var-file="terraform-dev.tfvars" -out=tfplan

# Review what will be applied
terraform show tfplan
```

**What to look for:**
- Resource counts (should match your intent: 1 DynamoDB table, 1 API Gateway, 6 Lambdas, 1 Step Functions state machine)
- Attribute changes (e.g., `runtime = "go1.x"` updated? `environment = "dev"` set correctly?)
- No unexpected deletions (e.g., `-/+ DynamoDB table` would delete data)

### 4.2 Apply Changes (DEV Environment)

Once the plan looks correct:

```bash
cd terraform

# Apply changes interactively (you'll be prompted with a summary, type 'yes' to confirm)
terraform apply -var-file="terraform-dev.tfvars"

# Or, apply without confirmation (useful for CI/CD)
terraform apply -var-file="terraform-dev.tfvars" -auto-approve

# Or, apply a pre-saved plan
terraform apply tfplan
```

**Expected output:**
```
Apply complete! Resources added: 11. Changes to infrastructure may take a few seconds to become available.

Outputs:
api_gateway_url = "https://abc123def.execute-api.ap-southeast-2.amazonaws.com"
dynamodb_table_name = "ecommerce-order-orders-dev"
...
```

### 4.3 Verify Deployment

After `terraform apply` succeeds:

```bash
# Describe the DynamoDB table
aws dynamodb describe-table --table-name ecommerce-order-orders-dev --region ap-southeast-2

# Test the API Gateway endpoint
curl -X POST "$(terraform output -raw api_gateway_url)/orders" \
  -H "Content-Type: application/json" \
  -d '{"customerId": "cust-123", "totalAmount": 99.99}'

# Check Lambda function
aws lambda function-configuration \
  --function-name ecommerce-order-orders-api-dev \
  --region ap-southeast-2
```

---

## 5. Configuration Management

### 5.1 Understanding `main.tf`

```terraform
provider "aws" {
  region = var.region
  default_tags {
    tags = {
      Project     = var.project_name      # From terraform-*tfvars
      Environment = var.environment       # From terraform-*tfvars
    }
  }
}
```

**Region configuration:** Define the `region` variable in `terraform-*tfvars` files (defaults to `ap-southeast-2` in `variables.tf`). If you need to change regions, update your AWS Settings (and rebuild the dependent infrastructure). Cross-region resources are not supported.

### 5.2 Environment Variable Files

**`terraform-dev.tfvars`:**
```hcl
region       = "ap-southeast-2"
environment  = "dev"
project_name = "ecommerce-order"
# Additional dev-specific overrides (e.g., smaller instance counts, longer log retention)
```

**`terraform-prod.tfvars`:**
```hcl
region       = "ap-southeast-2"
environment  = "prod"
project_name = "ecommerce-order"
# Production-specific overrides (e.g., enhanced monitoring, redundancy, stricter retention)
```

**Using different environments:**
```bash
# Deploy DEV
terraform plan -var-file="terraform-dev.tfvars"
terraform apply -var-file="terraform-dev.tfvars"

# Deploy PROD (separate state files, independent infrastructure)
terraform plan -var-file="terraform-prod.tfvars"
terraform apply -var-file="terraform-prod.tfvars"
```

---

## 6. Working with State Files

### 6.1 Local State (Default)

By default, Terraform stores state in `terraform.tfstate` (and `terraform.tfstate.backup`) in the current directory.

```bash
# Inspect the current state
terraform state list

# Show details about a specific resource
terraform state show aws_dynamodb_table.orders

# Query state outputs
terraform output
terraform output -raw api_gateway_url
```

### 6.2 Remote State (Recommended for Teams)

For multi-person teams, store state remotely (e.g., S3 with DynamoDB state locking). Uncomment and configure in `main.tf`:

```terraform
terraform {
  backend "s3" {
    bucket         = "your-terraform-state-bucket"
    key            = "ecommerce-order/dev/terraform.tfstate"
    region         = "ap-southeast-2"
    encrypt        = true
    dynamodb_table = "terraform-state-lock"
  }
}
```

Then:
```bash
terraform init  # Migrates state to S3
terraform plan -var-file="terraform-dev.tfvars"
```

### 6.3 Destroy & Clean Up

To remove all infrastructure (e.g., to reduce AWS costs):

```bash
cd terraform

# Preview what will be destroyed
terraform plan -destroy -var-file="terraform-dev.tfvars"

# Destroy all resources (you'll be prompted, type 'yes' to confirm)
terraform destroy -var-file="terraform-dev.tfvars"

# Or, destroy without confirmation
terraform destroy -var-file="terraform-dev.tfvars" -auto-approve
```

**Warning:** This deletes the DynamoDB table, Lambdas, API Gateway, and Step Functions state machine. Any data in DynamoDB is lost.

---

## 7. Lambda Deployment & Go Binaries

### 7.1 Lambda Handler Function Naming

AWS Lambda expects Go binaries to be named `bootstrap` for the provided.al2023 runtime or named after the handler for other runtimes. Verify your Go build outputs:

```bash
# Expected structure after compilation
cmd/api/bootstrap                           # Compiled Go binary (executable)
cmd/verifications/address-check/bootstrap   # Compiled Go binary
cmd/verifications/credit-check/bootstrap    # Compiled Go binary
...
```

### 7.2 Packaging Lambda Functions

`lambda.tf` uses the `archive_file` data source to ZIP each compiled Go binary:

```terraform
data "archive_file" "api_lambda" {
  type        = "zip"
  source_file = "${path.module}/../cmd/api/bootstrap"
  output_path = "${path.module}/../cmd/api/lambda.zip"
}

resource "aws_lambda_function" "api" {
  filename      = data.archive_file.api_lambda.output_path
  function_name = "ecommerce-order-orders-api-${var.environment}"
  role          = aws_iam_role.lambda_role.arn
  handler       = "bootstrap"  # Entry point for Lambda execution
  runtime       = "provided.al2023"  # Custom Go runtime
  source_code_hash = data.archive_file.api_lambda.output_base64sha256
}
```

**Workflow:**
1. Edit Go source in `cmd/api/main.go`
2. Recompile: `GOOS=linux GOARCH=amd64 go build -o cmd/api/bootstrap cmd/api/main.go`
3. Plan & apply Terraform: `terraform plan -var-file="terraform-dev.tfvars"` detects the binary change via hash mismatch
4. `terraform apply` re-ZIPs and redeploys the Lambda

### 7.3 Debugging Lambda Runtime Errors

If a Lambda fails after deployment:

```bash
# Check Lambda configuration
aws lambda get-function-configuration \
  --function-name ecommerce-order-orders-api-dev \
  --region ap-southeast-2

# Invoke the Lambda manually (test)
aws lambda invoke \
  --function-name ecommerce-order-orders-api-dev \
  --payload '{"body": "test"}' \
  --region ap-southeast-2 \
  response.json

# Check CloudWatch logs
aws logs tail /aws/lambda/ecommerce-order-orders-api-dev --follow --region ap-southeast-2
```

---

## 8. Step Functions State Machine

### 8.1 Viewing the ASL Definition

The state machine is defined in `step_functions.tf` (or loaded from `statemachine/order_workflow.asl.json`):

```bash
# Output the current state machine definition
terraform show | grep -A 50 "definition ="

# Or, retrieve it from AWS after deployment
aws stepfunctions describe-state-machine \
  --state-machine-arn "arn:aws:states:ap-southeast-2:ACCOUNT_ID:stateMachine:ecommerce-order-workflow-dev" \
  --region ap-southeast-2 \
  --query definition --output text | jq .
```

### 8.2 Testing the Workflow

After deployment, test the order orchestration end-to-end:

```bash
# Invoke the order ingestion API
curl -X POST "https://YOUR_API_GATEWAY_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "customerId": "cust-12345",
    "totalAmount": 150.00,
    "shippingAddress": {
      "street": "123 Main St",
      "city": "Sydney",
      "country": "AU"
    }
  }'

# The API returns an executionArn
# Monitor the Step Functions execution in CloudWatch or AWS Console
aws stepfunctions get-execution-history \
  --execution-arn "arn:aws:states:ap-southeast-2:ACCOUNT_ID:execution:..." \
  --region ap-southeast-2
```

### 8.3 Updating the Workflow

If you modify the ASL definition:

1. **Edit** the definition file or inline JSON in `step_functions.tf`
2. **Plan & apply:**
   ```bash
   terraform plan -var-file="terraform-dev.tfvars"
   terraform apply -var-file="terraform-dev.tfvars"
   ```
3. **Existing executions** continue with the old definition; new executions use the updated definition.

---

## 9. Common Scenarios & Troubleshooting

### Scenario: "terraform plan" fails with "Error initializing the backend"

**Cause:** Terraform not initialized or backend migration needed.
**Fix:**
```bash
cd terraform
terraform init
terraform plan -var-file="terraform-dev.tfvars"
```

### Scenario: "aws lambda invoke" returns "ResourceNotFoundException"

**Cause:** Lambda function not deployed or wrong region/environment.
**Fix:**
```bash
# Verify Lambda exists
aws lambda list-functions --region ap-southeast-2 | grep ecommerce-order

# If empty, check Terraform outputs
cd terraform
terraform output

# Redeploy if necessary
terraform apply -var-file="terraform-dev.tfvars" -auto-approve
```

### Scenario: "DynamoDB table not found" at runtime

**Cause:** Lambda IAM role lacks DynamoDB permissions.
**Fix:**
- Verify IAM role in `lambda.tf` includes `dynamodb:GetItem`, `dynamodb:PutItem`, etc.
- Check CloudWatch logs for permission errors
- Redeploy: `terraform apply -var-file="terraform-dev.tfvars" -auto-approve`

### Scenario: Go binary compilation fails

**Cause:** Incorrect GOOS/GOARCH or Go version issues.
**Fix:**
```bash
# Ensure Go 1.18+
go version

# Compile with explicit flags
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
  -ldflags="-w -s" \
  -o cmd/api/bootstrap \
  cmd/api/main.go

# Verify binary is executable
file cmd/api/bootstrap
# Should output: "ELF 64-bit LSB executable, x86-64, ..."
```

### Scenario: API Gateway returns 500 Internal Server Error

**Cause:** Lambda function crashed or timing out.
**Fix:**
```bash
# Check Lambda CloudWatch logs immediately
aws logs tail /aws/lambda/ecommerce-order-orders-api-dev \
  --follow --since 2m --region ap-southeast-2

# Increase Lambda timeout in lambda.tf (default is often 30s)
resource "aws_lambda_function" "api" {
  ...
  timeout = 60  # Increase from 30
}

# Redeploy
terraform apply -var-file="terraform-dev.tfvars" -auto-approve
```

---

## 10. Production Deployment Checklist

Before deploying to production:

- [ ] All Go binaries compiled successfully for Linux amd64
- [ ] `terraform plan -var-file="terraform-prod.tfvars"` reviewed and approved
- [ ] IAM roles and policies reviewed for least-privilege access
- [ ] DynamoDB table configured with appropriate backup/retention settings
- [ ] Lambda environment variables set correctly (e.g., secrets from AWS Secrets Manager)
- [ ] Step Functions workflow ASL validated (no syntax errors)
- [ ] API Gateway has CloudWatch logging enabled and appropriate throttling/quota settings
- [ ] Health check / monitoring configured (e.g., CloudWatch alarms on Lambda error rate)
- [ ] Disaster recovery / state migration plan documented (e.g., backup DynamoDB table)

**Deploy production:**
```bash
cd terraform
terraform plan -var-file="terraform-prod.tfvars"
# Review carefully
terraform apply -var-file="terraform-prod.tfvars"
```

---

## 11. Further Resources

- **Terraform AWS Provider Docs:** https://registry.terraform.io/providers/hashicorp/aws/latest
- **AWS Lambda Go Runtime:** https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html
- **Step Functions ASL:** https://docs.aws.amazon.com/step-functions/latest/dg/concepts-amazon-states-language.html
- **DynamoDB with Terraform:** https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/dynamodb_table
