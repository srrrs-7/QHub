variable "project" {
  description = "Project name"
  type        = string
}

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "database_name" {
  description = "Database name"
  type        = string
}

variable "master_username" {
  description = "Master username"
  type        = string
  default     = "postgres"
}

variable "engine_version" {
  description = "PostgreSQL engine version"
  type        = string
  default     = "16.4"
}

variable "engine_version_major" {
  description = "PostgreSQL major version (for parameter group)"
  type        = string
  default     = "16"
}

variable "instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t4g.micro" # Smallest ARM-based instance
}

variable "allocated_storage" {
  description = "Allocated storage in GB"
  type        = number
  default     = 20 # Minimum for gp3
}

variable "max_allocated_storage" {
  description = "Maximum allocated storage for autoscaling (0 to disable)"
  type        = number
  default     = 0 # Disabled for cost control
}

variable "storage_type" {
  description = "Storage type (gp2, gp3, io1)"
  type        = string
  default     = "gp2" # gp2 is cheaper for small workloads
}

variable "multi_az" {
  description = "Enable Multi-AZ deployment"
  type        = bool
  default     = false # Single AZ for dev
}

variable "db_subnet_group_name" {
  description = "DB subnet group name"
  type        = string
}

variable "security_group_ids" {
  description = "Security group IDs"
  type        = list(string)
}

variable "enable_iam_auth" {
  description = "Enable IAM database authentication"
  type        = bool
  default     = true
}

variable "backup_retention_period" {
  description = "Backup retention period in days"
  type        = number
  default     = 1 # Minimum for cost savings
}

variable "deletion_protection" {
  description = "Enable deletion protection"
  type        = bool
  default     = false
}

variable "skip_final_snapshot" {
  description = "Skip final snapshot on deletion"
  type        = bool
  default     = true
}

variable "enable_performance_insights" {
  description = "Enable Performance Insights"
  type        = bool
  default     = false # Disabled for cost savings
}

variable "enable_cloudwatch_logs" {
  description = "Enable CloudWatch Logs export"
  type        = bool
  default     = false # Disabled for cost savings
}

variable "tags" {
  description = "Additional tags"
  type        = map(string)
  default     = {}
}

# -----------------------------------------------------------------------------
# Database Users Configuration
# -----------------------------------------------------------------------------
variable "create_db_users" {
  description = "Whether to create database users (api_user, migrate_user)"
  type        = bool
  default     = false
}

variable "api_db_username" {
  description = "Username for API database user"
  type        = string
  default     = "api_user"
}

variable "migrate_db_username" {
  description = "Username for migration database user"
  type        = string
  default     = "migrate_user"
}
