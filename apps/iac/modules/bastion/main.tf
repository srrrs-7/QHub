# Bastion Host Module
# Creates a minimal EC2 instance for Session Manager access to private resources
# No SSH key required - access via AWS Systems Manager Session Manager only

data "aws_ami" "amazon_linux_2023" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-*-x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }
}

# -----------------------------------------------------------------------------
# IAM Role for Session Manager
# -----------------------------------------------------------------------------
resource "aws_iam_role" "bastion" {
  name = "${var.project}-${var.environment}-bastion-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "bastion_ssm" {
  role       = aws_iam_role.bastion.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

# Allow bastion to access RDS via IAM auth if enabled
resource "aws_iam_role_policy" "bastion_rds" {
  count = var.enable_rds_iam_auth ? 1 : 0

  name = "${var.project}-${var.environment}-bastion-rds-policy"
  role = aws_iam_role.bastion.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "rds-db:connect"
        ]
        Resource = "arn:aws:rds-db:${var.aws_region}:${data.aws_caller_identity.current.account_id}:dbuser:${var.rds_resource_id}/*"
      }
    ]
  })
}

# Allow bastion to read secrets from Secrets Manager
resource "aws_iam_role_policy" "bastion_secrets" {
  count = length(var.secrets_arns) > 0 ? 1 : 0

  name = "${var.project}-${var.environment}-bastion-secrets-policy"
  role = aws_iam_role.bastion.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource = var.secrets_arns
      }
    ]
  })
}

resource "aws_iam_instance_profile" "bastion" {
  name = "${var.project}-${var.environment}-bastion-profile"
  role = aws_iam_role.bastion.name

  tags = var.tags
}

data "aws_caller_identity" "current" {}

# -----------------------------------------------------------------------------
# Security Group
# -----------------------------------------------------------------------------
resource "aws_security_group" "bastion" {
  name        = "${var.project}-${var.environment}-bastion-sg"
  description = "Security group for bastion host (Session Manager only)"
  vpc_id      = var.vpc_id

  # No inbound rules needed - Session Manager uses outbound HTTPS

  egress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTPS for SSM"
  }

  egress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = var.db_security_group_ids
    description     = "PostgreSQL to database"
  }

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-bastion-sg"
  })
}

# Allow bastion to access database (ingress rule on DB security group)
resource "aws_vpc_security_group_ingress_rule" "db_from_bastion" {
  count = length(var.db_security_group_ids)

  security_group_id            = var.db_security_group_ids[count.index]
  description                  = "PostgreSQL from bastion"
  from_port                    = 5432
  to_port                      = 5432
  ip_protocol                  = "tcp"
  referenced_security_group_id = aws_security_group.bastion.id
}

# -----------------------------------------------------------------------------
# EC2 Instance
# -----------------------------------------------------------------------------
resource "aws_instance" "bastion" {
  ami                    = data.aws_ami.amazon_linux_2023.id
  instance_type          = var.instance_type
  subnet_id              = var.subnet_id
  vpc_security_group_ids = [aws_security_group.bastion.id]
  iam_instance_profile   = aws_iam_instance_profile.bastion.name

  # No SSH key - access via Session Manager only
  associate_public_ip_address = false

  # Minimal root volume
  root_block_device {
    volume_size           = 8
    volume_type           = "gp3"
    encrypted             = true
    delete_on_termination = true
  }

  # Install PostgreSQL client for DB access
  user_data = base64encode(<<-EOF
    #!/bin/bash
    dnf install -y postgresql15
    EOF
  )

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required" # IMDSv2 required
    http_put_response_hop_limit = 1
  }

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-bastion"
  })

  lifecycle {
    ignore_changes = [ami]
  }
}
