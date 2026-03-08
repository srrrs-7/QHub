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

variable "create_github_oidc_provider" {
  description = "Create GitHub OIDC provider (only once per account)"
  type        = bool
  default     = false
}

variable "github_oidc_provider_arn" {
  description = "Existing GitHub OIDC provider ARN"
  type        = string
  default     = null
}

variable "github_repository" {
  description = "GitHub repository (owner/repo)"
  type        = string
}

variable "ecr_repository_arns" {
  description = "ECR repository ARNs"
  type        = list(string)
}

variable "ecs_cluster_name" {
  description = "ECS cluster name"
  type        = string
}

variable "ecs_task_role_arns" {
  description = "ECS task role ARNs (for iam:PassRole)"
  type        = list(string)
}

variable "ecs_job_task_definition_arns" {
  description = "ECS Job task definition ARNs (for RunTask)"
  type        = list(string)
  default     = []
}

variable "vpc_subnet_arns" {
  description = "VPC subnet ARNs (for RunTask network configuration)"
  type        = list(string)
  default     = []
}

variable "rds_resource_id" {
  description = "RDS cluster resource ID for IAM auth"
  type        = string
  default     = null
}

variable "enable_rds_iam_auth" {
  description = "Enable RDS IAM authentication (set to true when RDS is created)"
  type        = bool
  default     = false
}

variable "rds_db_username" {
  description = "RDS database username for IAM auth"
  type        = string
  default     = null
}

variable "tags" {
  description = "Additional tags for resources"
  type        = map(string)
  default     = {}
}
