provider "aws" {
  region = "eu-west-1"
}

resource "aws_s3_bucket" "contract_attachments" {
  bucket = "contract-attachments-demo"
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
    from_port   = 0
    to_port     = 65535
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
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
  storage_encrypted   = true
  publicly_accessible = false
  skip_final_snapshot = true
}
