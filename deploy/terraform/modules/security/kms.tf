resource "aws_kms_key" "primary" {
  description             = "Customer Managed Key for ${var.project_name}-${var.environment}"
  deletion_window_in_days = 30
  enable_key_rotation     = true

  tags = {
    Name        = "${var.project_name}-${var.environment}-kms"
    Environment = var.environment
    Project     = var.project_name
  }
}

resource "aws_kms_alias" "primary" {
  name          = "alias/${var.project_name}-${var.environment}"
  target_key_id = aws_kms_key.primary.key_id
}

output "kms_key_arn" {
  value = aws_kms_key.primary.arn
}

output "kms_key_id" {
  value = aws_kms_key.primary.key_id
}
