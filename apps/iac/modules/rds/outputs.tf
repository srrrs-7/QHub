output "instance_id" {
  description = "RDS instance ID"
  value       = aws_db_instance.main.id
}

output "instance_arn" {
  description = "RDS instance ARN"
  value       = aws_db_instance.main.arn
}

output "endpoint" {
  description = "RDS instance endpoint"
  value       = aws_db_instance.main.address
}

output "port" {
  description = "RDS instance port"
  value       = aws_db_instance.main.port
}

output "database_name" {
  description = "Database name"
  value       = aws_db_instance.main.db_name
}

output "master_username" {
  description = "Master username"
  value       = aws_db_instance.main.username
}

output "resource_id" {
  description = "RDS resource ID (for IAM auth)"
  value       = aws_db_instance.main.resource_id
}

output "secret_arn" {
  description = "Secrets Manager secret ARN for DB credentials"
  value       = aws_secretsmanager_secret.db_password.arn
}

# -----------------------------------------------------------------------------
# Database Users Outputs
# -----------------------------------------------------------------------------
output "api_user_secret_arn" {
  description = "Secrets Manager secret ARN for API user credentials"
  value       = var.create_db_users ? aws_secretsmanager_secret.api_user[0].arn : null
}

output "migrate_user_secret_arn" {
  description = "Secrets Manager secret ARN for migrate user credentials"
  value       = var.create_db_users ? aws_secretsmanager_secret.migrate_user[0].arn : null
}

output "api_db_username" {
  description = "API database username"
  value       = var.create_db_users ? var.api_db_username : null
}

output "migrate_db_username" {
  description = "Migration database username"
  value       = var.create_db_users ? var.migrate_db_username : null
}
