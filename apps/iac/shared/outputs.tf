# Shared Outputs Configuration
# This file is symlinked from each environment directory
# All environments share the same outputs structure

# -----------------------------------------------------------------------------
# VPC
# -----------------------------------------------------------------------------
output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "private_subnet_ids" {
  description = "Private subnet IDs"
  value       = module.vpc.private_subnet_ids
}

output "public_subnet_ids" {
  description = "Public subnet IDs"
  value       = module.vpc.public_subnet_ids
}

# -----------------------------------------------------------------------------
# Database
# -----------------------------------------------------------------------------
output "aurora_cluster_endpoint" {
  description = "Aurora cluster endpoint for write operations"
  value       = module.aurora.cluster_endpoint
}

output "aurora_cluster_reader_endpoint" {
  description = "Aurora cluster reader endpoint for read operations"
  value       = module.aurora.cluster_reader_endpoint
}

output "aurora_master_password_secret_arn" {
  description = "ARN of the Secrets Manager secret containing the master password"
  value       = module.aurora.master_password_secret_arn
  sensitive   = true
}

# -----------------------------------------------------------------------------
# ECR
# -----------------------------------------------------------------------------
output "ecr_repository_url" {
  description = "ECR repository URL for API container images"
  value       = module.ecr_api.repository_url
}

# -----------------------------------------------------------------------------
# ECS
# -----------------------------------------------------------------------------
output "ecs_cluster_name" {
  description = "ECS cluster name"
  value       = module.ecs.cluster_name
}

output "ecs_service_name" {
  description = "ECS service name"
  value       = module.ecs.service_name
}

output "ecs_task_definition_family" {
  description = "ECS task definition family"
  value       = module.ecs.task_definition_family
}

# -----------------------------------------------------------------------------
# ALB
# -----------------------------------------------------------------------------
output "alb_dns_name" {
  description = "ALB DNS name"
  value       = module.alb.alb_dns_name
}

# -----------------------------------------------------------------------------
# CloudFront
# -----------------------------------------------------------------------------
output "cloudfront_distribution_id" {
  description = "CloudFront distribution ID"
  value       = module.cloudfront.distribution_id
}

output "cloudfront_domain_name" {
  description = "CloudFront distribution domain name"
  value       = module.cloudfront.distribution_domain_name
}

# -----------------------------------------------------------------------------
# S3
# -----------------------------------------------------------------------------
output "s3_assets_bucket" {
  description = "S3 bucket name for static assets"
  value       = module.s3_assets.bucket_id
}

# -----------------------------------------------------------------------------
# Cognito
# -----------------------------------------------------------------------------
output "cognito_user_pool_id" {
  description = "Cognito User Pool ID"
  value       = module.cognito.user_pool_id
}

output "cognito_client_id" {
  description = "Cognito App Client ID"
  value       = module.cognito.client_id
}

output "cognito_user_pool_endpoint" {
  description = "Cognito User Pool endpoint"
  value       = module.cognito.user_pool_endpoint
}

# -----------------------------------------------------------------------------
# IAM
# -----------------------------------------------------------------------------
output "github_actions_role_arn" {
  description = "IAM role ARN for GitHub Actions OIDC authentication"
  value       = module.iam.github_actions_role_arn
}

# -----------------------------------------------------------------------------
# URLs
# -----------------------------------------------------------------------------
output "app_url" {
  description = "Application URL (CloudFront)"
  value       = var.domain_name != null ? "https://${var.domain_name}" : "https://${module.cloudfront.distribution_domain_name}"
}

output "api_url" {
  description = "API URL (ALB)"
  value       = var.domain_name != null ? "https://api.${var.domain_name}" : "http://${module.alb.alb_dns_name}"
}
