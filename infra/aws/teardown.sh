#!/bin/bash
# infra/aws/teardown.sh — rimuove tutta l'infra creata da provision.sh.
#
# Uso:
#   infra/aws/teardown.sh prod
#
# Ordine di rimozione (inverso rispetto a provision.sh, per rispettare le
# dipendenze): Dashboard → Alarms → SNS → Log Groups → SSM param →
# Backup (selection → plan → vault) → IAM (inline + attached + role +
# instance-profile) → EC2 (disassociate EIP, release, detach volumi,
# terminate instance, delete volumi) → Security Group → subnet + RT +
# IGW + VPC → Key pair → S3 (solo backups; invoices NON cancellabile per
# Object Lock COMPLIANCE 10 anni).
#
# ATTENZIONE:
# - Il bucket `tms-<env>-invoices-<account_id>` NON viene cancellato:
#   Object Lock COMPLIANCE impedisce la rimozione degli oggetti prima
#   della scadenza (10 anni). Anche vuotare il bucket non è possibile.
#   Lo script lascia una nota a schermo.
# - Il bucket `tms-<env>-backups-<account_id>` viene vuotato e cancellato.
# - In prod, `disable-api-termination` deve essere rimosso prima di terminate.
# - Richiede conferma interattiva (`YES`) prima di partire.

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
export REPO_ROOT

# shellcheck source=lib/common.sh
source "$SCRIPT_DIR/lib/common.sh"

ENV="${1:-}"
[ -n "$ENV" ] || die "uso: $0 <env>   es. $0 prod"
export ENV

CONFIG_FILE="$SCRIPT_DIR/config/${ENV}.env"
[ -f "$CONFIG_FILE" ] || die "config file non trovato: $CONFIG_FILE"
# shellcheck source=/dev/null
source "$CONFIG_FILE"
export AWS_DEFAULT_REGION="$AWS_REGION"

check_aws_cli

ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
INVOICES_BUCKET="tms-${ENV}-invoices-${ACCOUNT_ID}"
BACKUPS_BUCKET="tms-${ENV}-backups-${ACCOUNT_ID}"

cat <<EOF

${C_RED}${C_RESET}
${C_RED}██ TEARDOWN env=$ENV ██${C_RESET}

Le risorse seguenti verranno DISTRUTTE (quelle che esistono):
  • VPC + subnet + SG + route table + IGW
  • EC2 instance + EBS volumi (dati inclusi)
  • Elastic IP
  • IAM ruoli + instance profile
  • Key pair SSH admin
  • AWS Backup vault + plan + selection
  • CloudWatch Log Groups + Alarms + Dashboard + SSM parameter
  • SNS topic + subscriptions
  • S3 bucket backups ($BACKUPS_BUCKET) — vuotato e cancellato

NON CANCELLATO (Object Lock COMPLIANCE):
  • S3 bucket invoices ($INVOICES_BUCKET) — resta fino alla scadenza 10a
    degli oggetti (rimane anche vuoto se non hai mai scritto fatture)

EOF

printf "Scrivi ${C_YELLOW}YES${C_RESET} per continuare: "
read -r CONFIRM
[ "$CONFIRM" = "YES" ] || die "operazione annullata"

# Soft-ignore errori: vogliamo procedere anche se singole risorse sono
# già sparite. Tracciamo i fallimenti.
set +e

log "── Dashboard ─────────────────────────────────────────────────────────"
aws cloudwatch delete-dashboards --dashboard-names "tms-$ENV" 2>/dev/null && ok "Dashboard eliminata" || dim "(dashboard non trovata)"

log "── CloudWatch Alarms ─────────────────────────────────────────────────"
ALARMS="tms-$ENV-ec2-status-check tms-$ENV-cpu-high tms-$ENV-disk-root-full tms-$ENV-disk-data-full tms-$ENV-memory-high"
# shellcheck disable=SC2086
aws cloudwatch delete-alarms --alarm-names $ALARMS 2>/dev/null && ok "5 alarm eliminati" || dim "(alarm già mancanti)"

log "── SNS topic ─────────────────────────────────────────────────────────"
TOPIC_ARN=$(aws sns list-topics --query "Topics[?contains(TopicArn,':tms-$ENV-alerts')].TopicArn" --output text)
if [ -n "$TOPIC_ARN" ]; then
    aws sns delete-topic --topic-arn "$TOPIC_ARN" && ok "SNS topic eliminato"
fi

log "── CloudWatch Log Groups ────────────────────────────────────────────"
for lg in "/tms/$ENV/host" "/tms/$ENV/app"; do
    aws logs delete-log-group --log-group-name "$lg" 2>/dev/null && ok "Log group $lg eliminato" || dim "(log group $lg non trovato)"
done

log "── SSM Parameter ────────────────────────────────────────────────────"
aws ssm delete-parameter --name "/tms/$ENV/cloudwatch-agent-config" 2>/dev/null && ok "SSM param eliminato" || dim "(SSM param non trovato)"

