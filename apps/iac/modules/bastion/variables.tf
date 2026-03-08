variable "project" {
  description = "Project name"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, stg, prd)"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "subnet_id" {
  description = "Subnet ID for the bastion host (should be a private subnet with SSM endpoints)"
  type        = string
}

variable "instance_type" {
  description = "EC2 instance type for bastion"
  type        = string
  default     = "t4g.nano" # Smallest ARM instance (~$3/month)
}

variable "db_security_group_ids" {
  description = "Security group IDs of the database to allow access"
  type        = list(string)
  default     = []
}

variable "enable_rds_iam_auth" {
  description = "Enable IAM authentication for RDS"
  type        = bool
  default     = false
}

variable "rds_resource_id" {
  description = "RDS resource ID for IAM authentication"
  type        = string
  default     = null
}

variable "secrets_arns" {
  description = "ARNs of secrets the bastion can access"
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Additional tags for resources"
  type        = map(string)
  default     = {}
}
