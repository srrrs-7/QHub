# ACM Module
# Creates SSL/TLS certificates with DNS validation

terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
  }
}

# -----------------------------------------------------------------------------
# ACM Certificate
# -----------------------------------------------------------------------------
resource "aws_acm_certificate" "main" {
  domain_name               = var.domain_name
  subject_alternative_names = var.subject_alternative_names
  validation_method         = "DNS"

  lifecycle {
    create_before_destroy = true
  }

  tags = merge(var.tags, {
    Name = var.domain_name
  })
}

# -----------------------------------------------------------------------------
# Certificate Validation
# -----------------------------------------------------------------------------
resource "aws_acm_certificate_validation" "main" {
  count = var.validate_certificate ? 1 : 0

  certificate_arn         = aws_acm_certificate.main.arn
  validation_record_fqdns = [for record in aws_acm_certificate.main.domain_validation_options : record.resource_record_name]

  depends_on = [var.validation_record_fqdns]
}

# -----------------------------------------------------------------------------
# Outputs for DNS validation records
# (These need to be created in Route53 or other DNS provider)
# -----------------------------------------------------------------------------
locals {
  domain_validation_options = {
    for dvo in aws_acm_certificate.main.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      type   = dvo.resource_record_type
      record = dvo.resource_record_value
    }
  }
}
