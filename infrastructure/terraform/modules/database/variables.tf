variable "environment" {
  description = "Environment name"
  type        = string
}

variable "service_name" {
  description = "Name of the service owning this database"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "subnet_ids" {
  description = "Database subnet IDs"
  type        = list(string)
}

variable "allowed_security_group_ids" {
  description = "Security groups allowed to connect to RDS"
  type        = list(string)
  default     = []
}

variable "engine_version" {
  description = "Aurora PostgreSQL engine version"
  type        = string
  default     = "15.4"
}

variable "database_name" {
  description = "Default database name"
  type        = string
}

variable "master_username" {
  description = "Master database username"
  type        = string
  default     = "fabric_admin"
}

variable "instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.r6g.large"
}

variable "reader_count" {
  description = "Number of reader instances"
  type        = number
  default     = 1
}

variable "backup_retention_period" {
  description = "Backup retention in days"
  type        = number
  default     = 30
}

variable "deletion_protection" {
  description = "Enable deletion protection"
  type        = bool
  default     = true
}

variable "skip_final_snapshot" {
  description = "Skip final snapshot on destroy"
  type        = bool
  default     = false
}

variable "apply_immediately" {
  description = "Apply changes immediately"
  type        = bool
  default     = false
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}
