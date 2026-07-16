# Backup a due livelli (issue #17):
#
# 1. **AWS Backup** per snapshot EBS del volume dati giornaliero,
#    retention 7 giorni + uno weekly con retention 28 giorni. RTO ~30 min
#    (crea nuovo volume da snapshot), RPO 24h.
#
# 2. **S3 tms-{env}-backups** per dump MongoDB settimanali pushati dal
#    cron sull'host (script in scripts/backup_mongo.sh). Lifecycle transita
#    i dump a Glacier Deep Archive dopo 1 anno.
#
# 3. **S3 tms-{env}-invoices** per PDF fatture (quando #41 li genera).
#    Object Lock COMPLIANCE 10 anni (obbligo fiscale italiano — DPR 600
#    art. 22 + CAD). Lifecycle: Standard → Standard-IA dopo 90 gg →
#    Glacier Deep Archive dopo 1 anno.
#
# 4. **IAM Instance Profile** agganciato alla EC2: consente all'host di
#    caricare dump su S3 e scrivere metriche CloudWatch (issue #19) senza
#    access key statiche.

# ── Random suffix per nomi bucket globali S3 ─────────────────────────────
# I nomi S3 sono globali e devono essere unici. Usiamo l'account ID per
# garantire univocità senza dover coordinare a mano.
data "aws_caller_identity" "current" {}

locals {
  account_id      = data.aws_caller_identity.current.account_id
  invoices_bucket = "tms-${var.environment}-invoices-${local.account_id}"
  backups_bucket  = "tms-${var.environment}-backups-${local.account_id}"
}

# ═════════════════════════════════════════════════════════════════════════
# S3 — Bucket invoices (Object Lock COMPLIANCE 10 anni)
# ═════════════════════════════════════════════════════════════════════════

resource "aws_s3_bucket" "invoices" {
  bucket = local.invoices_bucket

  # Object Lock deve essere abilitato alla creazione del bucket; non è
  # retroattivo. Una volta abilitato non può essere disabilitato.
  object_lock_enabled = true

  tags = {
    Name    = local.invoices_bucket
    Purpose = "fatture-definitive-10y-retention"
  }
}

