output "public_ip" {
  description = "Elastic IP dell'istanza. Configurarlo come record A su Google DNS."
  value       = aws_eip.main.public_ip
}

output "ssh_command" {
  description = "Comando SSH pronto per l'accesso admin."
  value       = "ssh -i ~/.ssh/tms_admin ubuntu@${aws_eip.main.public_ip}"
}

output "dns_instructions" {
  description = "Istruzioni per configurare il DNS (Google DNS del dominio wheretech.it)."
  value       = <<-EOT

    Configura manualmente su Google DNS, zona `wheretech.it`:

      Record A  ${var.app_fqdn}       → ${aws_eip.main.public_ip}   TTL 300
      Record A  ${var.api_fqdn}       → ${aws_eip.main.public_ip}   TTL 300

    Dopo la propagazione (5-15 min):

      dig +short ${var.app_fqdn}
      # deve restituire ${aws_eip.main.public_ip}

    Poi sulla EC2:

      ssh ubuntu@${aws_eip.main.public_ip}
      sudo certbot certonly --standalone \
           -d ${var.app_fqdn} -d ${var.api_fqdn} \
           -m ${var.admin_email} --agree-tos -n

    Infine scommentare il server block HTTPS in frontend/nginx/default.conf
    e ricaricare nginx: `docker compose restart nginx`.

  EOT
}

output "instance_id" {
  description = "ID istanza EC2 (utile per console, debug, AWS Backup)."
  value       = aws_instance.main.id
}

output "data_volume_id" {
  description = "ID volume EBS dati (usato da AWS Backup in issue #17)."
  value       = aws_ebs_volume.data.id
}

output "invoices_bucket" {
  description = "S3 bucket per fatture (Object Lock COMPLIANCE 10 anni)."
  value       = aws_s3_bucket.invoices.id
}

output "backups_bucket" {
  description = "S3 bucket per dump MongoDB settimanali."
  value       = aws_s3_bucket.backups.id
}

output "backup_vault" {
  description = "AWS Backup vault (snapshot EBS del volume dati)."
  value       = aws_backup_vault.main.name
}

output "alerts_topic_arn" {
  description = "SNS topic per gli alarm CloudWatch. Subscription email da confermare."
  value       = aws_sns_topic.alerts.arn
}

output "cloudwatch_dashboard" {
  description = "URL dashboard CloudWatch."
  value       = "https://${var.region}.console.aws.amazon.com/cloudwatch/home?region=${var.region}#dashboards:name=${aws_cloudwatch_dashboard.main.dashboard_name}"
}
