#!/bin/bash
# infra/aws/provision.sh — provisioning idempotente dell'infra AWS.
#
# Uso:
#   cp infra/aws/config/prod.env.example infra/aws/config/prod.env
#   # edita prod.env
#   infra/aws/provision.sh prod
#
# Lo script è idempotente: rieseguirlo dopo il primo apply NON crea duplicati,
# riconosce cosa esiste già tramite tag (Project=LoginBusiness + Environment=X
# + Name=<resource-name>).
#
# Gli ID delle risorse create vengono scritti in `infra/aws/.aws-state/<env>.env`
# (gitignored) per uso rapido + teardown.
#
# Sezioni eseguite in quest'ordine (riflettono dipendenze):
#   1. Precondizioni (aws-cli, env file, chiavi SSH)
#   2. VPC + subnet + IGW + route table + Security Group
#   3. IAM (role EC2, instance profile, role AWS Backup)
#   4. S3 buckets (invoices con Object Lock, backups)
#   5. KeyPair SSH admin
#   6. EC2 instance + EBS data volume + Elastic IP
#   7. AWS Backup vault + plan + selection
#   8. SSM Parameter CloudWatch Agent config
#   9. CloudWatch Log Groups
#  10. SNS topic + subscription email + 5 Alarms + Dashboard

set -euo pipefail

# Su MSYS/Git Bash (Windows) qualsiasi argomento che inizia con `/` viene
# convertito automaticamente in un path Windows: `/dev/sdf` → `C:/Program
# Files/Git/dev/sdf`, ecc. Per aws-cli questo rompe parametri come
# `--device /dev/sdf`, `--role-arn arn:aws:iam::...:role/xxx`, ARN con
# slash, path S3, URL HTTP con path leading slash. Disabilitiamo la
# conversione per tutti gli argomenti. I path che *dobbiamo* passare come
# path Windows (es. fileb://... per file locali) continuano a essere
# gestiti esplicitamente con `win_path`.
export MSYS_NO_PATHCONV=1
export MSYS2_ARG_CONV_EXCL='*'

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
export REPO_ROOT

# shellcheck source=lib/common.sh
source "$SCRIPT_DIR/lib/common.sh"

# ═════════════════════════════════════════════════════════════════════════
# 1. PRECONDIZIONI
# ═════════════════════════════════════════════════════════════════════════

ENV="${1:-}"
[ -n "$ENV" ] || die "uso: $0 <env>   es. $0 prod"

CONFIG_FILE="$SCRIPT_DIR/config/${ENV}.env"
[ -f "$CONFIG_FILE" ] || die "config file non trovato: $CONFIG_FILE
cp $CONFIG_FILE.example $CONFIG_FILE
# poi edita i valori"

# shellcheck source=/dev/null
source "$CONFIG_FILE"
export ENV

require_var ENV
require_var AWS_REGION
require_var INSTANCE_TYPE
require_var ADMIN_CIDR
require_var ADMIN_SSH_PUBKEY_FILE
require_var APP_FQDN
require_var API_FQDN
require_var ADMIN_EMAIL

[ -f "$ADMIN_SSH_PUBKEY_FILE" ] || die "ADMIN_SSH_PUBKEY_FILE non trovato: $ADMIN_SSH_PUBKEY_FILE"
ADMIN_SSH_PUBKEY=$(cat "$ADMIN_SSH_PUBKEY_FILE")

DEPLOY_SSH_PUBKEY=""
if [ -n "${DEPLOY_SSH_PUBKEY_FILE:-}" ] && [ -f "$DEPLOY_SSH_PUBKEY_FILE" ]; then
    DEPLOY_SSH_PUBKEY=$(cat "$DEPLOY_SSH_PUBKEY_FILE")
fi

export AWS_DEFAULT_REGION="$AWS_REGION"

check_aws_cli
log "Provisioning env=$ENV region=$AWS_REGION"
log "State file: $(state_file)"

ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
state_set ACCOUNT_ID "$ACCOUNT_ID"

# Validazione CIDR admin
[[ "$ADMIN_CIDR" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+/[0-9]+$ ]] \
    || die "ADMIN_CIDR deve essere in formato CIDR IPv4 (es. 93.147.12.34/32)"

# ═════════════════════════════════════════════════════════════════════════
# HELPER: idempotent_create_*
# ═════════════════════════════════════════════════════════════════════════

STD_TAGS="Key=Project,Value=LoginBusiness Key=Environment,Value=$ENV Key=ManagedBy,Value=aws-cli-script"

# ═════════════════════════════════════════════════════════════════════════
# 2. NETWORKING: VPC + subnet + IGW + route table + Security Group
# ═════════════════════════════════════════════════════════════════════════

log "── 2. Networking ─────────────────────────────────────────────────────"

# --- VPC ---
VPC_ID=$(aws ec2 describe-vpcs \
    --filters "Name=tag:Project,Values=LoginBusiness" \
              "Name=tag:Environment,Values=$ENV" \
              "Name=tag:Name,Values=tms-$ENV" \
    --query 'Vpcs[0].VpcId' --output text 2>/dev/null || echo "None")

if [ "$VPC_ID" = "None" ] || [ -z "$VPC_ID" ]; then
    VPC_ID=$(aws ec2 create-vpc \
        --cidr-block 10.0.0.0/16 \
        --tag-specifications "ResourceType=vpc,Tags=[{Key=Name,Value=tms-$ENV},{Key=Project,Value=LoginBusiness},{Key=Environment,Value=$ENV}]" \
        --query 'Vpc.VpcId' --output text)
    aws ec2 modify-vpc-attribute --vpc-id "$VPC_ID" --enable-dns-support
    aws ec2 modify-vpc-attribute --vpc-id "$VPC_ID" --enable-dns-hostnames
    ok "Creato VPC $VPC_ID"
