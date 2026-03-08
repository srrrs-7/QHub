# Aurora PostgreSQL Module
# Creates Aurora PostgreSQL Serverless v2 cluster with IAM authentication

# -----------------------------------------------------------------------------
# Random password for master user (if not provided)
# -----------------------------------------------------------------------------
resource "random_password" "master" {
  count = var.master_password == null ? 1 : 0

  length           = 32
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

# -----------------------------------------------------------------------------
# Secrets Manager for master password
# -----------------------------------------------------------------------------
resource "aws_secretsmanager_secret" "master_password" {
  name                    = "${var.project}-${var.environment}-aurora-master-password"
  description             = "Master password for Aurora PostgreSQL cluster"
  recovery_window_in_days = var.secret_recovery_window_days

  tags = var.tags
}

resource "aws_secretsmanager_secret_version" "master_password" {
  secret_id = aws_secretsmanager_secret.master_password.id
  secret_string = jsonencode({
    username = var.master_username
    password = var.master_password != null ? var.master_password : random_password.master[0].result
    host     = aws_rds_cluster.main.endpoint
    port     = aws_rds_cluster.main.port
    dbname   = var.database_name
  })
}

# -----------------------------------------------------------------------------
# DB Cluster Parameter Group
# -----------------------------------------------------------------------------
resource "aws_rds_cluster_parameter_group" "main" {
  name        = "${var.project}-${var.environment}-aurora-pg-cluster-params"
  family      = "aurora-postgresql${var.engine_version_major}"
  description = "Aurora PostgreSQL cluster parameter group for ${var.project}-${var.environment}"

  dynamic "parameter" {
    for_each = var.cluster_parameters
    content {
      name         = parameter.value.name
      value        = parameter.value.value
      apply_method = lookup(parameter.value, "apply_method", "immediate")
    }
  }

  tags = var.tags
}

# -----------------------------------------------------------------------------
# DB Instance Parameter Group
# -----------------------------------------------------------------------------
resource "aws_db_parameter_group" "main" {
  name        = "${var.project}-${var.environment}-aurora-pg-instance-params"
  family      = "aurora-postgresql${var.engine_version_major}"
  description = "Aurora PostgreSQL instance parameter group for ${var.project}-${var.environment}"

  dynamic "parameter" {
    for_each = var.instance_parameters
    content {
      name         = parameter.value.name
      value        = parameter.value.value
      apply_method = lookup(parameter.value, "apply_method", "immediate")
    }
  }

  tags = var.tags
}

# -----------------------------------------------------------------------------
# Aurora Cluster
# -----------------------------------------------------------------------------
resource "aws_rds_cluster" "main" {
  cluster_identifier = "${var.project}-${var.environment}-aurora-cluster"
  engine             = "aurora-postgresql"
  engine_mode        = "provisioned"
  engine_version     = var.engine_version
  database_name      = var.database_name
  master_username    = var.master_username
  master_password    = var.master_password != null ? var.master_password : random_password.master[0].result

  db_subnet_group_name            = var.db_subnet_group_name
  vpc_security_group_ids          = var.security_group_ids
  db_cluster_parameter_group_name = aws_rds_cluster_parameter_group.main.name

  # Serverless v2 configuration
  serverlessv2_scaling_configuration {
    min_capacity = var.serverless_min_capacity
    max_capacity = var.serverless_max_capacity
  }

  # Storage encryption
  storage_encrypted = true
  kms_key_id        = var.kms_key_arn

  # IAM authentication
  iam_database_authentication_enabled = var.enable_iam_auth

  # Backup configuration
  backup_retention_period   = var.backup_retention_period
  preferred_backup_window   = var.preferred_backup_window
  copy_tags_to_snapshot     = true
  deletion_protection       = var.deletion_protection
  skip_final_snapshot       = var.skip_final_snapshot
  final_snapshot_identifier = var.skip_final_snapshot ? null : "${var.project}-${var.environment}-aurora-final-snapshot"

  # Maintenance
  preferred_maintenance_window = var.preferred_maintenance_window
  apply_immediately            = var.apply_immediately

  # Performance Insights
  enabled_cloudwatch_logs_exports = var.enabled_cloudwatch_logs_exports

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-aurora-cluster"
  })

  lifecycle {
    ignore_changes = [master_password]
  }
}

# -----------------------------------------------------------------------------
# Aurora Instances (Serverless v2)
# -----------------------------------------------------------------------------
resource "aws_rds_cluster_instance" "main" {
  count = var.instance_count

  identifier              = "${var.project}-${var.environment}-aurora-instance-${count.index + 1}"
  cluster_identifier      = aws_rds_cluster.main.id
  instance_class          = "db.serverless"
  engine                  = aws_rds_cluster.main.engine
  engine_version          = aws_rds_cluster.main.engine_version
  db_parameter_group_name = aws_db_parameter_group.main.name

  # Performance Insights
  performance_insights_enabled          = var.performance_insights_enabled
  performance_insights_retention_period = var.performance_insights_enabled ? var.performance_insights_retention_period : null
  performance_insights_kms_key_id       = var.performance_insights_enabled ? var.kms_key_arn : null

  # Monitoring
  monitoring_interval = var.enhanced_monitoring_interval
  monitoring_role_arn = var.enhanced_monitoring_interval > 0 ? aws_iam_role.rds_monitoring[0].arn : null

  # Auto minor version upgrade
  auto_minor_version_upgrade = var.auto_minor_version_upgrade

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-aurora-instance-${count.index + 1}"
  })
}

# -----------------------------------------------------------------------------
# Enhanced Monitoring IAM Role
# -----------------------------------------------------------------------------
resource "aws_iam_role" "rds_monitoring" {
  count = var.enhanced_monitoring_interval > 0 ? 1 : 0

  name = "${var.project}-${var.environment}-rds-monitoring-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "monitoring.rds.amazonaws.com"
        }
      }
    ]
  })

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "rds_monitoring" {
  count = var.enhanced_monitoring_interval > 0 ? 1 : 0

  role       = aws_iam_role.rds_monitoring[0].name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
}

# -----------------------------------------------------------------------------
# IAM User for Database Authentication
# -----------------------------------------------------------------------------
resource "aws_rds_cluster_role_association" "main" {
  count = var.enable_iam_auth && var.iam_roles != null ? length(var.iam_roles) : 0

  db_cluster_identifier = aws_rds_cluster.main.id
  feature_name          = "s3Import"
  role_arn              = var.iam_roles[count.index]
}
