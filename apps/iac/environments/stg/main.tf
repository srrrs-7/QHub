# Staging Environment
# Symlink to dev main.tf with different variables
# This file is identical to dev/main.tf

terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }

  backend "s3" {}
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = var.project
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}

provider "aws" {
  alias  = "us_east_1"
  region = "us-east-1"

  default_tags {
    tags = {
      Project     = var.project
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}

locals {
  tags = {
    Project     = var.project
    Environment = var.environment
  }
}

module "vpc" {
  source = "../../modules/vpc"

  project              = var.project
  environment          = var.environment
  aws_region           = var.aws_region
  vpc_cidr             = var.vpc_cidr
  az_count             = var.az_count
  enable_nat_gateway   = var.enable_nat_gateway
  single_nat_gateway   = var.single_nat_gateway
  enable_flow_logs     = var.enable_flow_logs
  enable_vpc_endpoints = var.enable_vpc_endpoints
  tags                 = local.tags
}

module "security_groups" {
  source = "../../modules/security-groups"

  project     = var.project
  environment = var.environment
  vpc_id      = module.vpc.vpc_id
  app_port    = var.app_port
  tags        = local.tags
}

module "ecr_api" {
  source = "../../modules/ecr"

  project         = var.project
  environment     = var.environment
  repository_name = "api"
  scan_on_push    = true
  tags            = local.tags
}

module "aurora" {
  source = "../../modules/aurora"

  project                 = var.project
  environment             = var.environment
  database_name           = var.database_name
  master_username         = var.database_master_username
  engine_version          = var.aurora_engine_version
  engine_version_major    = var.aurora_engine_version_major
  instance_count          = var.aurora_instance_count
  serverless_min_capacity = var.aurora_min_capacity
  serverless_max_capacity = var.aurora_max_capacity
  db_subnet_group_name    = module.vpc.db_subnet_group_name
  security_group_ids      = [module.security_groups.aurora_security_group_id]
  enable_iam_auth         = true
  backup_retention_period = var.aurora_backup_retention_period
  deletion_protection     = var.aurora_deletion_protection
  skip_final_snapshot     = var.aurora_skip_final_snapshot
  tags                    = local.tags
}

module "acm" {
  source = "../../modules/acm"

  domain_name               = var.domain_name
  subject_alternative_names = ["*.${var.domain_name}"]
  validate_certificate      = false
  tags                      = local.tags
}

module "acm_cloudfront" {
  source = "../../modules/acm"
  providers = {
    aws = aws.us_east_1
  }

  domain_name               = var.domain_name
  subject_alternative_names = ["*.${var.domain_name}"]
  validate_certificate      = false
  tags                      = local.tags
}

module "alb" {
  source = "../../modules/alb"

  project             = var.project
  environment         = var.environment
  vpc_id              = module.vpc.vpc_id
  subnet_ids          = module.vpc.public_subnet_ids
  security_group_ids  = [module.security_groups.alb_security_group_id]
  certificate_arn     = module.acm.certificate_arn
  target_port         = var.app_port
  health_check_path   = var.health_check_path
  deletion_protection = var.alb_deletion_protection
  tags                = local.tags
}

module "ecs" {
  source = "../../modules/ecs"

  project                  = var.project
  environment              = var.environment
  aws_region               = var.aws_region
  service_name             = "api"
  container_name           = "api"
  container_image          = "${module.ecr_api.repository_url}:latest"
  container_port           = var.app_port
  task_cpu                 = var.ecs_task_cpu
  task_memory              = var.ecs_task_memory
  desired_count            = var.ecs_desired_count
  subnet_ids               = module.vpc.private_subnet_ids
  security_group_ids       = [module.security_groups.ecs_security_group_id]
  target_group_arn         = module.alb.api_target_group_arn
  enable_autoscaling       = var.ecs_enable_autoscaling
  autoscaling_min_capacity = var.ecs_min_capacity
  autoscaling_max_capacity = var.ecs_max_capacity
  rds_resource_id          = module.aurora.cluster_resource_id
  rds_db_username          = var.database_app_username
  enable_rds_iam_auth      = true
  enable_execute_command   = var.ecs_enable_execute_command

  environment_variables = [
    { name = "DB_HOST", value = module.aurora.cluster_endpoint },
    { name = "DB_PORT", value = tostring(module.aurora.cluster_port) },
    { name = "DB_NAME", value = module.aurora.database_name },
    { name = "DB_USER", value = var.database_app_username },
    { name = "ENVIRONMENT", value = var.environment }
  ]

  tags = local.tags
}

module "s3_assets" {
  source = "../../modules/s3"

  bucket_name         = "${var.project}-${var.environment}-assets"
  enable_versioning   = true
  block_public_access = true
  cors_rules = [{
    allowed_headers = ["*"]
    allowed_methods = ["GET", "HEAD"]
    allowed_origins = ["https://${var.domain_name}"]
  }]
  tags = local.tags
}

module "cloudfront" {
  source = "../../modules/cloudfront"

  project                = var.project
  environment            = var.environment
  s3_origin_domain_name  = module.s3_assets.bucket_regional_domain_name
  alb_origin_domain_name = module.alb.alb_dns_name
  enable_s3_origin       = true
  aliases                = [var.domain_name, "www.${var.domain_name}"]
  certificate_arn        = module.acm_cloudfront.certificate_arn
  price_class            = var.cloudfront_price_class
  web_acl_id             = var.enable_waf ? module.waf[0].web_acl_arn : null
  custom_error_responses = [{ error_code = 404, response_code = 200, response_page_path = "/index.html" }]
  tags                   = local.tags
}

# S3 bucket policy for CloudFront OAC
resource "aws_s3_bucket_policy" "assets_cloudfront_oac" {
  bucket = module.s3_assets.bucket_id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowCloudFrontServicePrincipal"
        Effect = "Allow"
        Principal = {
          Service = "cloudfront.amazonaws.com"
        }
        Action   = "s3:GetObject"
        Resource = "${module.s3_assets.bucket_arn}/*"
        Condition = {
          StringEquals = {
            "AWS:SourceArn" = module.cloudfront.distribution_arn
          }
        }
      }
    ]
  })
}

