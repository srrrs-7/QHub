# Route53 Module
# Creates Route53 hosted zone and DNS records

# -----------------------------------------------------------------------------
# Hosted Zone (optional - can use existing)
# -----------------------------------------------------------------------------
resource "aws_route53_zone" "main" {
  count = var.create_hosted_zone ? 1 : 0

  name    = var.domain_name
  comment = "Managed by Terraform for ${var.project}"

  tags = merge(var.tags, {
    Name = var.domain_name
  })
}

locals {
  zone_id = var.create_hosted_zone ? aws_route53_zone.main[0].zone_id : var.hosted_zone_id
}

# -----------------------------------------------------------------------------
# A Record for CloudFront
# -----------------------------------------------------------------------------
resource "aws_route53_record" "cloudfront" {
  count = var.enable_cloudfront_record ? 1 : 0

  zone_id = local.zone_id
  name    = var.subdomain != null ? "${var.subdomain}.${var.domain_name}" : var.domain_name
  type    = "A"

  alias {
    name                   = var.cloudfront_domain_name
    zone_id                = var.cloudfront_hosted_zone_id
    evaluate_target_health = false
  }
}

resource "aws_route53_record" "cloudfront_aaaa" {
  count = var.enable_cloudfront_record ? 1 : 0

  zone_id = local.zone_id
  name    = var.subdomain != null ? "${var.subdomain}.${var.domain_name}" : var.domain_name
  type    = "AAAA"

  alias {
    name                   = var.cloudfront_domain_name
    zone_id                = var.cloudfront_hosted_zone_id
    evaluate_target_health = false
  }
}

# -----------------------------------------------------------------------------
# A Record for ALB (if not using CloudFront)
# -----------------------------------------------------------------------------
resource "aws_route53_record" "alb" {
  count = !var.enable_cloudfront_record ? 1 : 0

  zone_id = local.zone_id
  name    = var.subdomain != null ? "${var.subdomain}.${var.domain_name}" : var.domain_name
  type    = "A"

  alias {
    name                   = var.alb_dns_name
    zone_id                = var.alb_zone_id
    evaluate_target_health = true
  }
}

# -----------------------------------------------------------------------------
# API Subdomain Record
# -----------------------------------------------------------------------------
resource "aws_route53_record" "api" {
  count = var.enable_api_record ? 1 : 0

  zone_id = local.zone_id
  name    = "${var.api_subdomain}.${var.domain_name}"
  type    = "A"

  alias {
    name                   = var.alb_dns_name
    zone_id                = var.alb_zone_id
    evaluate_target_health = true
  }
}

# -----------------------------------------------------------------------------
# ACM Certificate Validation Records
# -----------------------------------------------------------------------------
resource "aws_route53_record" "acm_validation" {
  for_each = var.acm_certificate_validation_records

  zone_id = local.zone_id
  name    = each.value.name
  type    = each.value.type
  ttl     = 300
  records = [each.value.record]
}

# -----------------------------------------------------------------------------
# Additional Records
# -----------------------------------------------------------------------------
resource "aws_route53_record" "additional" {
  for_each = var.additional_records

  zone_id = local.zone_id
  name    = each.value.name
  type    = each.value.type
  ttl     = lookup(each.value, "ttl", null)
  records = lookup(each.value, "records", null)

  dynamic "alias" {
    for_each = lookup(each.value, "alias", null) != null ? [each.value.alias] : []
    content {
      name                   = alias.value.name
      zone_id                = alias.value.zone_id
      evaluate_target_health = lookup(alias.value, "evaluate_target_health", false)
    }
  }
}
