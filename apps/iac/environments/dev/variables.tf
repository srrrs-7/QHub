# -----------------------------------------------------------------------------
# General
# -----------------------------------------------------------------------------
variable "project" {
  description = "Project name"
  type        = string
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "ap-northeast-1"
}

variable "aws_profile" {
  description = "AWS CLI profile name"
  type        = string
  default     = null
}

# -----------------------------------------------------------------------------
# VPC
# -----------------------------------------------------------------------------
variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
  default     = "10.0.0.0/16"
}

variable "az_count" {
  description = "Number of availability zones"
  type        = number
  default     = 2
}

variable "enable_nat_gateway" {
  description = "Enable NAT Gateway"
  type        = bool
  default     = true
}

variable "single_nat_gateway" {
  description = "Use single NAT Gateway (cost saving)"
  type        = bool
  default     = true # Single NAT for dev to save costs
}

variable "enable_flow_logs" {
  description = "Enable VPC Flow Logs"
  type        = bool
  default     = true
}

variable "enable_vpc_endpoints" {
  description = "Enable VPC Endpoints (S3 Gateway)"
  type        = bool
  default     = true
}

variable "enable_interface_endpoints" {
  description = "Enable Interface VPC Endpoints (ECR, Logs, SecretsManager). Disable for cost savings."
  type        = bool
  default     = false # Disabled for dev to save ~$30/month
}

# -----------------------------------------------------------------------------
# Application
# -----------------------------------------------------------------------------
variable "app_port" {
  description = "Application port"
  type        = number
  default     = 8080
}

variable "health_check_path" {
  description = "Health check path"
  type        = string
  default     = "/health"
}

# -----------------------------------------------------------------------------
# Database
# -----------------------------------------------------------------------------
variable "database_name" {
  description = "Database name"
  type        = string
}

variable "database_master_username" {
  description = "Database master username"
  type        = string
  default     = "postgres"
}

variable "database_app_username" {
  description = "Database application username (for IAM auth)"
  type        = string
  default     = "app_user"
}

variable "create_db_users" {
  description = "Create database users (api_user, migrate_user) via PostgreSQL provider. Requires network access to DB."
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

variable "database_type" {
  description = "Database type: 'aurora' for Aurora Serverless v2, 'rds' for single RDS instance"
  type        = string
  default     = "rds" # RDS is cheaper for dev
}

# Aurora settings (used when database_type = "aurora")
variable "aurora_engine_version" {
  description = "Aurora PostgreSQL engine version"
  type        = string
  default     = "16.4"
}

variable "aurora_engine_version_major" {
  description = "Aurora PostgreSQL major version"
  type        = string
  default     = "16"
}

variable "aurora_instance_count" {
  description = "Number of Aurora instances"
  type        = number
  default     = 1 # Single instance for dev
}

variable "aurora_min_capacity" {
  description = "Aurora Serverless v2 minimum ACU"
  type        = number
  default     = 0.5
}

variable "aurora_max_capacity" {
  description = "Aurora Serverless v2 maximum ACU"
  type        = number
  default     = 4 # Lower max for dev
}

variable "aurora_backup_retention_period" {
  description = "Aurora backup retention period"
  type        = number
  default     = 7
}

variable "aurora_deletion_protection" {
  description = "Aurora deletion protection"
  type        = bool
  default     = false # Disabled for dev
}

variable "aurora_skip_final_snapshot" {
  description = "Skip final snapshot on deletion"
  type        = bool
  default     = true # Skip for dev
}

# RDS settings (used when database_type = "rds")
variable "rds_engine_version" {
  description = "RDS PostgreSQL engine version"
  type        = string
  default     = "16.4"
}

variable "rds_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t4g.micro" # Smallest ARM-based instance (~$13/month)
}

variable "rds_allocated_storage" {
  description = "RDS allocated storage in GB"
  type        = number
  default     = 20
}

variable "rds_backup_retention_period" {
  description = "RDS backup retention period in days"
  type        = number
  default     = 1 # Minimum for cost savings
}

