# infra/aws — Provisioning AWS con aws-cli

Questa cartella contiene gli script bash che provisionano l'infra AWS per il
TMS LoginBusiness. Sostituisce la versione Terraform (archiviata in
`../terraform-archive/`) per preferenza dell'utente.

## Prerequisiti (una tantum)

```bash
# 1. AWS CLI configurato con un IAM user admin (non root account).
aws configure
aws sts get-caller-identity         # verifica che risponda

# 2. Chiave SSH amministrativa.
ssh-keygen -t ed25519 -f ~/.ssh/tms_admin -C "tms admin"

# 3. (Opzionale finché la CI/CD non serve) chiave SSH per il deploy user.
ssh-keygen -t ed25519 -f ~/.ssh/tms_deploy -C "tms deploy"

# 4. Copia il file di configurazione e compilalo.
cp infra/aws/config/prod.env.example infra/aws/config/prod.env
# Apri con l'editor e imposta:
#   ADMIN_CIDR=<il_tuo_ip>/32      (trovalo con: curl -s https://ifconfig.me)
#   ADMIN_SSH_PUBKEY_FILE=$HOME/.ssh/tms_admin.pub
#   DEPLOY_SSH_PUBKEY_FILE=$HOME/.ssh/tms_deploy.pub   (o lascia vuoto ora)
#   APP_FQDN / API_FQDN / ADMIN_EMAIL / ALERT_EMAIL
```

## Provisioning

```bash
infra/aws/provision.sh prod
```

Lo script è **idempotente**: rieseguirlo dopo il primo apply non crea
duplicati. Identifica le risorse esistenti via tag
`Project=LoginBusiness + Environment=<env> + Name=<resource>`. Gli ID delle
risorse create vengono scritti in `infra/aws/.aws-state/<env>.env`
(gitignored) per riferimento rapido e per il teardown.

