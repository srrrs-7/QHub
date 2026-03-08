variable "project" {
  description = "Project name"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, stg, prd)"
  type        = string
}

variable "scope" {
  description = "Scope of the WAF (REGIONAL or CLOUDFRONT)"
  type        = string
  default     = "REGIONAL"
}

variable "rate_limit" {
  description = "Rate limit per 5 minutes (0 to disable)"
  type        = number
  default     = 2000
}

variable "ip_block_list" {
  description = "List of IP addresses to block (CIDR notation)"
  type        = list(string)
  default     = null
}

variable "ip_allow_list" {
  description = "List of IP addresses to allow (CIDR notation)"
  type        = list(string)
  default     = null
}

variable "common_ruleset_excluded_rules" {
  description = "Rules to exclude from AWS Common Rule Set"
  type        = list(string)
  default     = []
}

variable "enable_logging" {
  description = "Enable WAF logging"
  type        = bool
  default     = true
}

variable "log_retention_days" {
  description = "CloudWatch log retention in days"
  type        = number
  default     = 30
}

variable "redacted_fields" {
  description = "Fields to redact in logs"
  type = list(object({
    type = string
    name = string
  }))
  default = [
    {
      type = "single_header"
      name = "authorization"
    }
  ]
}

variable "tags" {
  description = "Additional tags for resources"
  type        = map(string)
  default     = {}
}
