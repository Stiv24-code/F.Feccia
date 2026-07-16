# CloudWatch monitoring (issue #19):
#
# 1. **SSM Parameter Store** con la config JSON del CloudWatch Agent
#    (l'host la legge con `amazon-cloudwatch-agent-ctl -fetch-config
#    -s ssm:<name>` — più pulito del copiarla inline nel user_data).
# 2. **Log group** centralizzato per syslog + docker logs + app.
# 3. **SNS topic** con subscription email (richiede conferma manuale).
# 4. **Alarms** su metriche standard CloudWatch + custom CWAgent:
#    - EC2 Status Check failed
#    - CPU > 80% sostenuto
#    - Disco root > 85%
#    - Disco /data > 85%
#    - Memoria > 85%

variable "alert_email" {
  description = "Email per ricevere gli alarm (CloudWatch → SNS)."
  type        = string
  default     = "" # vuoto = non crea subscription (da aggiungere a mano)
}

# ─── Log group ─────────────────────────────────────────────────────────
resource "aws_cloudwatch_log_group" "host" {
  name              = "/tms/${var.environment}/host"
  retention_in_days = var.environment == "prod" ? 30 : 7

  tags = {
    Name = "tms-${var.environment}-host-logs"
  }
}

# ─── SSM Parameter con config CloudWatch Agent ─────────────────────────
# L'agent legge la config da qui all'avvio. Modifiche alla config non
# richiedono re-provisioning dell'istanza: basta `systemctl restart
# amazon-cloudwatch-agent` dopo un `terraform apply`.
resource "aws_ssm_parameter" "cw_agent_config" {
  name = "/tms/${var.environment}/cloudwatch-agent-config"
  type = "String"
  tier = "Standard"

  value = jsonencode({
    agent = {
      metrics_collection_interval = 60
      run_as_user                 = "cwagent"
    }
    metrics = {
      namespace = "TMS/${var.environment}"
      append_dimensions = {
        InstanceId = "$${aws:InstanceId}"
      }
      metrics_collected = {
        cpu = {
          measurement                 = ["cpu_usage_idle", "cpu_usage_iowait"]
          metrics_collection_interval = 60
          totalcpu                    = true
        }
        mem = {
          measurement                 = ["mem_used_percent"]
          metrics_collection_interval = 60
        }
        disk = {
          measurement                 = ["used_percent"]
          metrics_collection_interval = 60
          resources                   = ["/", "/data"]
          drop_device                 = true
        }
        swap = {
          measurement                 = ["swap_used_percent"]
          metrics_collection_interval = 60
        }
      }
    }
    logs = {
      logs_collected = {
        files = {
          collect_list = [
            {
              file_path       = "/var/log/syslog"
              log_group_name  = aws_cloudwatch_log_group.host.name
              log_stream_name = "{instance_id}/syslog"
              timezone        = "UTC"
            },
            {
              file_path       = "/var/log/auth.log"
              log_group_name  = aws_cloudwatch_log_group.host.name
              log_stream_name = "{instance_id}/auth"
              timezone        = "UTC"
            },
            {
              file_path       = "/var/log/tms-user-data.log"
              log_group_name  = aws_cloudwatch_log_group.host.name
              log_stream_name = "{instance_id}/tms-user-data"
              timezone        = "UTC"
            },
            {
              file_path       = "/var/log/tms-deploy.log"
              log_group_name  = aws_cloudwatch_log_group.host.name
              log_stream_name = "{instance_id}/tms-deploy"
              timezone        = "UTC"
            },
            {
              file_path       = "/var/log/tms-backup-mongo.log"
              log_group_name  = aws_cloudwatch_log_group.host.name
              log_stream_name = "{instance_id}/tms-backup"
              timezone        = "UTC"
            }
          ]
        }
      }
    }
  })
}

# ─── SNS topic per gli alarm ───────────────────────────────────────────
resource "aws_sns_topic" "alerts" {
  name = "tms-${var.environment}-alerts"

  tags = {
    Name = "tms-${var.environment}-alerts"
  }
}

# Subscription email opzionale. Se alert_email != "", Terraform crea la
# subscription; AWS manda un'email di conferma che va cliccata a mano
# prima che gli alarm arrivino.
resource "aws_sns_topic_subscription" "alerts_email" {
  count     = var.alert_email != "" ? 1 : 0
  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "email"
  endpoint  = var.alert_email
}

# ─── Alarms ─────────────────────────────────────────────────────────────

# 1. EC2 status check (AWS-side, infra): 1 datapoint di fail ≙ problema VM.
resource "aws_cloudwatch_metric_alarm" "status_check" {
  alarm_name          = "tms-${var.environment}-ec2-status-check"
  alarm_description   = "EC2 status check failed (host o system problem)"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 2
  metric_name         = "StatusCheckFailed"
  namespace           = "AWS/EC2"
  period              = 60
  statistic           = "Maximum"
  threshold           = 1
  treat_missing_data  = "breaching"

  dimensions = {
    InstanceId = aws_instance.main.id
  }

  alarm_actions = [aws_sns_topic.alerts.arn]
  ok_actions    = [aws_sns_topic.alerts.arn]
}

