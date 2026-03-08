variable "project" {
  description = "Project name"
  type        = string
}

variable "domain_name" {
  description = "Domain name"
  type        = string
}

variable "subdomain" {
  description = "Subdomain (e.g., www, app)"
  type        = string
  default     = null
}

variable "api_subdomain" {
  description = "API subdomain (e.g., api)"
  type        = string
  default     = null
}

variable "create_hosted_zone" {
  description = "Create a new hosted zone"
  type        = bool
  default     = false
}

variable "hosted_zone_id" {
  description = "Existing hosted zone ID (required if create_hosted_zone is false)"
  type        = string
  default     = null
}

variable "cloudfront_domain_name" {
  description = "CloudFront distribution domain name"
  type        = string
  default     = null
}

variable "enable_cloudfront_record" {
  description = "Enable CloudFront DNS records (set to true when CloudFront is created)"
  type        = bool
  default     = false
}

variable "cloudfront_hosted_zone_id" {
  description = "CloudFront hosted zone ID"
  type        = string
  default     = "Z2FDTNDATAQYW2" # Global CloudFront hosted zone ID
}

variable "alb_dns_name" {
  description = "ALB DNS name"
  type        = string
  default     = null
}

variable "enable_api_record" {
  description = "Enable API subdomain record (set to true when ALB is created)"
  type        = bool
  default     = false
}

variable "alb_zone_id" {
  description = "ALB hosted zone ID"
  type        = string
  default     = null
}

variable "acm_certificate_validation_records" {
  description = "ACM certificate validation records"
  type = map(object({
    name   = string
    type   = string
    record = string
  }))
  default = {}
}

variable "additional_records" {
  description = "Additional DNS records"
  type = map(object({
    name    = string
    type    = string
    ttl     = optional(number)
    records = optional(list(string))
    alias = optional(object({
      name                   = string
      zone_id                = string
      evaluate_target_health = optional(bool)
    }))
  }))
  default = {}
}

variable "tags" {
  description = "Additional tags for resources"
  type        = map(string)
  default     = {}
}
