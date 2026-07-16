terraform {
  required_version = ">= 1.7, < 2.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.70"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }

  # Backend S3 remoto con lock DynamoDB.
  # Le chiavi della state (`key`) sono per-environment: `prod/terraform.tfstate`,
  # `staging/terraform.tfstate`. Vengono impostate al `terraform init` con
  # `-backend-config=environments/<env>.backend.hcl`.
  #
  # BOOTSTRAP (una tantum, manuale, prima del primo `terraform init`):
  #   aws s3api create-bucket \
  #     --bucket tms-tfstate-<account_id> \
  #     --region eu-west-1 \
  #     --create-bucket-configuration LocationConstraint=eu-west-1
  #   aws s3api put-bucket-versioning \
  #     --bucket tms-tfstate-<account_id> \
  #     --versioning-configuration Status=Enabled
  #   aws s3api put-bucket-encryption \
  #     --bucket tms-tfstate-<account_id> \
  #     --server-side-encryption-configuration '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"}}]}'
  #   aws s3api put-public-access-block \
  #     --bucket tms-tfstate-<account_id> \
  #     --public-access-block-configuration BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true
  #   aws dynamodb create-table \
  #     --table-name tms-tfstate-lock \
  #     --attribute-definitions AttributeName=LockID,AttributeType=S \
  #     --key-schema AttributeName=LockID,KeyType=HASH \
  #     --billing-mode PAY_PER_REQUEST \
  #     --region eu-west-1
  backend "s3" {
    region         = "eu-west-1"
    encrypt        = true
    dynamodb_table = "tms-tfstate-lock"
    # `bucket` e `key` arrivano dal file -backend-config all'init.
  }
}

provider "aws" {
  region = var.region

  default_tags {
    tags = {
      Project     = "LoginBusiness"
      Environment = var.environment
      ManagedBy   = "terraform"
      Repository  = "gcaporossi-wheretech/LoginBusiness"
    }
  }
}
