variable "environment" {
  description = "Environment name"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
}

variable "subnet_ids" {
  description = "Subnet IDs for broker nodes"
  type        = list(string)
}

variable "kafka_version" {
  description = "Kafka version"
  type        = string
  default     = "3.5.1"
}

variable "instance_type" {
  description = "Broker instance type"
  type        = string
  default     = "kafka.m5.large"
}

variable "number_of_broker_nodes" {
  description = "Number of broker nodes (must be multiple of AZ count)"
  type        = number
  default     = 3
}

variable "volume_size" {
  description = "EBS volume size per broker (GB)"
  type        = number
  default     = 1000
}

variable "scram_username" {
  description = "SASL/SCRAM username"
  type        = string
  default     = "fabric-admin"
}

variable "scram_password" {
  description = "SASL/SCRAM password"
  type        = string
  sensitive   = true
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}
