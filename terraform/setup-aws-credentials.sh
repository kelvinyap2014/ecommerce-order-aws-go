#!/bin/bash
# Helper script to extract AWS credentials from new AWS experience cache
# and set them as environment variables for Terraform

CACHE_FILE=$(ls ~/.aws/login/cache/*.json 2>/dev/null | head -1)

if [ -z "$CACHE_FILE" ]; then
    echo "Error: No AWS login cache found. Run 'aws sts get-caller-identity' first to authenticate."
    exit 1
fi

# Extract credentials from JSON cache using jq
AWS_ACCESS_KEY_ID=$(jq -r '.accessToken.accessKeyId' "$CACHE_FILE" 2>/dev/null)
AWS_SECRET_ACCESS_KEY=$(jq -r '.accessToken.secretAccessKey' "$CACHE_FILE" 2>/dev/null)
AWS_SESSION_TOKEN=$(jq -r '.accessToken.sessionToken' "$CACHE_FILE" 2>/dev/null)

if [ -z "$AWS_ACCESS_KEY_ID" ] || [ "$AWS_ACCESS_KEY_ID" = "null" ]; then
    echo "Error: Failed to extract credentials from cache file: $CACHE_FILE"
    exit 1
fi

# Export credentials
export AWS_ACCESS_KEY_ID
export AWS_SECRET_ACCESS_KEY
export AWS_SESSION_TOKEN

echo "✅ AWS credentials loaded from new AWS experience cache"
echo "Access Key ID: ${AWS_ACCESS_KEY_ID:0:10}..."
echo "Ready to run terraform. Example:"
echo "  terraform plan -var-file='terraform-dev.tfvars'"
