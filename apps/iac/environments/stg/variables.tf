# Staging Environment Variables
# Staging-grade defaults with moderate availability and security

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
  default     = "stg"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "ap-northeast-1"
}

# -----------------------------------------------------------------------------
# VPC
# -----------------------------------------------------------------------------
variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
  default     = "10.1.0.0/16"
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
  default     = true
}

variable "enable_flow_logs" {
  description = "Enable VPC Flow Logs"
  type        = bool
  default     = true
}

variable "enable_vpc_endpoints" {
  description = "Enable VPC Endpoints"
  type        = bool
  default     = true
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
  default     = 2
}

variable "aurora_min_capacity" {
  description = "Aurora Serverless v2 minimum ACU"
  type        = number
  default     = 0.5
}

variable "aurora_max_capacity" {
  description = "Aurora Serverless v2 maximum ACU"
  type        = number
  default     = 8
}

variable "aurora_backup_retention_period" {
  description = "Aurora backup retention period (days)"
  type        = number
  default     = 14
}

variable "aurora_deletion_protection" {
  description = "Aurora deletion protection"
  type        = bool
  default     = true
}

variable "aurora_skip_final_snapshot" {
  description = "Skip final snapshot on deletion"
  type        = bool
  default     = false
}

# -----------------------------------------------------------------------------
# ECS
# -----------------------------------------------------------------------------
variable "ecs_task_cpu" {
  description = "ECS task CPU units"
  type        = number
  default     = 512
}

variable "ecs_task_memory" {
  description = "ECS task memory (MB)"
  type        = number
  default     = 1024
}

variable "ecs_desired_count" {
  description = "ECS desired task count"
  type        = number
  default     = 2
}

variable "ecs_enable_autoscaling" {
  description = "Enable ECS auto scaling"
  type        = bool
  default     = true
}

variable "ecs_min_capacity" {
  description = "ECS minimum capacity"
  type        = number
  default     = 2
}

variable "ecs_max_capacity" {
  description = "ECS maximum capacity"
  type        = number
  default     = 4
}

variable "ecs_enable_execute_command" {
  description = "Enable ECS Exec for debugging"
  type        = bool
  default     = true
}

# -----------------------------------------------------------------------------
# ALB
# -----------------------------------------------------------------------------
variable "alb_deletion_protection" {
  description = "ALB deletion protection"
  type        = bool
  default     = true
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
  default     = true
}

variable "waf_rate_limit" {
  description = "WAF rate limit (requests per 5 minutes)"
  type        = number
  default     = 2000
}

# -----------------------------------------------------------------------------
# Domain
# -----------------------------------------------------------------------------
variable "domain_name" {
  description = "Domain name"
  type        = string
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
  description = "Cognito MFA configuration (OFF, OPTIONAL, ON)"
  type        = string
  default     = "OPTIONAL"
}

variable "cognito_deletion_protection" {
  description = "Cognito deletion protection"
  type        = bool
  default     = true
}

variable "cognito_advanced_security_mode" {
  description = "Cognito advanced security mode (OFF, AUDIT, ENFORCED)"
  type        = string
  default     = "AUDIT"
}

variable "cognito_callback_urls" {
  description = "Cognito callback URLs"
  type        = list(string)
  default     = []
}

variable "cognito_logout_urls" {
  description = "Cognito logout URLs"
  type        = list(string)
  default     = []
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
