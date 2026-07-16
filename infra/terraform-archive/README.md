# Terraform — Infrastruttura AWS LoginBusiness

Provisioning dell'ambiente AWS minimal per `tms.wheretech.it`. Una sola EC2
`t3.small` in `eu-west-1` (Dublino) con volume dati EBS separato, Elastic IP
fisso, Security Group hardened.

> DNS: **NON gestito da Terraform**. Il dominio `wheretech.it` sta su Google
> DNS. Dopo `terraform apply` l'output `dns_instructions` elenca i record A
> da configurare a mano.

## Prerequisiti (una tantum)

1. **AWS CLI** configurato con un IAM user (non root) che abbia almeno le
   policy `AmazonEC2FullAccess`, `AmazonVPCFullAccess`, `AmazonS3FullAccess`
   (per lo state), `AWSBackupFullAccess` (per issue #17), `CloudWatchFullAccess`
   (per #19). `aws sts get-caller-identity` deve rispondere.
2. **Terraform** ≥ 1.7 (`terraform version`).
3. **SSH key** admin: `ssh-keygen -t ed25519 -f ~/.ssh/tms_admin -C "tms admin"`.
4. **Bucket + tabella lock per lo state** (una volta per account AWS):
   ```bash
   ACC=$(aws sts get-caller-identity --query Account --output text)
   aws s3api create-bucket \
     --bucket tms-tfstate-$ACC \
     --region eu-west-1 \
     --create-bucket-configuration LocationConstraint=eu-west-1
   aws s3api put-bucket-versioning \
     --bucket tms-tfstate-$ACC \
     --versioning-configuration Status=Enabled
   aws s3api put-bucket-encryption \
     --bucket tms-tfstate-$ACC \
     --server-side-encryption-configuration '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"}}]}'
   aws s3api put-public-access-block \
     --bucket tms-tfstate-$ACC \
     --public-access-block-configuration BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true
   aws dynamodb create-table \
     --table-name tms-tfstate-lock \
     --attribute-definitions AttributeName=LockID,AttributeType=S \
     --key-schema AttributeName=LockID,KeyType=HASH \
     --billing-mode PAY_PER_REQUEST \
     --region eu-west-1
   ```

## Preparare un ambiente

```bash
cd infra/terraform
cp environments/prod.tfvars.example environments/prod.tfvars
cp environments/prod.backend.hcl.example environments/prod.backend.hcl
# Editare entrambi con i valori reali (IP admin, bucket name, ...)

terraform init -backend-config=environments/prod.backend.hcl
terraform plan  -var-file=environments/prod.tfvars -out=plan.out
# Rivedere che il plan mostri ~15 risorse da creare, nessuna da distruggere.
terraform apply plan.out
```

Output atteso:

```
public_ip = "52.XX.XX.XX"
ssh_command = "ssh -i ~/.ssh/tms_admin ubuntu@52.XX.XX.XX"
dns_instructions = <<EOT
  Configura manualmente su Google DNS...
EOT
```

## Configurazione DNS post-apply

Copia `public_ip` dall'output. Sul pannello Google DNS (zona `wheretech.it`)
aggiungi i record A come indicato in `dns_instructions`. Attendi la
propagazione (5-15 min) — verifica con `dig +short tms.wheretech.it`.

## Certificati TLS

I certificati Let's Encrypt si generano **dopo** che il DNS risolve
correttamente (vedi `dns_instructions` nell'output). Il cloud-init installa
certbot ma non lancia la prima richiesta (serve DNS operativo).

## Destroy

```bash
terraform destroy -var-file=environments/prod.tfvars
```

Attenzione: `disable_api_termination = true` è settato per `prod`. Per
distruggere davvero prod serve prima:

```bash
aws ec2 modify-instance-attribute --instance-id <id> --no-disable-api-termination
```

poi `terraform destroy`.

## File

- `versions.tf`    — provider AWS + backend S3 + default tags.
- `variables.tf`   — input (region, instance_type, admin_cidr, ssh_public_key, FQDN).
- `main.tf`        — VPC + subnet + IGW + SG + EC2 + EBS + EIP + key pair.
- `outputs.tf`     — public_ip, ssh_command, dns_instructions.
- `user_data.sh.tftpl` — bootstrap host (placeholder minimale; completato da #16).
- `environments/`  — file `.tfvars` e `.backend.hcl` per prod/staging (gitignored).

## Costo stimato (eu-west-1, prod, run 24/7)

| Risorsa | $/mese | € approx |
|---|---:|---:|
| EC2 t3.small on-demand | ~$15 | ~€14 |
| EBS gp3 50 GB (30 root + 20 data) | ~$4 | ~€4 |
| Elastic IP (attached) | $0 | €0 |
| VPC + IGW + SG | $0 | €0 |
| **Totale base (solo #15)** | **~$19** | **~€18** |

Aggiungi ~€2 S3+backup con #17, ~€3 data transfer out, ~€1 CloudWatch con
#19. Target ~€24-25/mese prod.
