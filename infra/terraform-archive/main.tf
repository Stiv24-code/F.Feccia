# Infra minimal: 1 VPC + 1 subnet pubblica + 1 EC2 t3.small (Ubuntu 22.04)
# + EBS data + Elastic IP + Security Group + SSH key pair.
#
# DNS NON gestito qui: dopo `terraform apply` si legge l'output
# `public_ip` e si configurano manualmente su Google DNS:
#   A tms.wheretech.it       -> <public_ip>
#   A api.tms.wheretech.it   -> <public_ip>

# ── Discovery AMI ────────────────────────────────────────────────────────
# Ubuntu 22.04 LTS ufficiale Canonical (owner 099720109477), amd64 + EBS gp3.
data "aws_ami" "ubuntu_2204" {
  most_recent = true
  owners      = ["099720109477"]

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }
}

# ── Networking ───────────────────────────────────────────────────────────
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = "tms-${var.environment}"
  }
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "tms-${var.environment}-igw"
  }
}

resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = "${var.region}a"
  map_public_ip_on_launch = true

  tags = {
    Name = "tms-${var.environment}-public-a"
  }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name = "tms-${var.environment}-rt-public"
  }
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

# ── Security Group ───────────────────────────────────────────────────────
# SSH solo dal tuo admin_cidr. HTTP+HTTPS aperte al mondo (HTTPS serve per
# l'app, HTTP solo per il challenge HTTP-01 di Let's Encrypt).
resource "aws_security_group" "ec2" {
  name        = "tms-${var.environment}-ec2"
  description = "Firewall TMS: SSH admin + HTTP + HTTPS"
  vpc_id      = aws_vpc.main.id

  ingress {
    description = "SSH dall'IP amministrativo"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = [var.admin_cidr]
  }

  ingress {
    description = "HTTP (redirect a HTTPS + ACME http-01 challenge)"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTPS"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    description = "Traffico uscita illimitato (apt, docker pull, AWS, OCSP)"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "tms-${var.environment}-ec2"
  }
}

# ── SSH key pairs ────────────────────────────────────────────────────────
resource "aws_key_pair" "admin" {
  key_name   = "tms-${var.environment}-admin"
  public_key = var.ssh_public_key
}

# ── EC2 Instance ─────────────────────────────────────────────────────────
resource "aws_instance" "main" {
  ami                         = data.aws_ami.ubuntu_2204.id
  instance_type               = var.instance_type
  subnet_id                   = aws_subnet.public.id
  vpc_security_group_ids      = [aws_security_group.ec2.id]
  key_name                    = aws_key_pair.admin.key_name
  iam_instance_profile        = aws_iam_instance_profile.ec2.name
  associate_public_ip_address = true
  disable_api_termination     = var.environment == "prod"

  root_block_device {
    volume_size           = var.root_volume_gb
    volume_type           = "gp3"
    encrypted             = true
    delete_on_termination = true
  }

  # Proteggi da stop accidentale che libererebbe l'EIP.
  # (l'EIP rimane attaccato al volume se lo associamo sotto con aws_eip_association.)

  user_data = templatefile("${path.module}/user_data.sh.tftpl", {
    admin_email           = var.admin_email
    app_fqdn              = var.app_fqdn
    api_fqdn              = var.api_fqdn
    deploy_ssh_public_key = var.deploy_ssh_public_key
    environment           = var.environment
    region                = var.region
    backups_bucket        = aws_s3_bucket.backups.id
    invoices_bucket       = aws_s3_bucket.invoices.id
  })
  user_data_replace_on_change = false

  # Se il user_data cambia significativamente (es. nuova versione di Docker),
  # l'utente può forzare una nuova istanza con:
  #   terraform taint aws_instance.main && terraform apply
  # oppure impostare user_data_replace_on_change = true.

  tags = {
    Name = "tms-${var.environment}"
  }

  # Il volume dati è gestito come risorsa separata per poter snapshot+restore
  # senza toccare il root. Attachment sotto.
  lifecycle {
    ignore_changes = [ami] # così un nuovo AMI base non forza replace
  }
}

# ── Volume EBS dedicato ai dati (MongoDB, log persistenti, backup locali) ──
resource "aws_ebs_volume" "data" {
  availability_zone = aws_instance.main.availability_zone
  size              = var.data_volume_gb
  type              = "gp3"
  encrypted         = true

  tags = {
    Name   = "tms-${var.environment}-data"
    Backup = "true" # tag usato da AWS Backup plan in issue #17
  }
}

resource "aws_volume_attachment" "data" {
  device_name                    = "/dev/sdf" # sarà presentato come /dev/nvme1n1 su istanze Nitro
  volume_id                      = aws_ebs_volume.data.id
  instance_id                    = aws_instance.main.id
  stop_instance_before_detaching = true
}

# ── Elastic IP (stabile tra replace istanza) ─────────────────────────────
resource "aws_eip" "main" {
  domain = "vpc"

  tags = {
    Name = "tms-${var.environment}-eip"
  }
}

resource "aws_eip_association" "main" {
  instance_id   = aws_instance.main.id
  allocation_id = aws_eip.main.id
}