# -----------------------------------------------------------------------------
# ECS
# -----------------------------------------------------------------------------
variable "ecs_task_cpu" {
  description = "ECS task CPU units"
  type        = number
  default     = 256
}

variable "ecs_task_memory" {
  description = "ECS task memory (MB)"
  type        = number
  default     = 512
}

variable "ecs_desired_count" {
  description = "ECS desired task count"
  type        = number
  default     = 1
}

variable "ecs_enable_autoscaling" {
  description = "Enable ECS auto scaling"
  type        = bool
  default     = false # Disabled for dev
}

variable "ecs_min_capacity" {
  description = "ECS minimum capacity"
  type        = number
  default     = 1
}

variable "ecs_max_capacity" {
  description = "ECS maximum capacity"
  type        = number
  default     = 2
}

variable "ecs_enable_execute_command" {
  description = "Enable ECS Exec for debugging"
  type        = bool
  default     = true # Enabled for dev
}

variable "ecs_use_fargate_spot" {
  description = "Use Fargate Spot for cost savings (up to 70% cheaper, but can be interrupted)"
  type        = bool
  default     = true # Enabled for dev to save costs
}

# -----------------------------------------------------------------------------
# ALB
# -----------------------------------------------------------------------------
variable "alb_deletion_protection" {
  description = "ALB deletion protection"
  type        = bool
  default     = false # Disabled for dev
}

# -----------------------------------------------------------------------------
# CloudFront
# -----------------------------------------------------------------------------
variable "cloudfront_price_class" {
  description = "CloudFront price class"
  type        = string
  default     = "PriceClass_200"
}

# -----------------------------------------------------------------------------
# WAF
# -----------------------------------------------------------------------------
variable "enable_waf" {
  description = "Enable WAF"
  type        = bool
  default     = false # Disabled for dev to save costs
}

variable "waf_rate_limit" {
  description = "WAF rate limit"
  type        = number
  default     = 2000
}

# -----------------------------------------------------------------------------
# Domain
# -----------------------------------------------------------------------------
variable "domain_name" {
  description = "Domain name (set to null to skip domain-related resources)"
  type        = string
  default     = null
}

variable "create_hosted_zone" {
  description = "Create Route53 hosted zone"
  type        = bool
  default     = false
}

variable "hosted_zone_id" {
  description = "Existing Route53 hosted zone ID"
  type        = string
  default     = null
}

# -----------------------------------------------------------------------------
# Cognito
# -----------------------------------------------------------------------------
variable "cognito_mfa_configuration" {
  description = "Cognito MFA configuration"
  type        = string
  default     = "OPTIONAL"
}

variable "cognito_deletion_protection" {
  description = "Cognito deletion protection"
  type        = bool
  default     = false # Disabled for dev
}

variable "cognito_advanced_security_mode" {
  description = "Cognito advanced security mode"
  type        = string
  default     = "OFF" # Disabled for dev
}

variable "cognito_callback_urls" {
  description = "Cognito callback URLs"
  type        = list(string)
  default     = ["http://localhost:3000/callback"]
}

variable "cognito_logout_urls" {
  description = "Cognito logout URLs"
  type        = list(string)
  default     = ["http://localhost:3000/logout"]
}

# -----------------------------------------------------------------------------
# GitHub Actions
# -----------------------------------------------------------------------------
variable "create_github_oidc_provider" {
  description = "Create GitHub OIDC provider"
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

# -----------------------------------------------------------------------------
# Bastion Host
# -----------------------------------------------------------------------------
variable "enable_bastion" {
  description = "Enable bastion host for Session Manager access to private resources"
  type        = bool
  default     = false
}

variable "bastion_instance_type" {
  description = "EC2 instance type for bastion"
  type        = string
  default     = "t4g.nano"
}

variable "enable_ssm_endpoints" {
  description = "Enable SSM VPC Endpoints for Session Manager (required for bastion without NAT)"
  type        = bool
  default     = false
}
