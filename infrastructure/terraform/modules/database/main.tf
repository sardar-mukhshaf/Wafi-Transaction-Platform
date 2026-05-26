# ------------------------------------------------------------------------------
# RDS Aurora PostgreSQL Module - Per-service, encrypted, multi-AZ
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
  cluster_name = "${var.environment}-fabric-${var.service_name}"
  common_tags = merge(var.tags, {
    ManagedBy   = "terraform"
    Module      = "database"
    ServiceName = var.service_name
  })
}

# KMS Key for RDS encryption
resource "aws_kms_key" "rds" {
  description             = "RDS encryption key for ${var.service_name}"
  deletion_window_in_days = 30
  enable_key_rotation     = true

  tags = merge(local.common_tags, {
    Name = "${local.cluster_name}-encryption"
  })
}

resource "aws_kms_alias" "rds" {
  name          = "alias/${local.cluster_name}-encryption"
  target_key_id = aws_kms_key.rds.key_id
}

# DB Subnet Group
resource "aws_db_subnet_group" "main" {
  name       = "${local.cluster_name}-subnet-group"
  subnet_ids = var.subnet_ids

  tags = merge(local.common_tags, {
    Name = "${local.cluster_name}-subnet-group"
  })
}

# DB Parameter Group
resource "aws_rds_cluster_parameter_group" "main" {
  name        = "${local.cluster_name}-params"
  family      = "aurora-postgresql15"
  description = "Custom parameters for ${var.service_name}"

  parameter {
    name  = "log_connections"
    value = "1"
  }

  parameter {
    name  = "log_disconnections"
    value = "1"
  }

  parameter {
    name  = "log_min_duration_statement"
    value = "1000"
  }

  parameter {
    name  = "ssl"
    value = "1"
  }

  tags = local.common_tags
}

# Security Group
resource "aws_security_group" "rds" {
  name_prefix = "${local.cluster_name}-"
  description = "Security group for ${var.service_name} Aurora cluster"
  vpc_id      = var.vpc_id

  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = var.allowed_security_group_ids
    description     = "PostgreSQL access from EKS nodes"
  }

  tags = merge(local.common_tags, {
    Name = "${local.cluster_name}-sg"
  })

  lifecycle {
    create_before_destroy = true
  }
}

# Random password generation
resource "random_password" "master" {
  length           = 32
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

# Master secret in Secrets Manager
resource "aws_secretsmanager_secret" "master" {
  name                    = "${local.cluster_name}-master-credentials"
  description             = "Master credentials for ${var.service_name} Aurora cluster"
  kms_key_id              = aws_kms_key.rds.arn
  recovery_window_in_days = 7

  tags = local.common_tags
}

resource "aws_secretsmanager_secret_version" "master" {
  secret_id = aws_secretsmanager_secret.master.id
  secret_string = jsonencode({
    username = var.master_username
    password = random_password.master.result
    engine   = "postgres"
    host     = aws_rds_cluster.main.endpoint
    port     = 5432
    dbname   = var.database_name
  })
}

# Aurora Cluster
resource "aws_rds_cluster" "main" {
  cluster_identifier              = local.cluster_name
  engine                          = "aurora-postgresql"
  engine_version                  = var.engine_version
  engine_mode                     = "provisioned"
  database_name                   = var.database_name
  master_username                 = var.master_username
  master_password                 = random_password.master.result
  db_subnet_group_name            = aws_db_subnet_group.main.name
  vpc_security_group_ids          = [aws_security_group.rds.id]
  db_cluster_parameter_group_name = aws_rds_cluster_parameter_group.main.name

  backup_retention_period = var.backup_retention_period
  preferred_backup_window = "03:00-04:00"

  enabled_cloudwatch_logs_exports = ["postgresql"]

  storage_encrypted = true
  kms_key_id        = aws_kms_key.rds.arn

  deletion_protection = var.deletion_protection
  skip_final_snapshot = var.skip_final_snapshot

  apply_immediately = var.apply_immediately

  tags = merge(local.common_tags, {
    Name = local.cluster_name
  })
}

# Aurora Writer Instance
resource "aws_rds_cluster_instance" "writer" {
  identifier           = "${local.cluster_name}-writer"
  cluster_identifier   = aws_rds_cluster.main.id
  instance_class       = var.instance_class
  engine               = aws_rds_cluster.main.engine
  db_subnet_group_name = aws_db_subnet_group.main.name

  monitoring_interval = 60
  monitoring_role_arn = aws_iam_role.rds_monitoring.arn

  performance_insights_enabled    = true
  performance_insights_kms_key_id = aws_kms_key.rds.arn

  auto_minor_version_upgrade = true

  tags = merge(local.common_tags, {
    Name = "${local.cluster_name}-writer"
  })
}

# Aurora Reader Instance(s)
resource "aws_rds_cluster_instance" "reader" {
  count = var.reader_count

  identifier           = "${local.cluster_name}-reader-${count.index + 1}"
  cluster_identifier   = aws_rds_cluster.main.id
  instance_class       = var.instance_class
  engine               = aws_rds_cluster.main.engine
  db_subnet_group_name = aws_db_subnet_group.main.name

  monitoring_interval = 60
  monitoring_role_arn = aws_iam_role.rds_monitoring.arn

  performance_insights_enabled    = true
  performance_insights_kms_key_id = aws_kms_key.rds.arn

  auto_minor_version_upgrade = true

  tags = merge(local.common_tags, {
    Name = "${local.cluster_name}-reader-${count.index + 1}"
  })
}

# IAM Role for RDS Enhanced Monitoring
resource "aws_iam_role" "rds_monitoring" {
  name = "${local.cluster_name}-monitoring-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "monitoring.rds.amazonaws.com"
      }
    }]
  })

  tags = local.common_tags
}

resource "aws_iam_role_policy_attachment" "rds_monitoring" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
  role       = aws_iam_role.rds_monitoring.name
}
