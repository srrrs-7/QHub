variable "project" {
  description = "Project name"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, stg, prd)"
  type        = string
}

variable "s3_origin_domain_name" {
  description = "S3 bucket regional domain name for origin"
  type        = string
  default     = null
}

variable "enable_s3_origin" {
  description = "Enable S3 origin (set to true when S3 bucket is created)"
  type        = bool
  default     = false
}

variable "alb_origin_domain_name" {
  description = "ALB domain name for origin"
  type        = string
  default     = null
}

variable "alb_custom_headers" {
  description = "Custom headers to send to ALB origin"
  type = list(object({
    name  = string
    value = string
  }))
  default = []
}

variable "aliases" {
  description = "Alternate domain names (CNAMEs)"
  type        = list(string)
  default     = []
}

variable "certificate_arn" {
  description = "ACM certificate ARN (must be in us-east-1)"
  type        = string
  default     = null
}

variable "default_root_object" {
  description = "Default root object"
  type        = string
  default     = "index.html"
}

variable "price_class" {
  description = "CloudFront price class"
  type        = string
  default     = "PriceClass_200"
}

variable "web_acl_id" {
  description = "WAF Web ACL ID"
  type        = string
  default     = null
}

variable "default_cache_policy_id" {
  description = "Cache policy ID for default behavior"
  type        = string
  default     = "658327ea-f89d-4fab-a63d-7e88639e58f6" # CachingOptimized
}

variable "default_origin_request_policy_id" {
  description = "Origin request policy ID for default behavior"
  type        = string
  default     = null
}

variable "response_headers_policy_id" {
  description = "Response headers policy ID"
  type        = string
  default     = null
}

variable "api_cache_policy_id" {
  description = "Cache policy ID for API behavior"
  type        = string
  default     = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad" # CachingDisabled
}

variable "api_origin_request_policy_id" {
  description = "Origin request policy ID for API behavior"
  type        = string
  default     = "216adef6-5c7f-47e4-b989-5492eafa07d3" # AllViewer
}

variable "create_api_cache_policy" {
  description = "Create custom API cache policy"
  type        = bool
  default     = false
}

variable "create_api_origin_request_policy" {
  description = "Create custom API origin request policy"
  type        = bool
  default     = false
}

variable "viewer_request_function_arn" {
  description = "CloudFront function ARN for viewer request"
  type        = string
  default     = null
}

variable "custom_error_responses" {
  description = "Custom error responses"
  type = list(object({
    error_code            = number
    response_code         = number
    response_page_path    = string
    error_caching_min_ttl = optional(number)
  }))
  default = []
}

variable "geo_restriction_type" {
  description = "Geo restriction type (none, whitelist, blacklist)"
  type        = string
  default     = "none"
}

variable "geo_restriction_locations" {
  description = "Geo restriction locations"
  type        = list(string)
  default     = []
}

variable "logging_bucket" {
  description = "S3 bucket for access logs"
  type        = string
  default     = null
}

variable "logging_prefix" {
  description = "Prefix for access logs"
  type        = string
  default     = "cloudfront/"
}

variable "logging_include_cookies" {
  description = "Include cookies in access logs"
  type        = bool
  default     = false
}

variable "tags" {
  description = "Additional tags for resources"
  type        = map(string)
  default     = {}
}