log "── AWS Backup ────────────────────────────────────────────────────────"
PLAN_ID=$(aws backup list-backup-plans \
    --query "BackupPlansList[?BackupPlanName=='tms-$ENV-plan'].BackupPlanId | [0]" \
    --output text 2>/dev/null || echo "None")
if [ "$PLAN_ID" != "None" ] && [ -n "$PLAN_ID" ]; then
    SELS=$(aws backup list-backup-selections --backup-plan-id "$PLAN_ID" \
        --query 'BackupSelectionsList[].SelectionId' --output text)
    for sid in $SELS; do
        aws backup delete-backup-selection --backup-plan-id "$PLAN_ID" --selection-id "$sid" 2>/dev/null
    done
    aws backup delete-backup-plan --backup-plan-id "$PLAN_ID" 2>/dev/null && ok "Backup plan + selections eliminati"
fi

# Vault si può eliminare solo se vuoto. Proviamo; se fallisce per recovery
# points esistenti, avvisa.
aws backup delete-backup-vault --backup-vault-name "tms-$ENV-vault" 2>/dev/null \
    && ok "Vault eliminato" \
    || warn "Vault contiene recovery points: cancellali prima con 'aws backup list-recovery-points-by-backup-vault' + delete-recovery-point, poi rilancia il teardown"

log "── EC2: disassociate EIP + terminate instance + delete volumi ──────"
INSTANCE_ID=$(aws ec2 describe-instances \
    --filters "Name=tag:Name,Values=tms-$ENV" \
              "Name=tag:Project,Values=LoginBusiness" \
              "Name=tag:Environment,Values=$ENV" \
              "Name=instance-state-name,Values=pending,running,stopped,stopping" \
    --query 'Reservations[0].Instances[0].InstanceId' --output text 2>/dev/null || echo "None")

if [ "$INSTANCE_ID" != "None" ] && [ -n "$INSTANCE_ID" ]; then
    if [ "$ENV" = "prod" ]; then
        aws ec2 modify-instance-attribute --instance-id "$INSTANCE_ID" --no-disable-api-termination 2>/dev/null
    fi
    # Disassociate EIP
    EIP_ASSOC=$(aws ec2 describe-addresses \
        --filters "Name=instance-id,Values=$INSTANCE_ID" \
        --query 'Addresses[0].AssociationId' --output text 2>/dev/null || echo "None")
    if [ "$EIP_ASSOC" != "None" ] && [ -n "$EIP_ASSOC" ]; then
        aws ec2 disassociate-address --association-id "$EIP_ASSOC"
    fi

    aws ec2 terminate-instances --instance-ids "$INSTANCE_ID" >/dev/null
    log "attendo terminate..."
    aws ec2 wait instance-terminated --instance-ids "$INSTANCE_ID"
    ok "Instance $INSTANCE_ID terminata"
fi

# Volume dati (se non auto-deleted)
DATA_VOLUME_ID=$(aws ec2 describe-volumes \
    --filters "Name=tag:Name,Values=tms-$ENV-data" \
              "Name=tag:Project,Values=LoginBusiness" \
              "Name=tag:Environment,Values=$ENV" \
              "Name=status,Values=available" \
    --query 'Volumes[0].VolumeId' --output text 2>/dev/null || echo "None")
if [ "$DATA_VOLUME_ID" != "None" ] && [ -n "$DATA_VOLUME_ID" ]; then
    aws ec2 delete-volume --volume-id "$DATA_VOLUME_ID" && ok "EBS dati $DATA_VOLUME_ID eliminato"
fi

# Release EIP
EIP_ALLOC=$(aws ec2 describe-addresses \
    --filters "Name=tag:Name,Values=tms-$ENV-eip" \
              "Name=tag:Project,Values=LoginBusiness" \
              "Name=tag:Environment,Values=$ENV" \
    --query 'Addresses[0].AllocationId' --output text 2>/dev/null || echo "None")
if [ "$EIP_ALLOC" != "None" ] && [ -n "$EIP_ALLOC" ]; then
    aws ec2 release-address --allocation-id "$EIP_ALLOC" && ok "Elastic IP rilasciato"
fi

log "── Key pair ─────────────────────────────────────────────────────────"
aws ec2 delete-key-pair --key-name "tms-$ENV-admin" 2>/dev/null && ok "Key pair eliminata" || dim "(key pair non trovata)"

log "── Security Group ───────────────────────────────────────────────────"
VPC_ID=$(aws ec2 describe-vpcs \
    --filters "Name=tag:Name,Values=tms-$ENV" \
              "Name=tag:Project,Values=LoginBusiness" \
              "Name=tag:Environment,Values=$ENV" \
    --query 'Vpcs[0].VpcId' --output text 2>/dev/null || echo "None")

