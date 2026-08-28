provider "aws" {
  region = "eu-west-1"
}

variable "admin_cidr_blocks" {
  description = "CIDR ranges allowed to reach administration services (SSH)"
  type        = list(string)
  default     = ["10.0.0.0/8"]
}

variable "access_log_bucket" {
  description = "Existing S3 bucket that receives server access logs"
  type        = string
  default     = "contract-attachments-demo-access-logs"
}

resource "aws_s3_bucket" "contract_attachments" {
  bucket = "contract-attachments-demo"
}

resource "aws_s3_bucket_policy" "contract_attachments_https_only" {
  bucket = aws_s3_bucket.contract_attachments.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "DenyNonHttpsAccess"
        Effect    = "Deny"
        Principal = "*"
        Action    = "s3:*"
        Resource = [
          aws_s3_bucket.contract_attachments.arn,
          "${aws_s3_bucket.contract_attachments.arn}/*",
        ]
        Condition = {
          Bool = {
            "aws:SecureTransport" = "false"
          }
        }
      },
    ]
  })
}

resource "aws_s3_bucket_logging" "contract_attachments" {
  bucket        = aws_s3_bucket.contract_attachments.id
  target_bucket = var.access_log_bucket
  target_prefix = "contract-attachments/"
}

resource "aws_s3_bucket_acl" "contract_attachments" {
  bucket = aws_s3_bucket.contract_attachments.id
  acl    = "public-read"
}

resource "aws_security_group" "contract_api" {
  name        = "contract-api"
  description = "Contract API ingress"

  ingress {
    description = "API"
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "SSH administration"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = var.admin_cidr_blocks
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_db_instance" "contracts" {
  identifier          = "contracts"
  engine              = "postgres"
  instance_class      = "db.t3.micro"
  allocated_storage   = 20
  username            = "contracts_admin"
  password            = "ChangeMe123!"
  storage_encrypted   = false
  publicly_accessible = true
  skip_final_snapshot = true

  backup_retention_period = 30
}