else
    ok "VPC esistente: $VPC_ID"
fi
state_set VPC_ID "$VPC_ID"

# --- Internet Gateway ---
IGW_ID=$(aws ec2 describe-internet-gateways \
    --filters "Name=attachment.vpc-id,Values=$VPC_ID" \
    --query 'InternetGateways[0].InternetGatewayId' --output text 2>/dev/null || echo "None")

if [ "$IGW_ID" = "None" ] || [ -z "$IGW_ID" ]; then
    IGW_ID=$(aws ec2 create-internet-gateway \
        --tag-specifications "ResourceType=internet-gateway,Tags=[{Key=Name,Value=tms-$ENV-igw},{Key=Project,Value=LoginBusiness},{Key=Environment,Value=$ENV}]" \
        --query 'InternetGateway.InternetGatewayId' --output text)
    aws ec2 attach-internet-gateway --internet-gateway-id "$IGW_ID" --vpc-id "$VPC_ID"
    ok "Creato IGW $IGW_ID e attaccato a VPC $VPC_ID"
else
    ok "IGW esistente: $IGW_ID"
fi
state_set IGW_ID "$IGW_ID"

# --- Subnet pubblica ---
SUBNET_ID=$(aws ec2 describe-subnets \
    --filters "Name=vpc-id,Values=$VPC_ID" \
              "Name=tag:Name,Values=tms-$ENV-public-a" \
    --query 'Subnets[0].SubnetId' --output text 2>/dev/null || echo "None")

if [ "$SUBNET_ID" = "None" ] || [ -z "$SUBNET_ID" ]; then
    SUBNET_ID=$(aws ec2 create-subnet \
        --vpc-id "$VPC_ID" \
        --cidr-block 10.0.1.0/24 \
        --availability-zone "${AWS_REGION}a" \
        --tag-specifications "ResourceType=subnet,Tags=[{Key=Name,Value=tms-$ENV-public-a},{Key=Project,Value=LoginBusiness},{Key=Environment,Value=$ENV}]" \
        --query 'Subnet.SubnetId' --output text)
    aws ec2 modify-subnet-attribute --subnet-id "$SUBNET_ID" --map-public-ip-on-launch
    ok "Creato subnet pubblica $SUBNET_ID"
else
    ok "Subnet esistente: $SUBNET_ID"
fi
state_set SUBNET_ID "$SUBNET_ID"

# --- Route table + default route ---
RTB_ID=$(aws ec2 describe-route-tables \
    --filters "Name=vpc-id,Values=$VPC_ID" \
              "Name=tag:Name,Values=tms-$ENV-rt-public" \
    --query 'RouteTables[0].RouteTableId' --output text 2>/dev/null || echo "None")

if [ "$RTB_ID" = "None" ] || [ -z "$RTB_ID" ]; then
    RTB_ID=$(aws ec2 create-route-table \
        --vpc-id "$VPC_ID" \
        --tag-specifications "ResourceType=route-table,Tags=[{Key=Name,Value=tms-$ENV-rt-public},{Key=Project,Value=LoginBusiness},{Key=Environment,Value=$ENV}]" \
        --query 'RouteTable.RouteTableId' --output text)
    aws ec2 create-route --route-table-id "$RTB_ID" --destination-cidr-block 0.0.0.0/0 --gateway-id "$IGW_ID" >/dev/null
    aws ec2 associate-route-table --route-table-id "$RTB_ID" --subnet-id "$SUBNET_ID" >/dev/null
    ok "Creata route table $RTB_ID con default route via IGW"
else
    ok "Route table esistente: $RTB_ID"
fi
state_set RTB_ID "$RTB_ID"

# --- Security Group ---
SG_ID=$(aws ec2 describe-security-groups \
    --filters "Name=vpc-id,Values=$VPC_ID" \
              "Name=group-name,Values=tms-$ENV-ec2" \
    --query 'SecurityGroups[0].GroupId' --output text 2>/dev/null || echo "None")

if [ "$SG_ID" = "None" ] || [ -z "$SG_ID" ]; then
    SG_ID=$(aws ec2 create-security-group \
        --vpc-id "$VPC_ID" \
        --group-name "tms-$ENV-ec2" \
        --description "Firewall TMS: SSH admin + HTTP + HTTPS" \
        --tag-specifications "ResourceType=security-group,Tags=[{Key=Name,Value=tms-$ENV-ec2},{Key=Project,Value=LoginBusiness},{Key=Environment,Value=$ENV}]" \
        --query 'GroupId' --output text)

    # Ingress SSH da admin CIDR
    aws ec2 authorize-security-group-ingress \
        --group-id "$SG_ID" \
        --ip-permissions "IpProtocol=tcp,FromPort=22,ToPort=22,IpRanges=[{CidrIp=$ADMIN_CIDR,Description=SSH admin}]" \
        >/dev/null
    # HTTP open (ACME HTTP-01 challenge + redirect a HTTPS)
    aws ec2 authorize-security-group-ingress \
        --group-id "$SG_ID" \
        --ip-permissions 'IpProtocol=tcp,FromPort=80,ToPort=80,IpRanges=[{CidrIp=0.0.0.0/0,Description=HTTP}]' \
        >/dev/null
    # HTTPS open
    aws ec2 authorize-security-group-ingress \
        --group-id "$SG_ID" \
        --ip-permissions 'IpProtocol=tcp,FromPort=443,ToPort=443,IpRanges=[{CidrIp=0.0.0.0/0,Description=HTTPS}]' \
        >/dev/null
    ok "Creato Security Group $SG_ID con regole SSH/80/443"