# 2. CPU alto sostenuto: > 80% per 10 minuti consecutivi.
resource "aws_cloudwatch_metric_alarm" "cpu_high" {
  alarm_name          = "tms-${var.environment}-cpu-high"
  alarm_description   = "CPU > 80% per 10 min. Potenziale DoS, worker bloccati o load non sostenibile."
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 10
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = 60
  statistic           = "Average"
  threshold           = 80

  dimensions = {
    InstanceId = aws_instance.main.id
  }

  alarm_actions = [aws_sns_topic.alerts.arn]
  ok_actions    = [aws_sns_topic.alerts.arn]
}

# 3. Disco root pieno (CWAgent metric). Sopra 85% è tempo di intervenire.
resource "aws_cloudwatch_metric_alarm" "disk_root_full" {
  alarm_name          = "tms-${var.environment}-disk-root-full"
  alarm_description   = "Disk / (root OS) > 85%. Docker layer, log, snap — ripulire."
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "disk_used_percent"
  namespace           = "TMS/${var.environment}"
  period              = 300
  statistic           = "Average"
  threshold           = 85
  treat_missing_data  = "notBreaching"

  dimensions = {
    InstanceId = aws_instance.main.id
    path       = "/"
  }

  alarm_actions = [aws_sns_topic.alerts.arn]
  ok_actions    = [aws_sns_topic.alerts.arn]
}

# 4. Disco dati pieno (MongoDB + backup locali).
resource "aws_cloudwatch_metric_alarm" "disk_data_full" {
  alarm_name          = "tms-${var.environment}-disk-data-full"
  alarm_description   = "Disk /data > 85%. MongoDB in crescita rapida, valutare cleanup o resize EBS."
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "disk_used_percent"
  namespace           = "TMS/${var.environment}"
  period              = 300
  statistic           = "Average"
  threshold           = 85
  treat_missing_data  = "notBreaching"

  dimensions = {
    InstanceId = aws_instance.main.id
    path       = "/data"
  }

  alarm_actions = [aws_sns_topic.alerts.arn]
  ok_actions    = [aws_sns_topic.alerts.arn]
}

# 5. Memoria alta sostenuta.
resource "aws_cloudwatch_metric_alarm" "memory_high" {
  alarm_name          = "tms-${var.environment}-memory-high"
  alarm_description   = "Memoria > 85% per 10 min. OOM Killer rischia di terminare container."
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 10
  metric_name         = "mem_used_percent"
  namespace           = "TMS/${var.environment}"
  period              = 60
  statistic           = "Average"
  threshold           = 85
  treat_missing_data  = "notBreaching"

  dimensions = {
    InstanceId = aws_instance.main.id
  }

  alarm_actions = [aws_sns_topic.alerts.arn]
  ok_actions    = [aws_sns_topic.alerts.arn]
}

# ─── Log group retention aggiuntivi (audit/app) ─────────────────────────
# L'app FastAPI stamperà log strutturati su stdout (issue #22, structlog);
# Docker li raccoglie in json-file driver, CloudWatch Agent li ha già in
# /var/lib/docker/containers/*/*.log che possiamo aggiungere al collect_list
# in un follow-up. Per ora il log group esiste pronto all'uso.
resource "aws_cloudwatch_log_group" "app" {
  name              = "/tms/${var.environment}/app"
  retention_in_days = var.environment == "prod" ? 30 : 7
}

# ─── Dashboard CloudWatch (opzionale ma utile) ─────────────────────────
resource "aws_cloudwatch_dashboard" "main" {
  dashboard_name = "tms-${var.environment}"

  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 12
        height = 6
        properties = {
          metrics = [
            ["AWS/EC2", "CPUUtilization", "InstanceId", aws_instance.main.id],
            ["TMS/${var.environment}", "mem_used_percent", "InstanceId", aws_instance.main.id]
          ]
          period = 60
          stat   = "Average"
          region = var.region
          title  = "CPU + Memory"
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 0
        width  = 12
        height = 6
        properties = {
          metrics = [
            ["TMS/${var.environment}", "disk_used_percent", "InstanceId", aws_instance.main.id, "path", "/"],
            [".", ".", ".", ".", ".", "/data"]
          ]
          period = 300
          stat   = "Average"
          region = var.region
          title  = "Disk usage"
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 6
        width  = 12
        height = 6
        properties = {
          metrics = [
            ["AWS/EC2", "NetworkIn", "InstanceId", aws_instance.main.id],
            [".", "NetworkOut", ".", "."]
          ]
          period = 60
          stat   = "Sum"
          region = var.region
          title  = "Network in/out"
        }
      }
    ]
  })
}
