# ============================================
# Terraform Configuration for Trading Engine
# ============================================

terraform {
  required_version = ">= 1.6"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.24"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.12"
    }
  }

  backend "s3" {
    bucket         = "trading-engine-terraform-state"
    key            = "production/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-state-lock"
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Environment = var.environment
      Project     = "trading-engine"
      ManagedBy   = "terraform"
    }
  }
}

# ============================================
# VPC and Networking
# ============================================

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "${var.environment}-trading-engine-vpc"
  cidr = var.vpc_cidr

  azs             = data.aws_availability_zones.available.names
  private_subnets = var.private_subnet_cidrs
  public_subnets  = var.public_subnet_cidrs

  enable_nat_gateway   = true
  enable_vpn_gateway   = false
  enable_dns_hostnames = true
  enable_dns_support   = true

  # Single NAT Gateway for cost optimization (use one_nat_gateway_per_az for production)
  single_nat_gateway = var.environment != "production"

  tags = {
    "kubernetes.io/cluster/${var.cluster_name}" = "shared"
  }

  public_subnet_tags = {
    "kubernetes.io/role/elb" = "1"
  }

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = "1"
  }
}

# ============================================
# EKS Cluster
# ============================================

module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 19.0"

  cluster_name    = var.cluster_name
  cluster_version = "1.28"

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  cluster_endpoint_public_access  = true
  cluster_endpoint_private_access = true

  # Cluster encryption
  cluster_encryption_config = {
    resources        = ["secrets"]
    provider_key_arn = aws_kms_key.eks.arn
  }

  # Node groups
  eks_managed_node_groups = {
    general = {
      min_size     = 3
      max_size     = 10
      desired_size = 3

      instance_types = ["t3.xlarge"]
      capacity_type  = "ON_DEMAND"

      labels = {
        role = "general"
      }

      taints = []
    }

    compute_optimized = {
      min_size     = 2
      max_size     = 20
      desired_size = 2

      instance_types = ["c6i.2xlarge"]
      capacity_type  = "SPOT"

      labels = {
        role = "compute"
      }

      taints = [{
        key    = "workload"
        value  = "compute"
        effect = "NoSchedule"
      }]
    }

    memory_optimized = {
      min_size     = 1
      max_size     = 5
      desired_size = 1

      instance_types = ["r6i.xlarge"]
      capacity_type  = "ON_DEMAND"

      labels = {
        role = "memory"
      }

      taints = [{
        key    = "workload"
        value  = "memory"
        effect = "NoSchedule"
      }]
    }
  }

  # Cluster add-ons
  cluster_addons = {
    coredns = {
      most_recent = true
    }
    kube-proxy = {
      most_recent = true
    }
    vpc-cni = {
      most_recent = true
    }
    aws-ebs-csi-driver = {
      most_recent = true
    }
  }
}

# ============================================
# RDS PostgreSQL
# ============================================

module "rds" {
  source  = "terraform-aws-modules/rds/aws"
  version = "~> 6.0"

  identifier = "${var.environment}-trading-engine-db"

  engine               = "postgres"
  engine_version       = "16.1"
  family               = "postgres16"
  major_engine_version = "16"
  instance_class       = var.db_instance_class

  allocated_storage     = 100
  max_allocated_storage = 1000
  storage_encrypted     = true
  kms_key_id            = aws_kms_key.rds.arn

  db_name  = "trading_engine"
  username = "trading_admin"
  port     = 5432

  multi_az               = var.environment == "production"
  db_subnet_group_name   = module.vpc.database_subnet_group
  vpc_security_group_ids = [aws_security_group.rds.id]

  # Backup configuration
  backup_retention_period = var.environment == "production" ? 30 : 7
  backup_window           = "03:00-04:00"
  maintenance_window      = "Mon:04:00-Mon:05:00"

  # Enhanced monitoring
  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]
  monitoring_interval             = 60
  monitoring_role_arn             = aws_iam_role.rds_monitoring.arn

  # Performance Insights
  performance_insights_enabled    = true
  performance_insights_kms_key_id = aws_kms_key.rds.arn

  # Read replicas
  create_db_instance = true

  # Parameters
  parameters = [
    {
      name  = "max_connections"
      value = "1000"
    },
    {
      name  = "shared_buffers"
      value = "{DBInstanceClassMemory/32768}"
    },
    {
      name  = "effective_cache_size"
      value = "{DBInstanceClassMemory/16384}"
    },
    {
      name  = "maintenance_work_mem"
      value = "2097152"
    },
    {
      name  = "checkpoint_completion_target"
      value = "0.9"
    },
    {
      name  = "wal_buffers"
      value = "16384"
    },
    {
      name  = "default_statistics_target"
      value = "100"
    },
    {
      name  = "random_page_cost"
      value = "1.1"
    },
    {
      name  = "effective_io_concurrency"
      value = "200"
    },
    {
      name  = "work_mem"
      value = "10485"
    },
    {
      name  = "min_wal_size"
      value = "2048"
    },
    {
      name  = "max_wal_size"
      value = "8192"
    }
  ]

  tags = {
    Name = "${var.environment}-trading-engine-db"
  }
}