else
    ok "Security Group esistente: $SG_ID"
fi
state_set SG_ID "$SG_ID"

# ═════════════════════════════════════════════════════════════════════════
# 3. IAM — role EC2 (Instance Profile) + role AWS Backup
# ═════════════════════════════════════════════════════════════════════════

log "── 3. IAM ────────────────────────────────────────────────────────────"

EC2_ROLE="tms-$ENV-ec2"
EC2_PROFILE="tms-$ENV-ec2"

if ! aws iam get-role --role-name "$EC2_ROLE" >/dev/null 2>&1; then
    aws iam create-role \
        --role-name "$EC2_ROLE" \
        --assume-role-policy-document '{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"ec2.amazonaws.com"},"Action":"sts:AssumeRole"}]}' \
        --tags "Key=Project,Value=LoginBusiness" "Key=Environment,Value=$ENV" \
        >/dev/null
    ok "Creato IAM role $EC2_ROLE"
else
    ok "IAM role esistente: $EC2_ROLE"
fi

# CloudWatch Agent managed policy
aws iam attach-role-policy \
    --role-name "$EC2_ROLE" \
    --policy-arn arn:aws:iam::aws:policy/CloudWatchAgentServerPolicy \
    >/dev/null 2>&1 || true
# (attach-role-policy è idempotente per nature)

# Instance Profile
if ! aws iam get-instance-profile --instance-profile-name "$EC2_PROFILE" >/dev/null 2>&1; then
    aws iam create-instance-profile --instance-profile-name "$EC2_PROFILE" >/dev/null
    aws iam add-role-to-instance-profile --instance-profile-name "$EC2_PROFILE" --role-name "$EC2_ROLE"
    ok "Creato Instance Profile $EC2_PROFILE"
else
    ok "Instance Profile esistente: $EC2_PROFILE"
fi
state_set EC2_ROLE "$EC2_ROLE"
state_set EC2_PROFILE "$EC2_PROFILE"

# AWS Backup role
BACKUP_ROLE="tms-$ENV-aws-backup"

if ! aws iam get-role --role-name "$BACKUP_ROLE" >/dev/null 2>&1; then
    aws iam create-role \
        --role-name "$BACKUP_ROLE" \
        --assume-role-policy-document '{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"backup.amazonaws.com"},"Action":"sts:AssumeRole"}]}' \
        --tags "Key=Project,Value=LoginBusiness" "Key=Environment,Value=$ENV" \
        >/dev/null
    ok "Creato IAM role $BACKUP_ROLE"
else
    ok "IAM role esistente: $BACKUP_ROLE"
fi

aws iam attach-role-policy --role-name "$BACKUP_ROLE" \
    --policy-arn arn:aws:iam::aws:policy/service-role/AWSBackupServiceRolePolicyForBackup \
    >/dev/null 2>&1 || true
aws iam attach-role-policy --role-name "$BACKUP_ROLE" \
    --policy-arn arn:aws:iam::aws:policy/service-role/AWSBackupServiceRolePolicyForRestores \
    >/dev/null 2>&1 || true
state_set BACKUP_ROLE "$BACKUP_ROLE"
BACKUP_ROLE_ARN=$(aws iam get-role --role-name "$BACKUP_ROLE" --query 'Role.Arn' --output text)
state_set BACKUP_ROLE_ARN "$BACKUP_ROLE_ARN"

# ═════════════════════════════════════════════════════════════════════════
# 4. S3 — bucket invoices (Object Lock 10y) + bucket backups
# ═════════════════════════════════════════════════════════════════════════

log "── 4. S3 ─────────────────────────────────────────────────────────────"

INVOICES_BUCKET="tms-${ENV}-invoices-${ACCOUNT_ID}"
BACKUPS_BUCKET="tms-${ENV}-backups-${ACCOUNT_ID}"

create_s3_bucket() {
    local bucket="$1"
    local object_lock="${2:-false}"   # true/false

    if aws s3api head-bucket --bucket "$bucket" >/dev/null 2>&1; then
        ok "Bucket $bucket esistente"
        return 0
    fi

    local create_args=(--bucket "$bucket" --region "$AWS_REGION")
    if [ "$AWS_REGION" != "us-east-1" ]; then
        create_args+=(--create-bucket-configuration "LocationConstraint=$AWS_REGION")
    fi
    if [ "$object_lock" = "true" ]; then
        create_args+=(--object-lock-enabled-for-bucket)
    fi

    aws s3api create-bucket "${create_args[@]}" >/dev/null
    ok "Creato bucket $bucket"

    # Versioning (obbligatorio per Object Lock, utile sempre)
    aws s3api put-bucket-versioning \
        --bucket "$bucket" \
        --versioning-configuration Status=Enabled

    # Encryption AES256
    aws s3api put-bucket-encryption \
        --bucket "$bucket" \
        --server-side-encryption-configuration '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"}}]}'

    # Public access block totale
    aws s3api put-public-access-block \
        --bucket "$bucket" \
        --public-access-block-configuration "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true"

    # Tag
    aws s3api put-bucket-tagging \
        --bucket "$bucket" \
        --tagging "TagSet=[{Key=Project,Value=LoginBusiness},{Key=Environment,Value=$ENV}]"
}

