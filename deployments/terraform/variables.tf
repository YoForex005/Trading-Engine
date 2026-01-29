# ============================================
# Terraform Variables
# ============================================

variable "aws_region" {
  description = "AWS region for all resources"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name (dev, staging, production)"
  type        = string

  validation {
    condition     = contains(["dev", "staging", "production"], var.environment)
    error_message = "Environment must be dev, staging, or production."
  }
}

variable "cluster_name" {
  description = "EKS cluster name"
  type        = string
  default     = "trading-engine"
}

# ============================================
# Network Configuration
# ============================================

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for private subnets"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for public subnets"
  type        = list(string)
  default     = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]
}

# ============================================
# Database Configuration
# ============================================

variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.r6i.xlarge"
}

variable "db_replica_instance_class" {
  description = "RDS read replica instance class"
  type        = string
  default     = "db.r6i.large"
}

# ============================================
# Redis Configuration
# ============================================

variable "redis_node_type" {
  description = "ElastiCache node type"
  type        = string
  default     = "cache.r6g.large"
}

# ============================================
# Tags
# ============================================

variable "tags" {
  description = "Additional tags for all resources"
  type        = map(string)
  default     = {}
}
