variable "region" {
  description = "Regione AWS (eu-west-1 Dublino preferita per costo e GDPR)."
  type        = string
  default     = "eu-west-1"
}

variable "environment" {
  description = "Nome ambiente (prod, staging)."
  type        = string
  validation {
    condition     = contains(["prod", "staging", "dev"], var.environment)
    error_message = "environment deve essere uno di: prod, staging, dev."
  }
}

variable "instance_type" {
  description = "Tipo EC2. t3.small è il target MVP (2 vCPU, 2 GB RAM, ~€14/mese)."
  type        = string
  default     = "t3.small"
}

variable "root_volume_gb" {
  description = "Dimensione volume root (OS) in GB. gp3 encrypted."
  type        = number
  default     = 30
}

variable "data_volume_gb" {
  description = "Dimensione volume dati separato (MongoDB, log, backup locale)."
  type        = number
  default     = 20
}

variable "admin_cidr" {
  description = "CIDR del tuo IP personale (/32) per SSH administrativo. Es: \"93.XXX.XXX.XXX/32\"."
  type        = string
  validation {
    condition     = can(regex("^[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+/[0-9]+$", var.admin_cidr))
    error_message = "admin_cidr deve essere in formato CIDR IPv4 (es. 93.147.12.34/32)."
  }
}

variable "ssh_public_key" {
  description = "Contenuto della chiave pubblica SSH admin (output di `cat ~/.ssh/tms_admin.pub`)."
  type        = string
  sensitive   = true
}

variable "deploy_ssh_public_key" {
  description = <<-EOT
    Chiave pubblica SSH usata dalla CI/CD per deployare (issue #18).
    Il cloud-init la installa in /home/deploy/.ssh/authorized_keys con
    restriction `command="..."` per consentire solo il comando di deploy.
  EOT
  type        = string
  sensitive   = true
  default     = "" # può essere aggiunta in una seconda tf apply
}

variable "app_fqdn" {
  description = "FQDN del frontend (es. tms.wheretech.it). Non gestiamo DNS in Terraform: record A su Google DNS a mano."
  type        = string
}

variable "api_fqdn" {
  description = "FQDN delle API (es. api.tms.wheretech.it)."
  type        = string
}

variable "admin_email" {
  description = "Email admin usata da certbot per notifiche rinnovo certificato."
  type        = string
  validation {
    condition     = can(regex("^[^@]+@[^@]+\\.[^@]+$", var.admin_email))
    error_message = "admin_email deve essere un'email valida."
  }
}