# Invoices con Object Lock 10 anni
create_s3_bucket "$INVOICES_BUCKET" true

# Object Lock retention default 10 anni COMPLIANCE (coerente con obbligo
# fiscale italiano DPR 600 art. 22). Applicato solo al primo setup,
# modifiche successive richiedono specifica sostituzione.
aws s3api put-object-lock-configuration \
    --bucket "$INVOICES_BUCKET" \
    --object-lock-configuration '{"ObjectLockEnabled":"Enabled","Rule":{"DefaultRetention":{"Mode":"COMPLIANCE","Days":3650}}}' \
    2>/dev/null || warn "Object Lock già configurato (OK idempotente)"

# Lifecycle invoices: Standard → IA dopo 90 gg → Glacier Deep Archive dopo 1 anno
aws s3api put-bucket-lifecycle-configuration \
    --bucket "$INVOICES_BUCKET" \
    --lifecycle-configuration '{
      "Rules": [{
        "ID": "transition-to-cold-storage",
        "Status": "Enabled",
        "Filter": {},
        "Transitions": [
          {"Days": 90,  "StorageClass": "STANDARD_IA"},
          {"Days": 365, "StorageClass": "DEEP_ARCHIVE"}
        ]
      }]
    }'

# Backups bucket (no Object Lock — i dump non sono documenti fiscali)
create_s3_bucket "$BACKUPS_BUCKET" false

aws s3api put-bucket-lifecycle-configuration \
    --bucket "$BACKUPS_BUCKET" \
    --lifecycle-configuration '{
      "Rules": [{
        "ID": "expire-old-backups",
        "Status": "Enabled",
        "Filter": {},
        "Transitions": [
          {"Days": 30,  "StorageClass": "STANDARD_IA"},
          {"Days": 120, "StorageClass": "DEEP_ARCHIVE"}
        ],
        "Expiration": {"Days": 365},
        "NoncurrentVersionExpiration": {"NoncurrentDays": 30}
      }]
    }'

ok "Lifecycle configurate su invoices e backups"

state_set INVOICES_BUCKET "$INVOICES_BUCKET"
state_set BACKUPS_BUCKET "$BACKUPS_BUCKET"

# IAM policy inline per EC2 role: scrittura sui 2 bucket.
# Il JSON viene passato direttamente come stringa per evitare problemi di
# path portability (aws-cli Windows non legge i path MSYS /tmp/*).
S3_POLICY_DOC=$(cat <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["s3:PutObject", "s3:PutObjectAcl", "s3:GetObject", "s3:ListBucket"],
    "Resource": [
      "arn:aws:s3:::${BACKUPS_BUCKET}",
      "arn:aws:s3:::${BACKUPS_BUCKET}/*",
      "arn:aws:s3:::${INVOICES_BUCKET}",
      "arn:aws:s3:::${INVOICES_BUCKET}/*"
    ]
  }]
}
EOF
)
aws iam put-role-policy \
    --role-name "$EC2_ROLE" \
    --policy-name s3-access \
    --policy-document "$S3_POLICY_DOC"
ok "Inline policy s3-access attaccata a $EC2_ROLE"

# ═════════════════════════════════════════════════════════════════════════
# 5. SSH KEY PAIR admin
# ═════════════════════════════════════════════════════════════════════════

log "── 5. SSH Key Pair admin ─────────────────────────────────────────────"

KEYPAIR_NAME="tms-$ENV-admin"
if aws ec2 describe-key-pairs --key-names "$KEYPAIR_NAME" >/dev/null 2>&1; then
    ok "Key pair $KEYPAIR_NAME esistente"
else
    # `--public-key-material` è tipato come blob: se passato come stringa
    # aws-cli prova a fare base64-decode della stringa (fallisce sul testo
    # OpenSSH). La strada robusta è `fileb://` — con win_path per
    # convertire il path MSYS in formato Windows dove serve.
    aws ec2 import-key-pair \
        --key-name "$KEYPAIR_NAME" \
        --public-key-material "fileb://$(win_path "$ADMIN_SSH_PUBKEY_FILE")" \
        --tag-specifications "ResourceType=key-pair,Tags=[{Key=Name,Value=$KEYPAIR_NAME},{Key=Project,Value=LoginBusiness},{Key=Environment,Value=$ENV}]" \
        >/dev/null
    ok "Key pair $KEYPAIR_NAME importata"
fi
state_set KEYPAIR_NAME "$KEYPAIR_NAME"

# ═════════════════════════════════════════════════════════════════════════
# 6. EC2 instance + EBS data volume + Elastic IP
# ═════════════════════════════════════════════════════════════════════════

log "── 6. EC2 + EBS + Elastic IP ─────────────────────────────────────────"

# AMI Ubuntu 22.04 LTS ufficiale Canonical
AMI_ID=$(aws ec2 describe-images \
    --owners 099720109477 \
    --filters "Name=name,Values=ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*" \
              "Name=state,Values=available" \
              "Name=architecture,Values=x86_64" \
    --query 'Images | sort_by(@, &CreationDate) | [-1].ImageId' --output text)

[ "$AMI_ID" != "None" ] || die "AMI Ubuntu 22.04 non trovata in $AWS_REGION"
ok "Ubuntu 22.04 AMI: $AMI_ID"

