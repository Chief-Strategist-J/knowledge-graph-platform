variable "project_name" { type = string }
variable "environment" { type = string }
variable "private_subnet_ids" { type = list(string) }
variable "kms_key_arn" { type = string }
variable "data_stores_sg_id" { type = string }
variable "db_instance_class" { type = string }
variable "redis_node_type" { type = string }