# Read replica for reporting/analytics
resource "aws_db_instance" "read_replica" {
  count = var.environment == "production" ? 2 : 0

  identifier          = "${var.environment}-trading-engine-db-replica-${count.index + 1}"
  replicate_source_db = module.rds.db_instance_id

  instance_class = var.db_replica_instance_class

  # Performance Insights
  performance_insights_enabled = true

  tags = {
    Name = "${var.environment}-trading-engine-db-replica-${count.index + 1}"
  }
}

# ============================================
# ElastiCache Redis
# ============================================

module "redis" {
  source  = "terraform-aws-modules/elasticache/aws"
  version = "~> 1.0"

  cluster_id           = "${var.environment}-trading-engine-redis"
  engine               = "redis"
  engine_version       = "7.1"
  node_type            = var.redis_node_type
  num_cache_nodes      = 1
  parameter_group_name = aws_elasticache_parameter_group.redis.name

  subnet_group_name  = module.vpc.elasticache_subnet_group_name
  security_group_ids = [aws_security_group.redis.id]

  # Redis configuration
  port = 6379

  # Automatic failover
  automatic_failover_enabled = var.environment == "production"
  multi_az_enabled           = var.environment == "production"

  # Maintenance
  maintenance_window = "sun:05:00-sun:06:00"
  snapshot_window    = "03:00-04:00"

  # Backup
  snapshot_retention_limit = var.environment == "production" ? 7 : 1

  # Encryption
  at_rest_encryption_enabled = true
  transit_encryption_enabled = true
  auth_token_enabled         = true

  tags = {
    Name = "${var.environment}-trading-engine-redis"
  }
}

resource "aws_elasticache_parameter_group" "redis" {
  name   = "${var.environment}-trading-engine-redis-params"
  family = "redis7"

  parameter {
    name  = "maxmemory-policy"
    value = "allkeys-lru"
  }

  parameter {
    name  = "timeout"
    value = "300"
  }
}

# ============================================
# S3 Buckets
# ============================================

# Backups bucket
module "s3_backups" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "~> 3.0"

  bucket = "${var.environment}-trading-engine-backups"

  versioning = {
    enabled = true
  }

  lifecycle_rule = [
    {
      id      = "archive"
      enabled = true

      transition = [
        {
          days          = 30
          storage_class = "STANDARD_IA"
        },
        {
          days          = 90
          storage_class = "GLACIER"
        }
      ]

      expiration = {
        days = 365
      }
    }
  ]

  server_side_encryption_configuration = {
    rule = {
      apply_server_side_encryption_by_default = {
        sse_algorithm     = "aws:kms"
        kms_master_key_id = aws_kms_key.s3.arn
      }
    }
  }
}

# Logs bucket
module "s3_logs" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "~> 3.0"

  bucket = "${var.environment}-trading-engine-logs"

  lifecycle_rule = [
    {
      id      = "delete-old-logs"
      enabled = true

      expiration = {
        days = 90
      }
    }
  ]
}

# ============================================
# KMS Keys
# ============================================

resource "aws_kms_key" "eks" {
  description             = "EKS cluster encryption key"
  deletion_window_in_days = 10
  enable_key_rotation     = true
}

resource "aws_kms_alias" "eks" {
  name          = "alias/${var.environment}-trading-engine-eks"
  target_key_id = aws_kms_key.eks.key_id
}

resource "aws_kms_key" "rds" {
  description             = "RDS encryption key"
  deletion_window_in_days = 10
  enable_key_rotation     = true
}

resource "aws_kms_alias" "rds" {
  name          = "alias/${var.environment}-trading-engine-rds"
  target_key_id = aws_kms_key.rds.key_id
}

resource "aws_kms_key" "s3" {
  description             = "S3 encryption key"
  deletion_window_in_days = 10
  enable_key_rotation     = true
}

resource "aws_kms_alias" "s3" {
  name          = "alias/${var.environment}-trading-engine-s3"
  target_key_id = aws_kms_key.s3.key_id
}

# ============================================
# Security Groups
# ============================================

resource "aws_security_group" "rds" {
  name_prefix = "${var.environment}-trading-engine-rds-"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [module.eks.node_security_group_id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.environment}-trading-engine-rds-sg"
  }
}

resource "aws_security_group" "redis" {
  name_prefix = "${var.environment}-trading-engine-redis-"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port       = 6379
    to_port         = 6379
    protocol        = "tcp"
    security_groups = [module.eks.node_security_group_id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.environment}-trading-engine-redis-sg"
  }
}

# ============================================
# IAM Roles
# ============================================

resource "aws_iam_role" "rds_monitoring" {
  name = "${var.environment}-trading-engine-rds-monitoring"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "monitoring.rds.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "rds_monitoring" {
  role       = aws_iam_role.rds_monitoring.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
}

# ============================================
# Data Sources
# ============================================

data "aws_availability_zones" "available" {
  state = "available"
}