# Render user_data.sh (iniettando variabili)
USER_DATA_TEMPLATE="$SCRIPT_DIR/user_data.sh"
[ -f "$USER_DATA_TEMPLATE" ] || die "user_data.sh mancante accanto a provision.sh"

RENDERED_USER_DATA=$(mktemp)
sed \
    -e "s|__ENVIRONMENT__|$ENV|g" \
    -e "s|__ADMIN_EMAIL__|$ADMIN_EMAIL|g" \
    -e "s|__APP_FQDN__|$APP_FQDN|g" \
    -e "s|__API_FQDN__|$API_FQDN|g" \
    -e "s|__AWS_REGION__|$AWS_REGION|g" \
    -e "s|__BACKUPS_BUCKET__|$BACKUPS_BUCKET|g" \
    -e "s|__INVOICES_BUCKET__|$INVOICES_BUCKET|g" \
    -e "s|__DEPLOY_SSH_PUBKEY__|$DEPLOY_SSH_PUBKEY|g" \
    "$USER_DATA_TEMPLATE" > "$RENDERED_USER_DATA"

# EC2 instance
INSTANCE_ID=$(aws ec2 describe-instances \
    --filters "Name=tag:Name,Values=tms-$ENV" \
              "Name=tag:Project,Values=LoginBusiness" \
              "Name=tag:Environment,Values=$ENV" \
              "Name=instance-state-name,Values=pending,running,stopped,stopping" \
    --query 'Reservations[0].Instances[0].InstanceId' --output text 2>/dev/null || echo "None")

if [ "$INSTANCE_ID" = "None" ] || [ -z "$INSTANCE_ID" ]; then
    BLOCK_DEVICE_MAPPING=$(cat <<EOF
[{"DeviceName":"/dev/sda1","Ebs":{"VolumeSize":$ROOT_VOLUME_GB,"VolumeType":"gp3","Encrypted":true,"DeleteOnTermination":true}}]
EOF
)
    INSTANCE_ID=$(aws ec2 run-instances \
        --image-id "$AMI_ID" \
        --instance-type "$INSTANCE_TYPE" \
        --subnet-id "$SUBNET_ID" \
        --security-group-ids "$SG_ID" \
        --key-name "$KEYPAIR_NAME" \
        --iam-instance-profile "Name=$EC2_PROFILE" \
        --block-device-mappings "$BLOCK_DEVICE_MAPPING" \
        --user-data "fileb://$(win_path "$RENDERED_USER_DATA")" \
        --associate-public-ip-address \
        --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=tms-$ENV},{Key=Project,Value=LoginBusiness},{Key=Environment,Value=$ENV}]" \
                             "ResourceType=volume,Tags=[{Key=Project,Value=LoginBusiness},{Key=Environment,Value=$ENV}]" \
        --query 'Instances[0].InstanceId' --output text)
    ok "Creata EC2 $INSTANCE_ID (attendo running...)"
    aws ec2 wait instance-running --instance-ids "$INSTANCE_ID"

    if [ "$ENV" = "prod" ]; then
        aws ec2 modify-instance-attribute --instance-id "$INSTANCE_ID" --disable-api-termination
        ok "API termination disabled su prod"
    fi
else
    ok "EC2 esistente: $INSTANCE_ID"
fi
rm -f "$RENDERED_USER_DATA"
state_set INSTANCE_ID "$INSTANCE_ID"

INSTANCE_AZ=$(aws ec2 describe-instances --instance-ids "$INSTANCE_ID" \
    --query 'Reservations[0].Instances[0].Placement.AvailabilityZone' --output text)

# EBS data volume
DATA_VOLUME_ID=$(aws ec2 describe-volumes \
    --filters "Name=tag:Name,Values=tms-$ENV-data" \
              "Name=tag:Project,Values=LoginBusiness" \
              "Name=tag:Environment,Values=$ENV" \
    --query 'Volumes[0].VolumeId' --output text 2>/dev/null || echo "None")

if [ "$DATA_VOLUME_ID" = "None" ] || [ -z "$DATA_VOLUME_ID" ]; then
    DATA_VOLUME_ID=$(aws ec2 create-volume \
        --availability-zone "$INSTANCE_AZ" \
        --size "$DATA_VOLUME_GB" \
        --volume-type gp3 \
        --encrypted \
        --tag-specifications "ResourceType=volume,Tags=[{Key=Name,Value=tms-$ENV-data},{Key=Project,Value=LoginBusiness},{Key=Environment,Value=$ENV},{Key=Backup,Value=true}]" \
        --query 'VolumeId' --output text)
    aws ec2 wait volume-available --volume-ids "$DATA_VOLUME_ID"
    ok "Creato EBS dati $DATA_VOLUME_ID"
else
    ok "EBS dati esistente: $DATA_VOLUME_ID"
fi
state_set DATA_VOLUME_ID "$DATA_VOLUME_ID"

# Attach (idempotente: se già attaccato nessun errore)
VOL_STATE=$(aws ec2 describe-volumes --volume-ids "$DATA_VOLUME_ID" \
    --query 'Volumes[0].State' --output text)
if [ "$VOL_STATE" = "available" ]; then
    aws ec2 attach-volume --volume-id "$DATA_VOLUME_ID" --instance-id "$INSTANCE_ID" --device /dev/sdf >/dev/null
    ok "EBS dati attaccato a $INSTANCE_ID come /dev/sdf"
fi

