variable "domain_name" {
  description = "Primary domain name for the certificate"
  type        = string
}

variable "subject_alternative_names" {
  description = "Subject alternative names (e.g., *.example.com)"
  type        = list(string)
  default     = []
}

variable "validate_certificate" {
  description = "Wait for certificate validation"
  type        = bool
  default     = true
}

variable "validation_record_fqdns" {
  description = "FQDNs of validation records (for dependency)"
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Additional tags for resources"
  type        = map(string)
  default     = {}
}