Durata tipica prima apply: 5-8 minuti (dominato dal wait su
`instance-running` e dall'allocazione dei bucket S3).

### Cosa viene creato

| Risorsa | Nome |
|---|---|
| VPC 10.0.0.0/16 | `tms-<env>` |
| Subnet pubblica 10.0.1.0/24 | `tms-<env>-public-a` |
| Internet Gateway | `tms-<env>-igw` |
| Route table + default route | `tms-<env>-rt-public` |
| Security Group | `tms-<env>-ec2` (SSH admin + 80/443) |
| EC2 Ubuntu 22.04 | `tms-<env>` |
| EBS root gp3 encrypted | (incluso nell'EC2) |
| EBS data gp3 encrypted 20GB | `tms-<env>-data` (tag `Backup=true`) |
| Elastic IP | `tms-<env>-eip` |
| SSH key pair admin | `tms-<env>-admin` |
| IAM role + instance profile | `tms-<env>-ec2` |
| IAM role AWS Backup | `tms-<env>-aws-backup` |
| S3 bucket fatture | `tms-<env>-invoices-<account_id>` (Object Lock 10y) |
| S3 bucket backup | `tms-<env>-backups-<account_id>` |
| AWS Backup vault | `tms-<env>-vault` |
| AWS Backup plan | `tms-<env>-plan` (daily 7gg + weekly 28gg) |
| SSM parameter CloudWatch config | `/tms/<env>/cloudwatch-agent-config` |
| CloudWatch Log groups | `/tms/<env>/host`, `/tms/<env>/app` |
| SNS topic + subscription email | `tms-<env>-alerts` |
| 5 CloudWatch Alarms | status check, CPU, disk /, disk /data, memoria |
| CloudWatch Dashboard | `tms-<env>` |

Totale: ~35 risorse.

## Post-provisioning (passi manuali)

Lo script stampa un riepilogo finale. I passi successivi sono:

### 1. DNS su Google
Copia l'`Public IP` dall'output di provision e aggiungi i record A sul
pannello Google DNS della zona `wheretech.it`:

```
A  tms.wheretech.it       <EIP>   TTL 300
A  api.tms.wheretech.it   <EIP>   TTL 300
```

Attendi la propagazione (5-15 min):

```bash
dig +short tms.wheretech.it      # deve mostrare l'EIP
```

### 2. Primo certificato Let's Encrypt
Dopo che il DNS risolve, SSH sull'host e richiedi il certificato
(richiede porta 80 libera → nginx container NON deve girare ancora):

```bash
ssh -i ~/.ssh/tms_admin ubuntu@<EIP>
sudo certbot certonly --standalone \
  -d tms.wheretech.it -d api.tms.wheretech.it \
  -m gcaporossi@wheretech.it --agree-tos -n
```

### 3. Conferma SNS email
Se hai configurato `ALERT_EMAIL`, AWS ti manda un'email con link di
conferma. Cliccalo prima che gli alarm possano notificarti.

### 4. GitHub Environment secrets (per CI/CD)
Nelle Settings del repo → Environments → `prod` (crea se non c'è),
aggiungi:

- `DEPLOY_HOST=<EIP>`
- `DEPLOY_SSH_KEY=<contenuto di ~/.ssh/tms_deploy>`
- `API_FQDN=api.tms.wheretech.it`
- `REACT_APP_BACKEND_URL=https://api.tms.wheretech.it`

Configura una protection rule che richieda un reviewer per environment
`prod`.

### 5. JWT_SECRET reale
Sull'host, completa `/opt/tms/.env` sostituendo il placeholder
`JWT_SECRET=PLACEHOLDER-...` con un valore reale:

```bash
ssh -i ~/.ssh/tms_admin ubuntu@<EIP>
sudo nano /opt/tms/.env
# genera: openssl rand -hex 32
# sudo chown deploy:deploy /opt/tms/.env
```

### 6. Primo deploy
Push su `develop` → staging, o su `main` → prod. La CI/CD
(`.github/workflows/deploy.yml`) builda le immagini Docker su GHCR e le
deploya via SSH wrapper.

## Teardown

```bash
infra/aws/teardown.sh prod
```

Richiede conferma interattiva (scrivi `YES`). Elimina tutto in ordine
inverso. **Eccezione**: il bucket `tms-<env>-invoices-<account_id>` NON
viene cancellato perché ha **Object Lock COMPLIANCE 10 anni**: gli
oggetti (fatture) non possono essere rimossi fino alla scadenza.

Se hai già fatto snapshot/backup points, il vault non può essere
cancellato finché non li rimuovi a mano:

```bash
aws backup list-recovery-points-by-backup-vault --backup-vault-name tms-prod-vault
aws backup delete-recovery-point --backup-vault-name tms-prod-vault \
  --recovery-point-arn <arn>
```

## File

- `provision.sh` — entry point idempotente (~500 righe, 10 sezioni numerate)
- `teardown.sh` — rimozione in ordine inverso (richiede `YES`)
- `user_data.sh` — template cloud-init con placeholder `__XXX__` sostituiti
  da provision.sh
- `lib/common.sh` — helper logging + state file + tag filter
- `config/prod.env.example` — config di esempio (copia in `prod.env`)
- `config/staging.env.example` — (da creare per lo staging, clone di prod)
- `.aws-state/<env>.env` — gitignored; generato da provision.sh con gli ID
  delle risorse create

## Costo stimato (eu-west-1 Dublino, prod 24/7)

| Voce | €/mese |
|---|---:|
| EC2 t3.small | ~14 |
| EBS gp3 50 GB (root + dati) | ~4 |
| Elastic IP (attached) | 0 |
| S3 invoices + backups | ~1 |
| AWS Backup snapshot | ~2 |
| CloudWatch (metriche + log) | ~1-2 |
| SNS + SSM | <1 |
| Data transfer out | ~3 |
| **Totale** | **~€25/mese** |

## Debug

```bash
# Stato locale delle risorse create
cat infra/aws/.aws-state/prod.env

# Vedi cosa esiste in AWS con i nostri tag
aws ec2 describe-instances \
  --filters "Name=tag:Project,Values=LoginBusiness" \
            "Name=tag:Environment,Values=prod"

# Log cloud-init (primo avvio EC2)
ssh -i ~/.ssh/tms_admin ubuntu@<EIP>
sudo cat /var/log/tms-user-data.log        # tutto l'output
sudo cat /var/log/tms-user-data.done       # marker di completamento

# Log deploy CI/CD (successivi)
sudo cat /var/log/tms-deploy.log

# Log dump MongoDB
sudo cat /var/log/tms-backup-mongo.log

# Stato container
ssh -i ~/.ssh/tms_admin ubuntu@<EIP>
cd /opt/tms && docker compose ps
```