# Elastic IP
EIP_ALLOC=$(aws ec2 describe-addresses \
    --filters "Name=tag:Name,Values=tms-$ENV-eip" \
              "Name=tag:Project,Values=LoginBusiness" \
              "Name=tag:Environment,Values=$ENV" \
    --query 'Addresses[0].AllocationId' --output text 2>/dev/null || echo "None")

if [ "$EIP_ALLOC" = "None" ] || [ -z "$EIP_ALLOC" ]; then
    EIP_ALLOC=$(aws ec2 allocate-address \
        --domain vpc \
        --tag-specifications "ResourceType=elastic-ip,Tags=[{Key=Name,Value=tms-$ENV-eip},{Key=Project,Value=LoginBusiness},{Key=Environment,Value=$ENV}]" \
        --query 'AllocationId' --output text)
    ok "Allocato Elastic IP $EIP_ALLOC"
else
    ok "Elastic IP esistente: $EIP_ALLOC"
fi
state_set EIP_ALLOC "$EIP_ALLOC"

# Associa (idempotente: se è già associato alla nostra istanza non fa niente)
ASSOC_INSTANCE=$(aws ec2 describe-addresses --allocation-ids "$EIP_ALLOC" \
    --query 'Addresses[0].InstanceId' --output text)
if [ "$ASSOC_INSTANCE" != "$INSTANCE_ID" ]; then
    aws ec2 associate-address --allocation-id "$EIP_ALLOC" --instance-id "$INSTANCE_ID" >/dev/null
fi

EIP=$(aws ec2 describe-addresses --allocation-ids "$EIP_ALLOC" \
    --query 'Addresses[0].PublicIp' --output text)
state_set PUBLIC_IP "$EIP"
ok "Elastic IP $EIP associato a $INSTANCE_ID"

# ═════════════════════════════════════════════════════════════════════════
# 7. AWS Backup — vault + plan + selection
# ═════════════════════════════════════════════════════════════════════════

log "── 7. AWS Backup ─────────────────────────────────────────────────────"

VAULT_NAME="tms-$ENV-vault"
if ! aws backup describe-backup-vault --backup-vault-name "$VAULT_NAME" >/dev/null 2>&1; then
    aws backup create-backup-vault --backup-vault-name "$VAULT_NAME" >/dev/null
    ok "Creato vault $VAULT_NAME"
else
    ok "Vault esistente: $VAULT_NAME"
fi
state_set VAULT_NAME "$VAULT_NAME"

# Backup plan
PLAN_ID=$(aws backup list-backup-plans \
    --query "BackupPlansList[?BackupPlanName=='tms-$ENV-plan'].BackupPlanId | [0]" \
    --output text 2>/dev/null || echo "None")

if [ "$PLAN_ID" = "None" ] || [ -z "$PLAN_ID" ]; then
    PLAN_JSON=$(cat <<EOF
{
  "BackupPlanName": "tms-$ENV-plan",
  "Rules": [
    {
      "RuleName": "daily-7d",
      "TargetBackupVaultName": "$VAULT_NAME",
      "ScheduleExpression": "cron(0 3 * * ? *)",
      "StartWindowMinutes": 60,
      "CompletionWindowMinutes": 180,
      "Lifecycle": {"DeleteAfterDays": 7}
    },
    {
      "RuleName": "weekly-28d",
      "TargetBackupVaultName": "$VAULT_NAME",
      "ScheduleExpression": "cron(0 3 ? * SUN *)",
      "StartWindowMinutes": 60,
      "CompletionWindowMinutes": 360,
      "Lifecycle": {"DeleteAfterDays": 28}
    }
  ]
}
EOF
)
    PLAN_ID=$(aws backup create-backup-plan --backup-plan "$PLAN_JSON" --query 'BackupPlanId' --output text)
    ok "Creato backup plan $PLAN_ID"
else
    ok "Backup plan esistente: $PLAN_ID"
fi
state_set BACKUP_PLAN_ID "$PLAN_ID"

# Backup selection
SEL_EXISTS=$(aws backup list-backup-selections --backup-plan-id "$PLAN_ID" \
    --query "BackupSelectionsList[?SelectionName=='tms-$ENV-ebs-data'] | length(@)" --output text 2>/dev/null || echo 0)

if [ "$SEL_EXISTS" = "0" ]; then
    SEL_JSON=$(cat <<EOF
{
  "SelectionName": "tms-$ENV-ebs-data",
  "IamRoleArn": "$BACKUP_ROLE_ARN",
  "ListOfTags": [
    {"ConditionType": "STRINGEQUALS", "ConditionKey": "Backup", "ConditionValue": "true"}
  ]
}
EOF
)
    aws backup create-backup-selection \
        --backup-plan-id "$PLAN_ID" \
        --backup-selection "$SEL_JSON" >/dev/null
    ok "Creata backup selection per tag Backup=true"
else
    ok "Backup selection esistente"
fi

# ═════════════════════════════════════════════════════════════════════════
# 8. SSM Parameter — config CloudWatch Agent
# ═════════════════════════════════════════════════════════════════════════

log "── 8. SSM Parameter (CloudWatch Agent config) ────────────────────────"