resource "aws_s3_bucket_versioning" "invoices" {
  bucket = aws_s3_bucket.invoices.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "invoices" {
  bucket = aws_s3_bucket.invoices.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "invoices" {
  bucket                  = aws_s3_bucket.invoices.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Object Lock: 10 anni COMPLIANCE mode. In COMPLIANCE nessuno — nemmeno root —
# può cancellare o modificare l'oggetto prima della scadenza. Coerente con
# l'obbligo fiscale di 10 anni.
resource "aws_s3_bucket_object_lock_configuration" "invoices" {
  bucket = aws_s3_bucket.invoices.id

  rule {
    default_retention {
      mode = "COMPLIANCE"
      days = 10 * 365 # 3650 giorni = 10 anni
    }
  }
}

# Lifecycle: dopo 90 gg sposta a Standard-IA (accesso infrequente) per
# ridurre il costo di storage; dopo 1 anno a Glacier Deep Archive (~€0.001/GB).
resource "aws_s3_bucket_lifecycle_configuration" "invoices" {
  bucket = aws_s3_bucket.invoices.id

  rule {
    id     = "transition-to-cold-storage"
    status = "Enabled"

    filter {} # si applica a tutti gli oggetti del bucket

    transition {
      days          = 90
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 365
      storage_class = "DEEP_ARCHIVE"
    }
  }
}

# ═════════════════════════════════════════════════════════════════════════
# S3 — Bucket backups (dump MongoDB settimanali)
# ═════════════════════════════════════════════════════════════════════════

resource "aws_s3_bucket" "backups" {
  bucket = local.backups_bucket

  tags = {
    Name    = local.backups_bucket
    Purpose = "mongodb-weekly-dumps"
  }
}

resource "aws_s3_bucket_versioning" "backups" {
  bucket = aws_s3_bucket.backups.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "backups" {
  bucket = aws_s3_bucket.backups.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "backups" {
  bucket                  = aws_s3_bucket.backups.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_lifecycle_configuration" "backups" {
  bucket = aws_s3_bucket.backups.id

  rule {
    id     = "expire-old-backups"
    status = "Enabled"

    filter {}

    # Dopo 30 giorni già su storage economico (la retention normale è
    # coperta da AWS Backup EBS).
    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 120
      storage_class = "DEEP_ARCHIVE"
    }

    # Cancellazione dopo 1 anno: i dump MongoDB non servono oltre.
    expiration {
      days = 365
    }

    noncurrent_version_expiration {
      noncurrent_days = 30
    }
  }
}

# ═════════════════════════════════════════════════════════════════════════
# AWS Backup — snapshot EBS giornaliero del volume dati
# ═════════════════════════════════════════════════════════════════════════

resource "aws_backup_vault" "main" {
  name = "tms-${var.environment}-vault"
}

# IAM role usato da AWS Backup per creare/ripristinare snapshot.
# Policy managed `AWSBackupServiceRolePolicyForBackup` è sufficiente.
resource "aws_iam_role" "backup" {
  name = "tms-${var.environment}-aws-backup"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "backup.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "backup" {
  role       = aws_iam_role.backup.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSBackupServiceRolePolicyForBackup"
}

resource "aws_iam_role_policy_attachment" "backup_restore" {
  role       = aws_iam_role.backup.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSBackupServiceRolePolicyForRestores"
}

resource "aws_backup_plan" "main" {
  name = "tms-${var.environment}-plan"

  rule {
    rule_name         = "daily-7d"
    target_vault_name = aws_backup_vault.main.name
    schedule          = "cron(0 3 * * ? *)" # 03:00 UTC ogni giorno
    start_window      = 60
    completion_window = 180

    lifecycle {
      delete_after = 7 # retention 7 giorni
    }
  }

  rule {
    rule_name         = "weekly-28d"
    target_vault_name = aws_backup_vault.main.name
    schedule          = "cron(0 3 ? * SUN *)" # domenica alle 03:00 UTC
    start_window      = 60
    completion_window = 360

    lifecycle {
      delete_after = 28
    }
  }
}

# Selezione risorse: tutti i volumi EBS con tag Backup=true (il volume dati
# creato in main.tf ha quel tag — root volume NON ha il tag quindi esente).
resource "aws_backup_selection" "ebs_data" {
  name         = "tms-${var.environment}-ebs-data"
  iam_role_arn = aws_iam_role.backup.arn
  plan_id      = aws_backup_plan.main.id

  selection_tag {
    type  = "STRINGEQUALS"
    key   = "Backup"
    value = "true"
  }
}

# ═════════════════════════════════════════════════════════════════════════
# IAM Instance Profile per EC2 (usato da backup_mongo.sh + CloudWatch Agent)
# ═════════════════════════════════════════════════════════════════════════

resource "aws_iam_role" "ec2" {
  name = "tms-${var.environment}-ec2"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

# S3 write su bucket specifici (dump mongo + upload fatture quando #41 li
# genererà lato backend con boto3).
resource "aws_iam_role_policy" "ec2_s3" {
  name = "s3-access"
  role = aws_iam_role.ec2.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:PutObject",
          "s3:PutObjectAcl",
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.backups.arn,
          "${aws_s3_bucket.backups.arn}/*",
          aws_s3_bucket.invoices.arn,
          "${aws_s3_bucket.invoices.arn}/*"
        ]
      }
    ]
  })
}

# CloudWatch PutMetric e Logs (usati in issue #19).
resource "aws_iam_role_policy_attachment" "ec2_cw_agent" {
  role       = aws_iam_role.ec2.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchAgentServerPolicy"
}

resource "aws_iam_instance_profile" "ec2" {
  name = "tms-${var.environment}-ec2"
  role = aws_iam_role.ec2.name
}
