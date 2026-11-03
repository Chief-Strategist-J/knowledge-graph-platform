provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = var.project_name
      Environment = var.environment
      ManagedBy   = "Terraform"
    }
  }
}

# Availability zones data source
data "aws_availability_zones" "available" {
  state = "available"
}

# Core VPC & Subnets
resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "${var.project_name}-${var.environment}-vpc"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "${var.project_name}-${var.environment}-igw"
  }
}

resource "aws_subnet" "public" {
  count                   = 2
  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(var.vpc_cidr, 4, count.index)
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name                                        = "${var.project_name}-${var.environment}-public-${count.index}"
    "kubernetes.io/role/elb"                    = "1"
    "kubernetes.io/cluster/${var.project_name}" = "shared"
  }
}

resource "aws_subnet" "private" {
  count             = 2
  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 4, count.index + 2)
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name                                        = "${var.project_name}-${var.environment}-private-${count.index}"
    "kubernetes.io/role/internal-elb"           = "1"
    "kubernetes.io/cluster/${var.project_name}" = "shared"
  }
}

resource "aws_eip" "nat" {
  domain = "vpc"
}

resource "aws_nat_gateway" "nat" {
  allocation_id = aws_eip.nat.id
  subnet_id     = aws_subnet.public[0].id

  depends_on = [aws_internet_gateway.gw]
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.gw.id
  }
}

resource "aws_route_table_association" "public" {
  count          = 2
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.nat.id
  }
}

resource "aws_route_table_association" "private" {
  count          = 2
  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private.id
}

# Module Layer 1: Security
module "security" {
  source = "./modules/security"

  project_name = var.project_name
  environment  = var.environment
  vpc_id       = aws_vpc.main.id
  vpc_cidr     = var.vpc_cidr
}

# Module Layer 2: Stateful Datastores
module "stateful" {
  source = "./modules/stateful"

  project_name       = var.project_name
  environment        = var.environment
  private_subnet_ids = aws_subnet.private[*].id
  kms_key_arn        = module.security.kms_key_arn
  data_stores_sg_id  = module.security.data_stores_sg_id
  db_instance_class  = var.db_instance_class
  redis_node_type    = var.redis_node_type
}

# Module Layer 3: Stateless Compute & Ingress
module "stateless" {
  source = "./modules/stateless"

  project_name            = var.project_name
  environment             = var.environment
  vpc_id                  = aws_vpc.main.id
  public_subnet_ids       = aws_subnet.public[*].id
  private_subnet_ids      = aws_subnet.private[*].id
  kms_key_arn             = module.security.kms_key_arn
  alb_sg_id               = module.security.alb_sg_id
  eks_nodes_sg_id         = module.security.eks_nodes_sg_id
  eks_cluster_role_arn    = module.security.eks_cluster_role_arn
  eks_nodes_role_arn      = module.security.eks_nodes_role_arn
  eks_node_instance_types = var.eks_node_instance_types
  eks_min_size            = var.eks_min_size
  eks_max_size            = var.eks_max_size
  eks_desired_size        = var.eks_desired_size
}
