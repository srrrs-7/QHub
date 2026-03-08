# CloudFront Module
# Creates CloudFront distribution for static assets and API

# -----------------------------------------------------------------------------
# Origin Access Control (for S3)
# -----------------------------------------------------------------------------
resource "aws_cloudfront_origin_access_control" "main" {
  count = var.enable_s3_origin ? 1 : 0

  name                              = "${var.project}-${var.environment}-oac"
  description                       = "OAC for ${var.project}-${var.environment}"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

# -----------------------------------------------------------------------------
# CloudFront Distribution
# -----------------------------------------------------------------------------
resource "aws_cloudfront_distribution" "main" {
  enabled             = true
  is_ipv6_enabled     = true
  comment             = "${var.project}-${var.environment} distribution"
  default_root_object = var.default_root_object
  aliases             = var.aliases
  price_class         = var.price_class
  web_acl_id          = var.web_acl_id

  # S3 Origin (for static assets)
  dynamic "origin" {
    for_each = var.s3_origin_domain_name != null ? [1] : []
    content {
      domain_name              = var.s3_origin_domain_name
      origin_id                = "S3Origin"
      origin_access_control_id = aws_cloudfront_origin_access_control.main[0].id
    }
  }

  # ALB Origin (for API)
  dynamic "origin" {
    for_each = var.alb_origin_domain_name != null ? [1] : []
    content {
      domain_name = var.alb_origin_domain_name
      origin_id   = "ALBOrigin"

      custom_origin_config {
        http_port              = 80
        https_port             = 443
        origin_protocol_policy = "https-only"
        origin_ssl_protocols   = ["TLSv1.2"]
      }

      dynamic "custom_header" {
        for_each = var.alb_custom_headers
        content {
          name  = custom_header.value.name
          value = custom_header.value.value
        }
      }
    }
  }

  # Default cache behavior (S3)
  default_cache_behavior {
    allowed_methods  = ["GET", "HEAD", "OPTIONS"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = var.s3_origin_domain_name != null ? "S3Origin" : "ALBOrigin"

    cache_policy_id            = var.default_cache_policy_id
    origin_request_policy_id   = var.default_origin_request_policy_id
    response_headers_policy_id = var.response_headers_policy_id

    viewer_protocol_policy = "redirect-to-https"
    compress               = true

    dynamic "function_association" {
      for_each = var.viewer_request_function_arn != null ? [1] : []
      content {
        event_type   = "viewer-request"
        function_arn = var.viewer_request_function_arn
      }
    }
  }

  # API cache behavior
  dynamic "ordered_cache_behavior" {
    for_each = var.alb_origin_domain_name != null ? [1] : []
    content {
      path_pattern     = "/api/*"
      allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
      cached_methods   = ["GET", "HEAD"]
      target_origin_id = "ALBOrigin"

      cache_policy_id          = var.api_cache_policy_id
      origin_request_policy_id = var.api_origin_request_policy_id

      viewer_protocol_policy = "https-only"
      compress               = true
    }
  }

  # Custom error responses
  dynamic "custom_error_response" {
    for_each = var.custom_error_responses
    content {
      error_code            = custom_error_response.value.error_code
      response_code         = custom_error_response.value.response_code
      response_page_path    = custom_error_response.value.response_page_path
      error_caching_min_ttl = lookup(custom_error_response.value, "error_caching_min_ttl", 300)
    }
  }

  # SSL Certificate
  viewer_certificate {
    acm_certificate_arn            = var.certificate_arn
    ssl_support_method             = var.certificate_arn != null ? "sni-only" : null
    minimum_protocol_version       = var.certificate_arn != null ? "TLSv1.2_2021" : null
    cloudfront_default_certificate = var.certificate_arn == null
  }

  # Restrictions
  restrictions {
    geo_restriction {
      restriction_type = var.geo_restriction_type
      locations        = var.geo_restriction_locations
    }
  }

  # Logging
  dynamic "logging_config" {
    for_each = var.logging_bucket != null ? [1] : []
    content {
      bucket          = var.logging_bucket
      prefix          = var.logging_prefix
      include_cookies = var.logging_include_cookies
    }
  }

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-cloudfront"
  })
}

# -----------------------------------------------------------------------------
# Cache Policies (Custom)
# -----------------------------------------------------------------------------
resource "aws_cloudfront_cache_policy" "api" {
  count = var.create_api_cache_policy ? 1 : 0

  name        = "${var.project}-${var.environment}-api-cache-policy"
  comment     = "Cache policy for API requests"
  default_ttl = 0
  max_ttl     = 1
  min_ttl     = 0

  parameters_in_cache_key_and_forwarded_to_origin {
    cookies_config {
      cookie_behavior = "none"
    }
    headers_config {
      header_behavior = "whitelist"
      headers {
        items = ["Authorization", "Host"]
      }
    }
    query_strings_config {
      query_string_behavior = "all"
    }
    enable_accept_encoding_brotli = true
    enable_accept_encoding_gzip   = true
  }
}

# -----------------------------------------------------------------------------
# Origin Request Policies (Custom)
# -----------------------------------------------------------------------------
resource "aws_cloudfront_origin_request_policy" "api" {
  count = var.create_api_origin_request_policy ? 1 : 0

  name    = "${var.project}-${var.environment}-api-origin-request-policy"
  comment = "Origin request policy for API requests"

  cookies_config {
    cookie_behavior = "all"
  }
  headers_config {
    header_behavior = "whitelist"
    headers {
      items = ["Accept", "Accept-Language", "Authorization", "Content-Type", "Origin", "Referer"]
    }
  }
  query_strings_config {
    query_string_behavior = "all"
  }
}
