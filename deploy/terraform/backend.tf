terraform {
  backend "s3" {
    bucket         = "kg-platform-tfstate"
    key            = "knowledge-graph-platform/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "kg-platform-tflocks"
  }
}
