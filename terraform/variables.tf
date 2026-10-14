variable "region" {
  type        = string
  description = "AWS region for deploying resources"
  default     = "ap-southeast-2"
}

variable "environment" {
  type        = string
  description = "Target deployment environment (e.g., dev, staging, prod)"
  default     = "dev"
}

variable "project_name" {
  type        = string
  description = "Name of the project used for tagging and resource naming"
  default     = "ecommerce-order"
}
