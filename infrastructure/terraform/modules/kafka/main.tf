# ------------------------------------------------------------------------------
# MSK Module - Managed Kafka with TLS, SCRAM auth, and multi-AZ
# ------------------------------------------------------------------------------

terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

locals {
  cluster_name = "${var.environment}-fabric-msk"
  common_tags = merge(var.tags, {
    ManagedBy = "terraform"
    Module    = "kafka"
  })
}

# KMS Key for MSK encryption
resource "aws_kms_key" "msk" {
  description             = "MSK encryption key"
  deletion_window_in_days = 30
  enable_key_rotation     = true

  tags = merge(local.common_tags, {
    Name = "${local.cluster_name}-encryption"
  })
}

resource "aws_kms_alias" "msk" {
  name          = "alias/${local.cluster_name}-encryption"
  target_key_id = aws_kms_key.msk.key_id
}

# MSK Cluster
resource "aws_msk_cluster" "main" {
  cluster_name           = local.cluster_name
  kafka_version          = var.kafka_version
  number_of_broker_nodes = var.number_of_broker_nodes

  broker_node_group_info {
    instance_type   = var.instance_type
    client_subnets  = var.subnet_ids
    security_groups = [aws_security_group.msk.id]

    storage_info {
      ebs_storage_info {
        volume_size = var.volume_size
      }
    }
  }

  encryption_info {
    encryption_at_rest_kms_key_arn = aws_kms_key.msk.arn
    encryption_in_transit {
      client_broker = "TLS"
      in_cluster    = true
    }
  }

  client_authentication {
    sasl {
      scram = true
    }
  }

  configuration_info {
    arn      = aws_msk_configuration.main.arn
    revision = aws_msk_configuration.main.latest_revision
  }

  logging_info {
    broker_logs {
      cloudwatch_logs {
        enabled   = true
        log_group = aws_cloudwatch_log_group.msk.name
      }
      s3 {
        enabled = true
        bucket  = aws_s3_bucket.msk_logs.id
        prefix  = "logs/msk-"
      }
    }
  }

  tags = merge(local.common_tags, {
    Name = local.cluster_name
  })
}

# MSK Configuration
resource "aws_msk_configuration" "main" {
  kafka_versions    = [var.kafka_version]
  name              = "${local.cluster_name}-config"
  server_properties = <<PROPERTIES
auto.create.topics.enable=false
default.replication.factor=3
min.insync.replicas=2
num.io.threads=8
num.network.threads=5
num.partitions=6
num.replica.fetchers=2
socket.receive.buffer.bytes=102400
socket.request.max.bytes=104857600
socket.send.buffer.bytes=102400
unclean.leader.election.enable=false
PROPERTIES
}

# MSK Security Group
resource "aws_security_group" "msk" {
  name_prefix = "${local.cluster_name}-"
  description = "Security group for MSK cluster"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 9094
    to_port     = 9094
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
    description = "TLS Kafka traffic"
  }

  ingress {
    from_port   = 9096
    to_port     = 9096
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
    description = "SASL/SCRAM Kafka traffic"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(local.common_tags, {
    Name = "${local.cluster_name}-sg"
  })

  lifecycle {
    create_before_destroy = true
  }
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "msk" {
  name              = "/aws/msk/${local.cluster_name}"
  retention_in_days = 30

  tags = local.common_tags
}

# S3 Bucket for MSK logs
resource "aws_s3_bucket" "msk_logs" {
  bucket = "${var.environment}-fabric-msk-logs-${data.aws_caller_identity.current.account_id}"

  tags = merge(local.common_tags, {
    Name = "${var.environment}-fabric-msk-logs"
  })
}

resource "aws_s3_bucket_versioning" "msk_logs" {
  bucket = aws_s3_bucket.msk_logs.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "msk_logs" {
  bucket = aws_s3_bucket.msk_logs.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.msk.arn
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "msk_logs" {
  bucket = aws_s3_bucket.msk_logs.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

data "aws_caller_identity" "current" {}

# Secret Manager for MSK SASL/SCRAM credentials
resource "aws_secretsmanager_secret" "msk" {
  name                    = "${local.cluster_name}-scram-credentials"
  description             = "SASL/SCRAM credentials for MSK cluster"
  kms_key_id              = aws_kms_key.msk.arn
  recovery_window_in_days = 7

  tags = local.common_tags
}

resource "aws_secretsmanager_secret_version" "msk" {
  secret_id = aws_secretsmanager_secret.msk.id
  secret_string = jsonencode({
    username = var.scram_username
    password = var.scram_password
  })
}

resource "aws_msk_scram_secret_association" "main" {
  cluster_arn     = aws_msk_cluster.main.arn
  secret_arn_list = [aws_secretsmanager_secret.msk.arn]
}
