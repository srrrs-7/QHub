output "zone_id" {
  description = "Route53 hosted zone ID"
  value       = local.zone_id
}

output "zone_name" {
  description = "Route53 hosted zone name"
  value       = var.domain_name
}

output "name_servers" {
  description = "Name servers for the hosted zone"
  value       = var.create_hosted_zone ? aws_route53_zone.main[0].name_servers : null
}

output "cloudfront_record_fqdn" {
  description = "FQDN of the CloudFront record"
  value       = var.cloudfront_domain_name != null ? aws_route53_record.cloudfront[0].fqdn : null
}

output "alb_record_fqdn" {
  description = "FQDN of the ALB record"
  value       = var.alb_dns_name != null && var.cloudfront_domain_name == null ? aws_route53_record.alb[0].fqdn : null
}

output "api_record_fqdn" {
  description = "FQDN of the API record"
  value       = var.api_subdomain != null ? aws_route53_record.api[0].fqdn : null
}
