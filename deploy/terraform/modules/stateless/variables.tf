variable "project_name" { type = string }
variable "environment" { type = string }
variable "vpc_id" { type = string }
variable "public_subnet_ids" { type = list(string) }
variable "private_subnet_ids" { type = list(string) }
variable "kms_key_arn" { type = string }
variable "alb_sg_id" { type = string }
variable "eks_nodes_sg_id" { type = string }
variable "eks_cluster_role_arn" { type = string }
variable "eks_nodes_role_arn" { type = string }
variable "eks_node_instance_types" { type = list(string) }
variable "eks_min_size" { type = number }
variable "eks_max_size" { type = number }
variable "eks_desired_size" { type = number }
variable "cluster_name" {
  type    = string
  default = ""
}
