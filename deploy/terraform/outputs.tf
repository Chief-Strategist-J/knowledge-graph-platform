output "vpc_id" {
  description = "VPC identifier"
  value       = aws_vpc.main.id
}

output "eks_cluster_name" {
  description = "EKS cluster name"
  value       = module.stateless.cluster_name
}

output "eks_cluster_endpoint" {
  description = "EKS control plane endpoint"
  value       = module.stateless.cluster_endpoint
}

output "alb_dns_name" {
  description = "Public ALB DNS name"
  value       = module.stateless.alb_dns_name
}

output "postgres_endpoint" {
  description = "PostgreSQL endpoint"
  value       = module.stateful.postgres_endpoint
}

output "redis_endpoint" {
  description = "Redis endpoint"
  value       = module.stateful.redis_endpoint
}

output "kms_key_arn" {
  description = "Primary KMS key ARN"
  value       = module.security.kms_key_arn
}

output "artifacts_bucket" {
  description = "S3 bucket for document & artifact storage"
  value       = module.stateful.artifacts_bucket_name
}