CW_AGENT_PARAM="/tms/$ENV/cloudwatch-agent-config"
CW_AGENT_CONFIG=$(cat <<EOF
{
  "agent": {"metrics_collection_interval": 60, "run_as_user": "cwagent"},
  "metrics": {
    "namespace": "TMS/$ENV",
    "append_dimensions": {"InstanceId": "\${aws:InstanceId}"},
    "metrics_collected": {
      "cpu":  {"measurement": ["cpu_usage_idle","cpu_usage_iowait"], "metrics_collection_interval": 60, "totalcpu": true},
      "mem":  {"measurement": ["mem_used_percent"], "metrics_collection_interval": 60},
      "disk": {"measurement": ["used_percent"], "metrics_collection_interval": 60, "resources": ["/","/data"], "drop_device": true},
      "swap": {"measurement": ["swap_used_percent"], "metrics_collection_interval": 60}
    }
  },
  "logs": {
    "logs_collected": {
      "files": {
        "collect_list": [
          {"file_path": "/var/log/syslog",                "log_group_name": "/tms/$ENV/host", "log_stream_name": "{instance_id}/syslog",   "timezone": "UTC"},
          {"file_path": "/var/log/auth.log",              "log_group_name": "/tms/$ENV/host", "log_stream_name": "{instance_id}/auth",     "timezone": "UTC"},
          {"file_path": "/var/log/tms-user-data.log",     "log_group_name": "/tms/$ENV/host", "log_stream_name": "{instance_id}/user-data","timezone": "UTC"},
          {"file_path": "/var/log/tms-deploy.log",        "log_group_name": "/tms/$ENV/host", "log_stream_name": "{instance_id}/deploy",   "timezone": "UTC"},
          {"file_path": "/var/log/tms-backup-mongo.log",  "log_group_name": "/tms/$ENV/host", "log_stream_name": "{instance_id}/backup",   "timezone": "UTC"}
        ]
      }
    }
  }
}
EOF
)

# `aws ssm put-parameter --tags` fallisce se il parametro esiste già con
# --overwrite (API requirement). Proviamo prima con tags per il primo
# create, poi fallback senza tags per l'overwrite successivo.
aws ssm put-parameter \
    --name "$CW_AGENT_PARAM" \
    --type String \
    --value "$CW_AGENT_CONFIG" \
    --tier Standard \
    --tags "Key=Project,Value=LoginBusiness" "Key=Environment,Value=$ENV" \
    >/dev/null 2>&1 \
    || aws ssm put-parameter \
        --name "$CW_AGENT_PARAM" \
        --type String \
        --value "$CW_AGENT_CONFIG" \
        --overwrite \
        --tier Standard \
        >/dev/null
ok "SSM parameter $CW_AGENT_PARAM aggiornato"

# ═════════════════════════════════════════════════════════════════════════
# 9. CloudWatch Log Groups
# ═════════════════════════════════════════════════════════════════════════

log "── 9. CloudWatch Log Groups ─────────────────────────────────────────"

LOG_GROUPS=("/tms/$ENV/host" "/tms/$ENV/app")
RETENTION_DAYS=$([ "$ENV" = "prod" ] && echo 30 || echo 7)

for lg in "${LOG_GROUPS[@]}"; do
    if ! aws logs describe-log-groups --log-group-name-prefix "$lg" \
         --query "logGroups[?logGroupName=='$lg'] | length(@)" --output text 2>/dev/null | grep -q '^1$'; then
        aws logs create-log-group --log-group-name "$lg" \
            --tags "Project=LoginBusiness,Environment=$ENV" 2>/dev/null || true
        ok "Creato log group $lg"
    else
        ok "Log group esistente: $lg"
    fi
    aws logs put-retention-policy --log-group-name "$lg" --retention-in-days "$RETENTION_DAYS" 2>/dev/null || true
done

# ═════════════════════════════════════════════════════════════════════════
# 10. SNS topic + subscription email + 5 Alarm + Dashboard
# ═════════════════════════════════════════════════════════════════════════

log "── 10. SNS + CloudWatch Alarms + Dashboard ──────────────────────────"

TOPIC_ARN=$(aws sns create-topic --name "tms-$ENV-alerts" --query 'TopicArn' --output text)
ok "SNS topic: $TOPIC_ARN"
state_set SNS_TOPIC_ARN "$TOPIC_ARN"

# Subscription email (richiede conferma manuale via email AWS)
if [ -n "${ALERT_EMAIL:-}" ]; then
    EXISTING=$(aws sns list-subscriptions-by-topic --topic-arn "$TOPIC_ARN" \
        --query "Subscriptions[?Endpoint=='$ALERT_EMAIL'] | length(@)" --output text)
    if [ "$EXISTING" = "0" ]; then
        aws sns subscribe --topic-arn "$TOPIC_ARN" --protocol email --notification-endpoint "$ALERT_EMAIL" >/dev/null
        warn "SNS subscription creata per $ALERT_EMAIL — conferma via email prima che gli alarm arrivino"
    else
        ok "SNS subscription $ALERT_EMAIL già presente"
    fi
fi

# 5 alarm
create_alarm() {
    local name="$1" desc="$2" namespace="$3" metric="$4"
    local threshold="$5" periods="$6" period_sec="$7" stat="$8"
    shift 8
    # dimensioni passate come coppie Name=Value
    local dims=()
    for d in "$@"; do dims+=("Name=${d%=*},Value=${d#*=}"); done

    aws cloudwatch put-metric-alarm \
        --alarm-name "$name" \
        --alarm-description "$desc" \
        --comparison-operator GreaterThanThreshold \
        --evaluation-periods "$periods" \
        --metric-name "$metric" \
        --namespace "$namespace" \
        --period "$period_sec" \
        --statistic "$stat" \
        --threshold "$threshold" \
        --treat-missing-data notBreaching \
        --alarm-actions "$TOPIC_ARN" \
        --ok-actions "$TOPIC_ARN" \
        --dimensions "${dims[@]}"
}

