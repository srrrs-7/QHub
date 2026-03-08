# RDS PostgreSQL Database Users
# Creates application and migration users with appropriate permissions
#
# Note: Requires PostgreSQL provider to be configured in the root module
# with access to the RDS instance

terraform {
  required_providers {
    postgresql = {
      source  = "cyrilgdn/postgresql"
      version = ">= 1.22"
    }
  }
}

# -----------------------------------------------------------------------------
# Random passwords for database users
# -----------------------------------------------------------------------------
resource "random_password" "api_user" {
  count = var.create_db_users ? 1 : 0

  length           = 32
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

resource "random_password" "migrate_user" {
  count = var.create_db_users ? 1 : 0

  length           = 32
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

# -----------------------------------------------------------------------------
# Secrets Manager for user passwords
# -----------------------------------------------------------------------------
resource "aws_secretsmanager_secret" "api_user" {
  count = var.create_db_users ? 1 : 0

  name                    = "${var.project}-${var.environment}-rds-api-user"
  description             = "Credentials for RDS PostgreSQL API user"
  recovery_window_in_days = var.deletion_protection ? 7 : 0

  tags = var.tags
}

resource "aws_secretsmanager_secret_version" "api_user" {
  count = var.create_db_users ? 1 : 0

  secret_id = aws_secretsmanager_secret.api_user[0].id
  secret_string = jsonencode({
    username = var.api_db_username
    password = random_password.api_user[0].result
    host     = aws_db_instance.main.address
    port     = aws_db_instance.main.port
    dbname   = var.database_name
  })
}

resource "aws_secretsmanager_secret" "migrate_user" {
  count = var.create_db_users ? 1 : 0

  name                    = "${var.project}-${var.environment}-rds-migrate-user"
  description             = "Credentials for RDS PostgreSQL migration user"
  recovery_window_in_days = var.deletion_protection ? 7 : 0

  tags = var.tags
}

resource "aws_secretsmanager_secret_version" "migrate_user" {
  count = var.create_db_users ? 1 : 0

  secret_id = aws_secretsmanager_secret.migrate_user[0].id
  secret_string = jsonencode({
    username = var.migrate_db_username
    password = random_password.migrate_user[0].result
    host     = aws_db_instance.main.address
    port     = aws_db_instance.main.port
    dbname   = var.database_name
  })
}

# -----------------------------------------------------------------------------
# PostgreSQL Roles (Users)
# -----------------------------------------------------------------------------
resource "postgresql_role" "api_user" {
  count = var.create_db_users ? 1 : 0

  name     = var.api_db_username
  login    = true
  password = random_password.api_user[0].result

  depends_on = [aws_db_instance.main]
}

resource "postgresql_role" "migrate_user" {
  count = var.create_db_users ? 1 : 0

  name     = var.migrate_db_username
  login    = true
  password = random_password.migrate_user[0].result

  # Migration user can create objects
  create_role     = false
  create_database = false
  superuser       = false

  depends_on = [aws_db_instance.main]
}

# -----------------------------------------------------------------------------
# PostgreSQL Grants
# -----------------------------------------------------------------------------

# API user: Read/Write on tables (SELECT, INSERT, UPDATE, DELETE)
resource "postgresql_grant" "api_tables" {
  count = var.create_db_users ? 1 : 0

  database    = var.database_name
  role        = postgresql_role.api_user[0].name
  schema      = "public"
  object_type = "table"
  privileges  = ["SELECT", "INSERT", "UPDATE", "DELETE"]

  depends_on = [postgresql_role.api_user]
}

# API user: Usage on sequences (for auto-increment columns)
resource "postgresql_grant" "api_sequences" {
  count = var.create_db_users ? 1 : 0

  database    = var.database_name
  role        = postgresql_role.api_user[0].name
  schema      = "public"
  object_type = "sequence"
  privileges  = ["USAGE", "SELECT"]

  depends_on = [postgresql_role.api_user]
}

# API user: Usage on schema
resource "postgresql_grant" "api_schema" {
  count = var.create_db_users ? 1 : 0

  database    = var.database_name
  role        = postgresql_role.api_user[0].name
  schema      = "public"
  object_type = "schema"
  privileges  = ["USAGE"]

  depends_on = [postgresql_role.api_user]
}

# Migrate user: Full control on tables
resource "postgresql_grant" "migrate_tables" {
  count = var.create_db_users ? 1 : 0

  database    = var.database_name
  role        = postgresql_role.migrate_user[0].name
  schema      = "public"
  object_type = "table"
  privileges  = ["ALL"]

  depends_on = [postgresql_role.migrate_user]
}

# Migrate user: Full control on sequences
resource "postgresql_grant" "migrate_sequences" {
  count = var.create_db_users ? 1 : 0

  database    = var.database_name
  role        = postgresql_role.migrate_user[0].name
  schema      = "public"
  object_type = "sequence"
  privileges  = ["ALL"]

  depends_on = [postgresql_role.migrate_user]
}

# Migrate user: Full control on schema (for creating tables)
resource "postgresql_grant" "migrate_schema" {
  count = var.create_db_users ? 1 : 0

  database    = var.database_name
  role        = postgresql_role.migrate_user[0].name
  schema      = "public"
  object_type = "schema"
  privileges  = ["CREATE", "USAGE"]

  depends_on = [postgresql_role.migrate_user]
}

# -----------------------------------------------------------------------------
# Default Privileges (for future tables created by migrate user)
# -----------------------------------------------------------------------------
resource "postgresql_default_privileges" "api_tables" {
  count = var.create_db_users ? 1 : 0

  database = var.database_name
  role     = postgresql_role.api_user[0].name
  owner    = postgresql_role.migrate_user[0].name
  schema   = "public"

  object_type = "table"
  privileges  = ["SELECT", "INSERT", "UPDATE", "DELETE"]

  depends_on = [postgresql_role.api_user, postgresql_role.migrate_user]
}

resource "postgresql_default_privileges" "api_sequences" {
  count = var.create_db_users ? 1 : 0

  database = var.database_name
  role     = postgresql_role.api_user[0].name
  owner    = postgresql_role.migrate_user[0].name
  schema   = "public"

  object_type = "sequence"
  privileges  = ["USAGE", "SELECT"]

  depends_on = [postgresql_role.api_user, postgresql_role.migrate_user]
}
