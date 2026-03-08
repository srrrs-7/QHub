# Cognito Module
# Creates Cognito User Pool and App Client for authentication

# -----------------------------------------------------------------------------
# User Pool
# -----------------------------------------------------------------------------
resource "aws_cognito_user_pool" "main" {
  name = "${var.project}-${var.environment}-user-pool"

  # Username configuration
  username_attributes      = var.username_attributes
  auto_verified_attributes = var.auto_verified_attributes

  # Username case sensitivity
  username_configuration {
    case_sensitive = var.username_case_sensitive
  }

  # Password policy
  password_policy {
    minimum_length                   = var.password_minimum_length
    require_lowercase                = var.password_require_lowercase
    require_numbers                  = var.password_require_numbers
    require_symbols                  = var.password_require_symbols
    require_uppercase                = var.password_require_uppercase
    temporary_password_validity_days = var.temporary_password_validity_days
  }

  # MFA configuration
  mfa_configuration = var.mfa_configuration

  dynamic "software_token_mfa_configuration" {
    for_each = var.mfa_configuration != "OFF" ? [1] : []
    content {
      enabled = true
    }
  }

  # Account recovery
  account_recovery_setting {
    recovery_mechanism {
      name     = "verified_email"
      priority = 1
    }
    dynamic "recovery_mechanism" {
      for_each = contains(var.auto_verified_attributes, "phone_number") ? [1] : []
      content {
        name     = "verified_phone_number"
        priority = 2
      }
    }
  }

  # Email configuration
  email_configuration {
    email_sending_account = var.ses_email_identity != null ? "DEVELOPER" : "COGNITO_DEFAULT"
    source_arn            = var.ses_email_identity
    from_email_address    = var.from_email_address
  }

  # User attribute schema
  dynamic "schema" {
    for_each = var.schema_attributes
    content {
      name                     = schema.value.name
      attribute_data_type      = schema.value.attribute_data_type
      developer_only_attribute = lookup(schema.value, "developer_only_attribute", false)
      mutable                  = lookup(schema.value, "mutable", true)
      required                 = lookup(schema.value, "required", false)

      dynamic "string_attribute_constraints" {
        for_each = schema.value.attribute_data_type == "String" ? [1] : []
        content {
          min_length = lookup(schema.value, "min_length", 0)
          max_length = lookup(schema.value, "max_length", 2048)
        }
      }

      dynamic "number_attribute_constraints" {
        for_each = schema.value.attribute_data_type == "Number" ? [1] : []
        content {
          min_value = lookup(schema.value, "min_value", null)
          max_value = lookup(schema.value, "max_value", null)
        }
      }
    }
  }

  # Verification message
  verification_message_template {
    default_email_option = "CONFIRM_WITH_CODE"
    email_subject        = var.email_verification_subject
    email_message        = var.email_verification_message
  }

  # User pool add-ons
  user_pool_add_ons {
    advanced_security_mode = var.advanced_security_mode
  }

  # Deletion protection
  deletion_protection = var.deletion_protection ? "ACTIVE" : "INACTIVE"

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-user-pool"
  })
}

# -----------------------------------------------------------------------------
# User Pool Domain
# -----------------------------------------------------------------------------
resource "aws_cognito_user_pool_domain" "main" {
  domain          = var.custom_domain != null ? var.custom_domain : "${var.project}-${var.environment}"
  user_pool_id    = aws_cognito_user_pool.main.id
  certificate_arn = var.custom_domain != null ? var.custom_domain_certificate_arn : null
}

# -----------------------------------------------------------------------------
# User Pool Client
# -----------------------------------------------------------------------------
resource "aws_cognito_user_pool_client" "main" {
  name         = "${var.project}-${var.environment}-client"
  user_pool_id = aws_cognito_user_pool.main.id

  # Token configuration
  access_token_validity  = var.access_token_validity
  id_token_validity      = var.id_token_validity
  refresh_token_validity = var.refresh_token_validity

  token_validity_units {
    access_token  = "hours"
    id_token      = "hours"
    refresh_token = "days"
  }

  # OAuth configuration
  allowed_oauth_flows                  = var.allowed_oauth_flows
  allowed_oauth_flows_user_pool_client = length(var.allowed_oauth_flows) > 0
  allowed_oauth_scopes                 = var.allowed_oauth_scopes
  callback_urls                        = var.callback_urls
  logout_urls                          = var.logout_urls
  supported_identity_providers         = var.supported_identity_providers

  # Security
  generate_secret                               = var.generate_secret
  prevent_user_existence_errors                 = "ENABLED"
  enable_token_revocation                       = true
  enable_propagate_additional_user_context_data = false

  # Auth flows
  explicit_auth_flows = var.explicit_auth_flows

  # Read/Write attributes
  read_attributes  = var.read_attributes
  write_attributes = var.write_attributes
}

# -----------------------------------------------------------------------------
# Resource Server (for custom scopes)
# -----------------------------------------------------------------------------
resource "aws_cognito_resource_server" "main" {
  count = length(var.resource_server_scopes) > 0 ? 1 : 0

  identifier   = "https://${var.project}-${var.environment}.api"
  name         = "${var.project}-${var.environment}-api"
  user_pool_id = aws_cognito_user_pool.main.id

  dynamic "scope" {
    for_each = var.resource_server_scopes
    content {
      scope_name        = scope.value.name
      scope_description = scope.value.description
    }
  }
}

# -----------------------------------------------------------------------------
# User Pool Groups
# -----------------------------------------------------------------------------
resource "aws_cognito_user_group" "main" {
  for_each = var.user_groups

  name         = each.key
  user_pool_id = aws_cognito_user_pool.main.id
  description  = each.value.description
  precedence   = each.value.precedence
  role_arn     = lookup(each.value, "role_arn", null)
}