if [ "$VPC_ID" != "None" ] && [ -n "$VPC_ID" ]; then
    SG_ID=$(aws ec2 describe-security-groups \
        --filters "Name=vpc-id,Values=$VPC_ID" "Name=group-name,Values=tms-$ENV-ec2" \
        --query 'SecurityGroups[0].GroupId' --output text 2>/dev/null || echo "None")
    if [ "$SG_ID" != "None" ] && [ -n "$SG_ID" ]; then
        aws ec2 delete-security-group --group-id "$SG_ID" && ok "SG eliminato"
    fi

    log "── Route tables, subnet, IGW, VPC ────────────────────────────────"
    # Subnet
    SUBNET_ID=$(aws ec2 describe-subnets \
        --filters "Name=vpc-id,Values=$VPC_ID" "Name=tag:Name,Values=tms-$ENV-public-a" \
        --query 'Subnets[0].SubnetId' --output text 2>/dev/null || echo "None")
    if [ "$SUBNET_ID" != "None" ] && [ -n "$SUBNET_ID" ]; then
        aws ec2 delete-subnet --subnet-id "$SUBNET_ID" && ok "Subnet eliminata"
    fi

    # Route table non-main
    RTB_ID=$(aws ec2 describe-route-tables \
        --filters "Name=vpc-id,Values=$VPC_ID" "Name=tag:Name,Values=tms-$ENV-rt-public" \
        --query 'RouteTables[0].RouteTableId' --output text 2>/dev/null || echo "None")
    if [ "$RTB_ID" != "None" ] && [ -n "$RTB_ID" ]; then
        aws ec2 delete-route-table --route-table-id "$RTB_ID" && ok "Route table eliminata"
    fi

    # IGW
    IGW_ID=$(aws ec2 describe-internet-gateways \
        --filters "Name=attachment.vpc-id,Values=$VPC_ID" \
        --query 'InternetGateways[0].InternetGatewayId' --output text 2>/dev/null || echo "None")
    if [ "$IGW_ID" != "None" ] && [ -n "$IGW_ID" ]; then
        aws ec2 detach-internet-gateway --internet-gateway-id "$IGW_ID" --vpc-id "$VPC_ID"
        aws ec2 delete-internet-gateway --internet-gateway-id "$IGW_ID" && ok "IGW eliminato"
    fi

    aws ec2 delete-vpc --vpc-id "$VPC_ID" && ok "VPC eliminato"
fi

log "── IAM (ruoli + instance profile) ───────────────────────────────────"
# EC2 instance profile
aws iam remove-role-from-instance-profile --instance-profile-name "tms-$ENV-ec2" --role-name "tms-$ENV-ec2" 2>/dev/null
aws iam delete-instance-profile --instance-profile-name "tms-$ENV-ec2" 2>/dev/null && ok "Instance profile eliminato"

# EC2 role policies
aws iam delete-role-policy --role-name "tms-$ENV-ec2" --policy-name "s3-access" 2>/dev/null
aws iam detach-role-policy --role-name "tms-$ENV-ec2" --policy-arn arn:aws:iam::aws:policy/CloudWatchAgentServerPolicy 2>/dev/null
aws iam delete-role --role-name "tms-$ENV-ec2" 2>/dev/null && ok "IAM role EC2 eliminato"

# Backup role
aws iam detach-role-policy --role-name "tms-$ENV-aws-backup" --policy-arn arn:aws:iam::aws:policy/service-role/AWSBackupServiceRolePolicyForBackup 2>/dev/null
aws iam detach-role-policy --role-name "tms-$ENV-aws-backup" --policy-arn arn:aws:iam::aws:policy/service-role/AWSBackupServiceRolePolicyForRestores 2>/dev/null
aws iam delete-role --role-name "tms-$ENV-aws-backup" 2>/dev/null && ok "IAM role Backup eliminato"

log "── S3 bucket backups (vuoto + delete) ───────────────────────────────"
if aws s3api head-bucket --bucket "$BACKUPS_BUCKET" 2>/dev/null; then
    aws s3api delete-objects --bucket "$BACKUPS_BUCKET" \
        --delete "$(aws s3api list-object-versions --bucket "$BACKUPS_BUCKET" \
            --output json --query '{Objects: Versions[].{Key:Key,VersionId:VersionId}}' 2>/dev/null)" 2>/dev/null || true
    aws s3api delete-objects --bucket "$BACKUPS_BUCKET" \
        --delete "$(aws s3api list-object-versions --bucket "$BACKUPS_BUCKET" \
            --output json --query '{Objects: DeleteMarkers[].{Key:Key,VersionId:VersionId}}' 2>/dev/null)" 2>/dev/null || true
    aws s3 rb "s3://$BACKUPS_BUCKET" --force && ok "S3 backups eliminato"
fi

log "── S3 bucket invoices ────────────────────────────────────────────────"
if aws s3api head-bucket --bucket "$INVOICES_BUCKET" 2>/dev/null; then
    warn "S3 $INVOICES_BUCKET NON può essere cancellato: Object Lock COMPLIANCE"
    warn "Resterà attivo (~€0.02/mese se vuoto) fino a che tutti gli oggetti scadono"
fi

log "── State locale ──────────────────────────────────────────────────────"
state_clear
ok "State file rimosso"

printf "\n${C_GREEN}████ TEARDOWN COMPLETATO ████${C_RESET}\n\n"
