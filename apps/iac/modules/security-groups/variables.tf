variable "project" {
  description = "Project name"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, stg, prd)"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "app_port" {
  description = "Application port for ECS tasks"
  type        = number
  default     = 8080
}

variable "admin_cidr_blocks" {
  description = "CIDR blocks for admin access to Aurora (e.g., VPN, bastion)"
  type        = list(string)
  default     = null
}

variable "bastion_security_group_id" {
  description = "Security group ID of the bastion host for database access"
  type        = string
  default     = null
}

variable "tags" {
  description = "Additional tags for resources"
  type        = map(string)
  default     = {}
}
