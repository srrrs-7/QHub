variable "project" {
  description = "Project name"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, stg, prd)"
  type        = string
}

variable "username_attributes" {
  description = "Username attributes (email, phone_number)"
  type        = list(string)
  default     = ["email"]
}

variable "auto_verified_attributes" {
  description = "Attributes to auto-verify"
  type        = list(string)
  default     = ["email"]
}

variable "username_case_sensitive" {
  description = "Username case sensitivity"
  type        = bool
  default     = false
}

variable "password_minimum_length" {
  description = "Minimum password length"
  type        = number
  default     = 12
}

variable "password_require_lowercase" {
  description = "Require lowercase in password"
  type        = bool
  default     = true
}

variable "password_require_numbers" {
  description = "Require numbers in password"
  type        = bool
  default     = true
}

variable "password_require_symbols" {
  description = "Require symbols in password"
  type        = bool
  default     = true
}

variable "password_require_uppercase" {
  description = "Require uppercase in password"
  type        = bool
  default     = true
}

variable "temporary_password_validity_days" {
  description = "Temporary password validity in days"
  type        = number
  default     = 7
}

variable "mfa_configuration" {
  description = "MFA configuration (OFF, ON, OPTIONAL)"
  type        = string
  default     = "OPTIONAL"
}

variable "advanced_security_mode" {
  description = "Advanced security mode (OFF, AUDIT, ENFORCED)"
  type        = string
  default     = "AUDIT"
}

variable "deletion_protection" {
  description = "Enable deletion protection"
  type        = bool
  default     = true
}

variable "ses_email_identity" {
  description = "SES email identity ARN for sending emails"
  type        = string
  default     = null
}

variable "from_email_address" {
  description = "From email address"
  type        = string
  default     = null
}

variable "email_verification_subject" {
  description = "Email verification subject"
  type        = string
  default     = "Your verification code"
}

variable "email_verification_message" {
  description = "Email verification message"
  type        = string
  default     = "Your verification code is {####}"
}

variable "schema_attributes" {
  description = "Custom schema attributes"
  type = list(object({
    name                     = string
    attribute_data_type      = string
    developer_only_attribute = optional(bool)
    mutable                  = optional(bool)
    required                 = optional(bool)
    min_length               = optional(number)
    max_length               = optional(number)
    min_value                = optional(string)
    max_value                = optional(string)
  }))
  default = []
}

variable "custom_domain" {
  description = "Custom domain for the user pool"
  type        = string
  default     = null
}

variable "custom_domain_certificate_arn" {
  description = "ACM certificate ARN for custom domain"
  type        = string
  default     = null
}

variable "access_token_validity" {
  description = "Access token validity in hours"
  type        = number
  default     = 1
}

variable "id_token_validity" {
  description = "ID token validity in hours"
  type        = number
  default     = 1
}

variable "refresh_token_validity" {
  description = "Refresh token validity in days"
  type        = number
  default     = 30
}

variable "allowed_oauth_flows" {
  description = "Allowed OAuth flows"
  type        = list(string)
  default     = ["code"]
}

variable "allowed_oauth_scopes" {
  description = "Allowed OAuth scopes"
  type        = list(string)
  default     = ["email", "openid", "profile"]
}

variable "callback_urls" {
  description = "Callback URLs"
  type        = list(string)
  default     = []
}

variable "logout_urls" {
  description = "Logout URLs"
  type        = list(string)
  default     = []
}

variable "supported_identity_providers" {
  description = "Supported identity providers"
  type        = list(string)
  default     = ["COGNITO"]
}

variable "generate_secret" {
  description = "Generate client secret"
  type        = bool
  default     = false
}

variable "explicit_auth_flows" {
  description = "Explicit auth flows"
  type        = list(string)
  default = [
    "ALLOW_REFRESH_TOKEN_AUTH",
    "ALLOW_USER_SRP_AUTH"
  ]
}

variable "read_attributes" {
  description = "Read attributes"
  type        = list(string)
  default     = ["email", "email_verified", "name"]
}

variable "write_attributes" {
  description = "Write attributes"
  type        = list(string)
  default     = ["email", "name"]
}

variable "resource_server_scopes" {
  description = "Resource server scopes"
  type = list(object({
    name        = string
    description = string
  }))
  default = []
}

variable "user_groups" {
  description = "User groups"
  type = map(object({
    description = string
    precedence  = number
    role_arn    = optional(string)
  }))
  default = {}
}

variable "tags" {
  description = "Additional tags for resources"
  type        = map(string)
  default     = {}
}
