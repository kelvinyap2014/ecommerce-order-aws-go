terraform {
  required_version = ">= 1.5.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "ap-southeast-2" # Change if you prefer another region (e.g., us-east-1)

  default_tags {
    tags = {
      Project     = var.project_name
      Environment = var.environment
    }
  }
}
