# Development Environment
# Main configuration file that orchestrates all modules

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
    postgresql = {
      source  = "cyrilgdn/postgresql"
      version = "~> 1.22"
    }
  }

  # Dev環境: ローカルバックエンド（状態ファイルはローカルに保存）
  # チーム開発やCI/CDが必要な場合はS3バックエンドに変更してください
  #
  # backend "s3" {
  #   bucket         = "your-terraform-state-bucket"
  #   key            = "dev/terraform.tfstate"
  #   region         = "ap-northeast-1"
  #   encrypt        = true
  #   dynamodb_table = "terraform-lock"
  # }
}

provider "aws" {
  region  = var.aws_region
  profile = var.aws_profile

  default_tags {
    tags = {
      Project     = var.project
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}

# Provider for ACM certificates (must be us-east-1 for CloudFront)
provider "aws" {
  alias   = "us_east_1"
  region  = "us-east-1"
  profile = var.aws_profile

  default_tags {
    tags = {
      Project     = var.project
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}

# PostgreSQL provider for database user management
# Note: This requires network access to the database from the Terraform execution environment
# For CI/CD, consider using a bastion host or VPN connection
#
# IMPORTANT: The PostgreSQL provider configuration uses data sources which creates
# a chicken-and-egg problem. The database must exist before users can be created.
# Run terraform apply in two steps:
#   1. First apply without create_db_users=true to create the database
#   2. Then apply with create_db_users=true to create the users
provider "postgresql" {
  host     = var.create_db_users ? local.db_endpoint : "localhost"
  port     = var.create_db_users ? local.db_port : 5432
  database = var.database_name
  username = var.database_master_username
  password = var.create_db_users ? jsondecode(data.aws_secretsmanager_secret_version.db_master[0].secret_string)["password"] : ""

  # SSL configuration for AWS RDS/Aurora
  sslmode = var.create_db_users ? "require" : "disable"

  # PostgreSQL provider doesn't need superuser for creating roles
  superuser = false
}

# Fetch master password from Secrets Manager for PostgreSQL provider
data "aws_secretsmanager_secret_version" "db_master" {
  count = var.create_db_users ? 1 : 0

  secret_id = local.db_master_secret_arn
}

locals {
  tags = {
    Project     = var.project
    Environment = var.environment
  }
}

# -----------------------------------------------------------------------------
# VPC
# -----------------------------------------------------------------------------
module "vpc" {
  source = "../../modules/vpc"

  project     = var.project
  environment = var.environment
  aws_region  = var.aws_region
  vpc_cidr    = var.vpc_cidr
  az_count    = var.az_count

  enable_nat_gateway         = var.enable_nat_gateway
  single_nat_gateway         = var.single_nat_gateway
  enable_flow_logs           = var.enable_flow_logs
  enable_vpc_endpoints       = var.enable_vpc_endpoints
  enable_interface_endpoints = var.enable_interface_endpoints
  enable_ssm_endpoints       = var.enable_ssm_endpoints

  tags = local.tags
}

# -----------------------------------------------------------------------------
# Security Groups
# -----------------------------------------------------------------------------
module "security_groups" {
  source = "../../modules/security-groups"

  project     = var.project
  environment = var.environment
  vpc_id      = module.vpc.vpc_id
  app_port    = var.app_port

  tags = local.tags
}

# -----------------------------------------------------------------------------
# ECR
# -----------------------------------------------------------------------------
module "ecr_api" {
  source = "../../modules/ecr"

  project         = var.project
  environment     = var.environment
  repository_name = "api"
  scan_on_push    = true

  tags = local.tags
}

module "ecr_migrate" {
  source = "../../modules/ecr"

  project         = var.project
  environment     = var.environment
  repository_name = "migrate"
  scan_on_push    = true

  tags = local.tags
}

# -----------------------------------------------------------------------------
# Database (Aurora or RDS based on database_type)
# -----------------------------------------------------------------------------

# Aurora PostgreSQL (when database_type = "aurora")
module "aurora" {
  count  = var.database_type == "aurora" ? 1 : 0
  source = "../../modules/aurora"

  project              = var.project
  environment          = var.environment
  database_name        = var.database_name
  master_username      = var.database_master_username
  engine_version       = var.aurora_engine_version
  engine_version_major = var.aurora_engine_version_major

  instance_count          = var.aurora_instance_count
  serverless_min_capacity = var.aurora_min_capacity
  serverless_max_capacity = var.aurora_max_capacity

  db_subnet_group_name = module.vpc.db_subnet_group_name
  security_group_ids   = [module.security_groups.aurora_security_group_id]

  enable_iam_auth         = true
  backup_retention_period = var.aurora_backup_retention_period
  deletion_protection     = var.aurora_deletion_protection
  skip_final_snapshot     = var.aurora_skip_final_snapshot

  # Database users
  create_db_users     = var.create_db_users
  api_db_username     = var.api_db_username
  migrate_db_username = var.migrate_db_username

  tags = local.tags
}

# RDS PostgreSQL (when database_type = "rds") - Cost optimized for dev
module "rds" {
  count  = var.database_type == "rds" ? 1 : 0
  source = "../../modules/rds"

  project         = var.project
  environment     = var.environment
  database_name   = var.database_name
  master_username = var.database_master_username
  engine_version  = var.rds_engine_version

  instance_class    = var.rds_instance_class
  allocated_storage = var.rds_allocated_storage

  db_subnet_group_name = module.vpc.db_subnet_group_name
  security_group_ids   = [module.security_groups.aurora_security_group_id]

  enable_iam_auth         = true
  backup_retention_period = var.rds_backup_retention_period
  deletion_protection     = false
  skip_final_snapshot     = true

  # Cost savings: disable optional features
  enable_performance_insights = false
  enable_cloudwatch_logs      = false

  # Database users
  create_db_users     = var.create_db_users
  api_db_username     = var.api_db_username
  migrate_db_username = var.migrate_db_username

  tags = local.tags
}

# Local values for database outputs (works with either Aurora or RDS)
locals {
  db_endpoint          = var.database_type == "aurora" ? module.aurora[0].cluster_endpoint : module.rds[0].endpoint
  db_port              = var.database_type == "aurora" ? module.aurora[0].cluster_port : module.rds[0].port
  db_name              = var.database_type == "aurora" ? module.aurora[0].database_name : module.rds[0].database_name
  db_resource_id       = var.database_type == "aurora" ? module.aurora[0].cluster_resource_id : module.rds[0].resource_id
  db_master_secret_arn = var.database_type == "aurora" ? module.aurora[0].master_password_secret_arn : module.rds[0].secret_arn
}

# -----------------------------------------------------------------------------
# ACM Certificates (only when domain is configured)
# -----------------------------------------------------------------------------
module "acm" {
  count  = var.domain_name != null ? 1 : 0
  source = "../../modules/acm"

  domain_name               = var.domain_name
  subject_alternative_names = ["*.${var.domain_name}"]
  validate_certificate      = false # Validation done via Route53 module

  tags = local.tags
}

# ACM for CloudFront (must be in us-east-1)
module "acm_cloudfront" {
  count  = var.domain_name != null ? 1 : 0
  source = "../../modules/acm"
  providers = {
    aws = aws.us_east_1
  }

  domain_name               = var.domain_name
  subject_alternative_names = ["*.${var.domain_name}"]
  validate_certificate      = false

  tags = local.tags
}

# -----------------------------------------------------------------------------
# ALB
# -----------------------------------------------------------------------------
module "alb" {
  source = "../../modules/alb"

  project            = var.project
  environment        = var.environment
  vpc_id             = module.vpc.vpc_id
  subnet_ids         = module.vpc.public_subnet_ids
  security_group_ids = [module.security_groups.alb_security_group_id]
  certificate_arn    = var.domain_name != null ? module.acm[0].certificate_arn : null

  target_port       = var.app_port
  health_check_path = var.health_check_path

  deletion_protection = var.alb_deletion_protection

  tags = local.tags
}

# -----------------------------------------------------------------------------
# ECS
# -----------------------------------------------------------------------------
module "ecs" {
  source = "../../modules/ecs"

  project         = var.project
  environment     = var.environment
  aws_region      = var.aws_region
  service_name    = "api"
  container_name  = "api"
  container_image = "${module.ecr_api.repository_url}:latest"
  container_port  = var.app_port

  task_cpu    = var.ecs_task_cpu
  task_memory = var.ecs_task_memory

  desired_count      = var.ecs_desired_count
  subnet_ids         = module.vpc.private_subnet_ids
  security_group_ids = [module.security_groups.ecs_security_group_id]
  target_group_arn   = module.alb.api_target_group_arn

  enable_autoscaling       = var.ecs_enable_autoscaling
  autoscaling_min_capacity = var.ecs_min_capacity
  autoscaling_max_capacity = var.ecs_max_capacity

  # Fargate Spot for cost savings (up to 70% cheaper)
  enable_fargate_spot = var.ecs_use_fargate_spot
  fargate_weight      = var.ecs_use_fargate_spot ? 0 : 1
  fargate_spot_weight = var.ecs_use_fargate_spot ? 1 : 0
  fargate_base_count  = 0

  # Database connection (works with both Aurora and RDS)
  rds_resource_id     = local.db_resource_id
  rds_db_username     = var.database_app_username
  enable_rds_iam_auth = true

  environment_variables = [
    {
      name  = "DB_HOST"
      value = local.db_endpoint
    },
    {
      name  = "DB_PORT"
      value = tostring(local.db_port)
    },
    {
      name  = "DB_NAME"
      value = local.db_name
    },
    {
      name  = "DB_USER"
      value = var.database_app_username
    },
    {
      name  = "ENVIRONMENT"
      value = var.environment
    }
  ]

  enable_execute_command = var.ecs_enable_execute_command

  # Cost savings: disable Container Insights
  enable_container_insights = false

  tags = local.tags
}

# -----------------------------------------------------------------------------
# ECS Job (Database Migration)
# -----------------------------------------------------------------------------
module "ecs_migrate" {
  source = "../../modules/ecs-job"

  project         = var.project
  environment     = var.environment
  aws_region      = var.aws_region
  job_name        = "migrate"
  container_name  = "migrate"
  container_image = "${module.ecr_migrate.repository_url}:latest"

  task_cpu    = 256
  task_memory = 512

  # Database connection
  environment_variables = [
    {
      name  = "DB_HOST"
      value = local.db_endpoint
    },
    {
      name  = "DB_PORT"
      value = tostring(local.db_port)
    },
    {
      name  = "DB_DBNAME"
      value = local.db_name
    },
    {
      name  = "DB_USERNAME"
      value = var.database_app_username
    }
  ]

  # Inject DB password from Secrets Manager
  secrets = var.database_type == "rds" ? [
    {
      name      = "DB_PASSWORD"
      valueFrom = "${module.rds[0].secret_arn}:password::"
    }
  ] : []

  db_secret_arn = var.database_type == "rds" ? module.rds[0].secret_arn : null

  # RDS IAM auth (for Aurora)
  enable_rds_iam_auth = var.database_type == "aurora"
  rds_resource_id     = var.database_type == "aurora" ? local.db_resource_id : null
  rds_db_username     = var.database_type == "aurora" ? var.database_app_username : null

  tags = local.tags
}

# -----------------------------------------------------------------------------
# S3 (Static Assets)
# -----------------------------------------------------------------------------
module "s3_assets" {
  source = "../../modules/s3"

  bucket_name         = "${var.project}-${var.environment}-assets"
  enable_versioning   = true
  block_public_access = true

  # CORS: CloudFrontドメインまたは独自ドメインからのアクセスを許可
  cors_rules = var.domain_name != null ? [
    {
      allowed_headers = ["*"]
      allowed_methods = ["GET", "HEAD"]
      allowed_origins = ["https://${var.domain_name}"]
    }
  ] : []

  tags = local.tags
}

# -----------------------------------------------------------------------------
# CloudFront
# -----------------------------------------------------------------------------
module "cloudfront" {
  source = "../../modules/cloudfront"

  project     = var.project
  environment = var.environment

  s3_origin_domain_name  = module.s3_assets.bucket_regional_domain_name
  alb_origin_domain_name = module.alb.alb_dns_name
  enable_s3_origin       = true

  # 独自ドメインがある場合のみaliasesとcertificateを設定
  aliases         = var.domain_name != null ? [var.domain_name, "www.${var.domain_name}"] : []
  certificate_arn = var.domain_name != null ? module.acm_cloudfront[0].certificate_arn : null

  price_class = var.cloudfront_price_class
  web_acl_id  = var.enable_waf ? module.waf[0].web_acl_arn : null

  custom_error_responses = [
    {
      error_code         = 404
      response_code      = 200
      response_page_path = "/index.html"
    }
  ]

  tags = local.tags
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

# -----------------------------------------------------------------------------
# WAF
# -----------------------------------------------------------------------------
module "waf" {
  count  = var.enable_waf ? 1 : 0
  source = "../../modules/waf"

  project     = var.project
  environment = var.environment
  scope       = "CLOUDFRONT"

  rate_limit     = var.waf_rate_limit
  enable_logging = true

  tags = local.tags

  providers = {
    aws = aws.us_east_1
  }
}

# -----------------------------------------------------------------------------
# Route53 (only when domain is configured)
# -----------------------------------------------------------------------------
module "route53" {
  count  = var.domain_name != null ? 1 : 0
  source = "../../modules/route53"

  project     = var.project
  domain_name = var.domain_name

  create_hosted_zone = var.create_hosted_zone
  hosted_zone_id     = var.hosted_zone_id

  cloudfront_domain_name    = module.cloudfront.distribution_domain_name
  cloudfront_hosted_zone_id = module.cloudfront.distribution_hosted_zone_id
  enable_cloudfront_record  = true

  api_subdomain     = "api"
  alb_dns_name      = module.alb.alb_dns_name
  alb_zone_id       = module.alb.alb_zone_id
  enable_api_record = true

  acm_certificate_validation_records = merge(
    module.acm[0].domain_validation_options,
    module.acm_cloudfront[0].domain_validation_options
  )

  tags = local.tags
}

# -----------------------------------------------------------------------------
# Cognito
# -----------------------------------------------------------------------------
module "cognito" {
  source = "../../modules/cognito"

  project     = var.project
  environment = var.environment

  mfa_configuration      = var.cognito_mfa_configuration
  deletion_protection    = var.cognito_deletion_protection
  advanced_security_mode = var.cognito_advanced_security_mode

  callback_urls = var.cognito_callback_urls
  logout_urls   = var.cognito_logout_urls

  tags = local.tags
}

# -----------------------------------------------------------------------------
# IAM (GitHub Actions)
# -----------------------------------------------------------------------------
module "iam" {
  source = "../../modules/iam"

  project     = var.project
  environment = var.environment
  aws_region  = var.aws_region

  create_github_oidc_provider = var.create_github_oidc_provider
  github_oidc_provider_arn    = var.github_oidc_provider_arn
  github_repository           = var.github_repository

  ecr_repository_arns = [
    module.ecr_api.repository_arn,
    module.ecr_migrate.repository_arn
  ]
  ecs_cluster_name = module.ecs.cluster_name
  ecs_task_role_arns = [
    module.ecs.execution_role_arn,
    module.ecs.task_role_arn,
    module.ecs_migrate.execution_role_arn,
    module.ecs_migrate.task_role_arn
  ]

  # ECS Job (migration) permissions
  ecs_job_task_definition_arns = [module.ecs_migrate.task_definition_arn]

  rds_resource_id     = local.db_resource_id
  rds_db_username     = var.database_app_username
  enable_rds_iam_auth = true

  tags = local.tags
}

# -----------------------------------------------------------------------------
# Bastion Host (Session Manager)
# -----------------------------------------------------------------------------
module "bastion" {
  count  = var.enable_bastion ? 1 : 0
  source = "../../modules/bastion"

  project     = var.project
  environment = var.environment
  aws_region  = var.aws_region

  vpc_id    = module.vpc.vpc_id
  subnet_id = module.vpc.private_subnet_ids[0]

  instance_type         = var.bastion_instance_type
  db_security_group_ids = [module.security_groups.aurora_security_group_id]

  # RDS IAM auth
  enable_rds_iam_auth = true
  rds_resource_id     = local.db_resource_id

  # Allow bastion to read DB credentials from Secrets Manager
  secrets_arns = compact([
    local.db_master_secret_arn,
    var.database_type == "aurora" && var.create_db_users ? module.aurora[0].api_user_secret_arn : null,
    var.database_type == "aurora" && var.create_db_users ? module.aurora[0].migrate_user_secret_arn : null,
    var.database_type == "rds" && var.create_db_users ? module.rds[0].api_user_secret_arn : null,
    var.database_type == "rds" && var.create_db_users ? module.rds[0].migrate_user_secret_arn : null,
  ])

  tags = local.tags
}
