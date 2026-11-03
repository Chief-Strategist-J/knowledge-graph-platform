variable "environment" {
  description = "Target deployment environment (dev, staging, prod)"
  type        = string
}

variable "aws_region" {
  description = "AWS region for resources"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Project moniker"
  type        = string
  default     = "knowledge-graph-platform"
}

variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
  default     = "10.0.0.0/16"
}

variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.r6g.xlarge"
}

variable "redis_node_type" {
  description = "ElastiCache node type"
  type        = string
  default     = "cache.r6g.large"
}

variable "eks_node_instance_types" {
  description = "EKS worker node instance types"
  type        = list(string)
  default     = ["m6i.xlarge", "m6a.xlarge"]
}

variable "eks_min_size" {
  description = "Minimum worker nodes"
  type        = number
  default     = 2
}

variable "eks_max_size" {
  description = "Maximum worker nodes"
  type        = number
  default     = 10
}

variable "eks_desired_size" {
  description = "Desired worker nodes"
  type        = number
  default     = 3
}
