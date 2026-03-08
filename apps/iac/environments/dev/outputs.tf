# Development Environment Outputs
# Values to configure in GitHub Environment Variables

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
output "database_endpoint" {
  description = "Database endpoint for write operations"
  value       = local.db_endpoint
}

output "database_port" {
  description = "Database port"
  value       = local.db_port
}

output "database_name" {
  description = "Database name"
  value       = local.db_name
}

# -----------------------------------------------------------------------------
# ECR
# -----------------------------------------------------------------------------
output "ecr_repository_api_url" {
  description = "ECR repository URL for API container images"
  value       = module.ecr_api.repository_url
}

output "ecr_repository_api_name" {
  description = "ECR repository name for API (ECR_REPOSITORY_API)"
  value       = module.ecr_api.repository_name
}

output "ecr_repository_migrate_url" {
  description = "ECR repository URL for migration container images"
  value       = module.ecr_migrate.repository_url
}

output "ecr_repository_migrate_name" {
  description = "ECR repository name for migration (ECR_REPOSITORY_MIGRATE)"
  value       = module.ecr_migrate.repository_name
}

# -----------------------------------------------------------------------------
# ECS
# -----------------------------------------------------------------------------
output "ecs_cluster_name" {
  description = "ECS cluster name (ECS_CLUSTER)"
  value       = module.ecs.cluster_name
}

output "ecs_service_name" {
  description = "ECS service name (ECS_SERVICE_API)"
  value       = module.ecs.service_name
}

output "ecs_task_definition_family" {
  description = "ECS task definition family for API"
  value       = module.ecs.task_definition_family
}

output "ecs_task_definition_migrate" {
  description = "ECS task definition family for migration (ECS_TASK_DEFINITION_MIGRATE)"
  value       = module.ecs_migrate.task_definition_family
}

# -----------------------------------------------------------------------------
# Security Groups
# -----------------------------------------------------------------------------
output "ecs_security_group_id" {
  description = "Security group ID for ECS tasks (ECS_SECURITY_GROUP_ID)"
  value       = module.security_groups.ecs_security_group_id
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

# -----------------------------------------------------------------------------
# GitHub Environment Variables Summary
# -----------------------------------------------------------------------------
output "github_environment_variables" {
  description = "Summary of values to set in GitHub Environment Variables for CD workflow"
  value = {
    AWS_REGION                  = var.aws_region
    AWS_ROLE_ARN                = module.iam.github_actions_role_arn
    ECR_REPOSITORY_API          = module.ecr_api.repository_name
    ECR_REPOSITORY_MIGRATE      = module.ecr_migrate.repository_name
    ECS_CLUSTER                 = module.ecs.cluster_name
    ECS_SERVICE_API             = module.ecs.service_name
    CONTAINER_NAME_API          = "api"
    ECS_TASK_DEFINITION_MIGRATE = module.ecs_migrate.task_definition_family
    ECS_SUBNET_IDS              = join(",", module.vpc.private_subnet_ids)
    ECS_SECURITY_GROUP_ID       = module.security_groups.ecs_security_group_id
    S3_BUCKET_WEB               = module.s3_assets.bucket_id
    CLOUDFRONT_DISTRIBUTION_ID  = module.cloudfront.distribution_id
  }
}
