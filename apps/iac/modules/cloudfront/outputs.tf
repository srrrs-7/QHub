output "distribution_id" {
  description = "CloudFront distribution ID"
  value       = aws_cloudfront_distribution.main.id
}

output "distribution_arn" {
  description = "CloudFront distribution ARN"
  value       = aws_cloudfront_distribution.main.arn
}

output "distribution_domain_name" {
  description = "CloudFront distribution domain name"
  value       = aws_cloudfront_distribution.main.domain_name
}

output "distribution_hosted_zone_id" {
  description = "CloudFront distribution hosted zone ID"
  value       = aws_cloudfront_distribution.main.hosted_zone_id
}

output "oac_id" {
  description = "Origin Access Control ID"
  value       = var.s3_origin_domain_name != null ? aws_cloudfront_origin_access_control.main[0].id : null
}

output "api_cache_policy_id" {
  description = "Custom API cache policy ID"
  value       = var.create_api_cache_policy ? aws_cloudfront_cache_policy.api[0].id : null
}

output "api_origin_request_policy_id" {
  description = "Custom API origin request policy ID"
  value       = var.create_api_origin_request_policy ? aws_cloudfront_origin_request_policy.api[0].id : null
}
