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

variable "job_name" {
  description = "Job name (e.g., migrate)"
  type        = string
}

variable "container_name" {
  description = "Container name"
  type        = string
}

variable "container_image" {
  description = "Container image URL"
  type        = string
}

variable "task_cpu" {
  description = "Task CPU units"
  type        = number
  default     = 256
}

variable "task_memory" {
  description = "Task memory (MB)"
  type        = number
  default     = 512
}

variable "environment_variables" {
  description = "Environment variables for container"
  type = list(object({
    name  = string
    value = string
  }))
  default = []
}

variable "secrets" {
  description = "Secrets for container (from Secrets Manager)"
  type = list(object({
    name      = string
    valueFrom = string
  }))
  default = []
}

variable "secrets_arns" {
  description = "ARNs of secrets for IAM policy"
  type        = list(string)
  default     = []
}

variable "db_secret_arn" {
  description = "ARN of database secret (for password injection)"
  type        = string
  default     = null
}

variable "log_retention_days" {
  description = "CloudWatch log retention in days"
  type        = number
  default     = 14
}

variable "rds_resource_id" {
  description = "RDS cluster/instance resource ID for IAM auth"
  type        = string
  default     = null
}

variable "enable_rds_iam_auth" {
  description = "Enable RDS IAM authentication"
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