module "waf" {
  count  = var.enable_waf ? 1 : 0
  source = "../../modules/waf"
  providers = {
    aws = aws.us_east_1
  }

  project        = var.project
  environment    = var.environment
  scope          = "CLOUDFRONT"
  rate_limit     = var.waf_rate_limit
  enable_logging = true
  tags           = local.tags
}

module "route53" {
  source = "../../modules/route53"

  project                            = var.project
  domain_name                        = var.domain_name
  create_hosted_zone                 = var.create_hosted_zone
  hosted_zone_id                     = var.hosted_zone_id
  cloudfront_domain_name             = module.cloudfront.distribution_domain_name
  cloudfront_hosted_zone_id          = module.cloudfront.distribution_hosted_zone_id
  enable_cloudfront_record           = true
  api_subdomain                      = "api"
  alb_dns_name                       = module.alb.alb_dns_name
  alb_zone_id                        = module.alb.alb_zone_id
  enable_api_record                  = true
  acm_certificate_validation_records = merge(module.acm.domain_validation_options, module.acm_cloudfront.domain_validation_options)
  tags                               = local.tags
}

module "cognito" {
  source = "../../modules/cognito"

  project                = var.project
  environment            = var.environment
  mfa_configuration      = var.cognito_mfa_configuration
  deletion_protection    = var.cognito_deletion_protection
  advanced_security_mode = var.cognito_advanced_security_mode
  callback_urls          = var.cognito_callback_urls
  logout_urls            = var.cognito_logout_urls
  tags                   = local.tags
}

module "iam" {
  source = "../../modules/iam"

  project                     = var.project
  environment                 = var.environment
  aws_region                  = var.aws_region
  create_github_oidc_provider = var.create_github_oidc_provider
  github_oidc_provider_arn    = var.github_oidc_provider_arn
  github_repository           = var.github_repository
  ecr_repository_arns         = [module.ecr_api.repository_arn]
  ecs_cluster_name            = module.ecs.cluster_name
  ecs_task_role_arns          = [module.ecs.execution_role_arn, module.ecs.task_role_arn]
  rds_resource_id             = module.aurora.cluster_resource_id
  rds_db_username             = var.database_app_username
  enable_rds_iam_auth         = true
  tags                        = local.tags
}