# 1. Status check (diverso: usa GreaterThanOrEqual, breaching missing)
aws cloudwatch put-metric-alarm \
    --alarm-name "tms-$ENV-ec2-status-check" \
    --alarm-description "EC2 status check failed" \
    --comparison-operator GreaterThanOrEqualToThreshold \
    --evaluation-periods 2 \
    --metric-name StatusCheckFailed \
    --namespace AWS/EC2 \
    --period 60 \
    --statistic Maximum \
    --threshold 1 \
    --treat-missing-data breaching \
    --alarm-actions "$TOPIC_ARN" \
    --ok-actions "$TOPIC_ARN" \
    --dimensions "Name=InstanceId,Value=$INSTANCE_ID"

# 2. CPU high
create_alarm "tms-$ENV-cpu-high" "CPU > 80% per 10 min" "AWS/EC2" "CPUUtilization" 80 10 60 Average \
    "InstanceId=$INSTANCE_ID"

# 3. Disk root full
create_alarm "tms-$ENV-disk-root-full" "Disk / > 85%" "TMS/$ENV" "disk_used_percent" 85 2 300 Average \
    "InstanceId=$INSTANCE_ID" "path=/"

# 4. Disk /data full
create_alarm "tms-$ENV-disk-data-full" "Disk /data > 85%" "TMS/$ENV" "disk_used_percent" 85 2 300 Average \
    "InstanceId=$INSTANCE_ID" "path=/data"

# 5. Memory high
create_alarm "tms-$ENV-memory-high" "Memoria > 85% per 10 min" "TMS/$ENV" "mem_used_percent" 85 10 60 Average \
    "InstanceId=$INSTANCE_ID"

ok "5 alarm configurati"

# Dashboard
DASH_JSON=$(cat <<EOF
{
  "widgets": [
    {
      "type": "metric", "x": 0, "y": 0, "width": 12, "height": 6,
      "properties": {
        "metrics": [
          ["AWS/EC2", "CPUUtilization", "InstanceId", "$INSTANCE_ID"],
          ["TMS/$ENV", "mem_used_percent", "InstanceId", "$INSTANCE_ID"]
        ],
        "period": 60, "stat": "Average", "region": "$AWS_REGION",
        "title": "CPU + Memory"
      }
    },
    {
      "type": "metric", "x": 12, "y": 0, "width": 12, "height": 6,
      "properties": {
        "metrics": [
          ["TMS/$ENV", "disk_used_percent", "InstanceId", "$INSTANCE_ID", "path", "/"],
          [".", ".", ".", ".", ".", "/data"]
        ],
        "period": 300, "stat": "Average", "region": "$AWS_REGION",
        "title": "Disk usage"
      }
    },
    {
      "type": "metric", "x": 0, "y": 6, "width": 12, "height": 6,
      "properties": {
        "metrics": [
          ["AWS/EC2", "NetworkIn",  "InstanceId", "$INSTANCE_ID"],
          [".",       "NetworkOut", ".",          "."]
        ],
        "period": 60, "stat": "Sum", "region": "$AWS_REGION",
        "title": "Network in/out"
      }
    }
  ]
}
EOF
)
aws cloudwatch put-dashboard \
    --dashboard-name "tms-$ENV" \
    --dashboard-body "$DASH_JSON" >/dev/null
ok "Dashboard CloudWatch: tms-$ENV"

# ═════════════════════════════════════════════════════════════════════════
# RIEPILOGO FINALE
# ═════════════════════════════════════════════════════════════════════════

printf "\n${C_GREEN}%s${C_RESET}\n" "════════════════════════════════════════════════════════════════"
printf "${C_GREEN}%s${C_RESET}\n" "  PROVISIONING COMPLETATO (env=$ENV)"
printf "${C_GREEN}%s${C_RESET}\n" "════════════════════════════════════════════════════════════════"
cat <<EOF

  Public IP:        $EIP
  EC2 instance:     $INSTANCE_ID
  VPC:              $VPC_ID
  Invoices bucket:  $INVOICES_BUCKET
  Backups bucket:   $BACKUPS_BUCKET
  Dashboard:        https://$AWS_REGION.console.aws.amazon.com/cloudwatch/home?region=$AWS_REGION#dashboards:name=tms-$ENV

  State file:       $(state_file)

  Next steps:
  ───────────
  1. DNS su Google — aggiungi A record:
       $APP_FQDN   →  $EIP   TTL 300
       $API_FQDN   →  $EIP   TTL 300

  2. Verifica propagazione:
       dig +short $APP_FQDN

  3. SSH + primo certificato TLS:
       ssh -i \$ADMIN_SSH_KEY ubuntu@$EIP
       sudo certbot certonly --standalone \\
         -d $APP_FQDN -d $API_FQDN \\
         -m $ADMIN_EMAIL --agree-tos -n

  4. Configura GitHub Environment secrets per il deploy CI/CD:
       DEPLOY_HOST=$EIP
       DEPLOY_SSH_KEY=(contenuto di ~/.ssh/tms_deploy)
       API_FQDN=$API_FQDN
       REACT_APP_BACKEND_URL=https://$API_FQDN

  5. Prima della CI/CD, completa /opt/tms/.env sull'host con JWT_SECRET reale.

  6. Se ALERT_EMAIL è configurata: clicca il link di conferma ricevuto via email
     per attivare le notifiche SNS.

EOF
