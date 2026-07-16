#!/bin/bash
# infra/aws/user_data.sh — cloud-init del primo boot EC2.
#
# Questo file è un template: `provision.sh` sostituisce i placeholder
# `__XXX__` con i valori reali prima di passarlo come user-data a `aws ec2
# run-instances`.
#
# Placeholder iniettati:
#   __ENVIRONMENT__        prod|staging
#   __ADMIN_EMAIL__        email per certbot
#   __APP_FQDN__           es. tms.wheretech.it
#   __API_FQDN__           es. api.tms.wheretech.it
#   __AWS_REGION__         eu-west-1
#   __BACKUPS_BUCKET__     nome bucket S3 per dump MongoDB
#   __INVOICES_BUCKET__    nome bucket S3 fatture
#   __DEPLOY_SSH_PUBKEY__  chiave pubblica CI/CD (vuota al primo apply OK)
#
# Log su /var/log/tms-user-data.log; marker /var/log/tms-user-data.done al
# completamento con successo.

set -euxo pipefail
exec > >(tee -a /var/log/tms-user-data.log) 2>&1

echo "[$(date -Iseconds)] cloud-init START env=__ENVIRONMENT__ app=__APP_FQDN__ api=__API_FQDN__"

export DEBIAN_FRONTEND=noninteractive

# ── 1. Base + unattended-upgrades ───────────────────────────────────────
apt-get update
apt-get upgrade -y
apt-get install -y \
    ca-certificates \
    curl \
    gnupg \
    lsb-release \
    ufw \
    fail2ban \
    unattended-upgrades \
    apt-listchanges \
    xfsprogs \
    awscli \
    cron \
    rsync

cat > /etc/apt/apt.conf.d/50unattended-upgrades-custom <<'EOF'
Unattended-Upgrade::Allowed-Origins {
    "${distro_id}:${distro_codename}-security";
    "${distro_id}ESMApps:${distro_codename}-apps-security";
    "${distro_id}ESM:${distro_codename}-infra-security";
};
Unattended-Upgrade::Automatic-Reboot "false";
Unattended-Upgrade::Remove-Unused-Kernel-Packages "true";
EOF
cat > /etc/apt/apt.conf.d/20auto-upgrades <<'EOF'
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";
APT::Periodic::AutocleanInterval "7";
EOF

# ── 2. Mount EBS data ─────────────────────────────────────────────────────
DATA_DEV=""
for candidate in /dev/nvme1n1 /dev/sdf /dev/xvdf; do
    if [ -b "$candidate" ]; then
        DATA_DEV="$candidate"
        break
    fi
done

if [ -z "$DATA_DEV" ]; then
    echo "[$(date -Iseconds)] ATTENZIONE: nessun device EBS dati trovato" >&2
else
    if ! blkid "$DATA_DEV" >/dev/null 2>&1; then
        mkfs.ext4 -L tms-data "$DATA_DEV"
    fi
    mkdir -p /data
    DATA_UUID=$(blkid -s UUID -o value "$DATA_DEV")
    if ! grep -q "$DATA_UUID" /etc/fstab; then
        echo "UUID=$DATA_UUID /data ext4 defaults,nofail,noatime 0 2" >> /etc/fstab
    fi
    mount -a
    mkdir -p /data/mongo /data/backups
    chown -R 999:999 /data/mongo
fi

# ── 3. Docker Engine + compose plugin ──────────────────────────────────
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
    | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
chmod a+r /etc/apt/keyrings/docker.gpg

ARCH=$(dpkg --print-architecture)
CODENAME=$(. /etc/os-release && echo "$VERSION_CODENAME")
echo "deb [arch=$ARCH signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $CODENAME stable" \
    > /etc/apt/sources.list.d/docker.list

apt-get update
apt-get install -y \
    docker-ce \
    docker-ce-cli \
    containerd.io \
    docker-buildx-plugin \
    docker-compose-plugin

systemctl enable --now docker

# ── 4. Utente deploy + wrapper tms-deploy ────────────────────────────────
if ! id -u deploy >/dev/null 2>&1; then
    useradd -m -s /bin/bash deploy
fi
usermod -aG docker deploy

mkdir -p /home/deploy/.ssh
chmod 700 /home/deploy/.ssh

mkdir -p /opt/tms
chown -R deploy:deploy /opt/tms

cat > /usr/local/bin/tms-deploy <<'DEPLOY_EOF'
#!/bin/bash
set -euo pipefail
APP_DIR=/opt/tms
LOG=/var/log/tms-deploy.log
TS=$(date -Iseconds)
exec >>"$LOG" 2>&1
echo "[$TS] deploy START"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
if [ ! -t 0 ]; then
    tar xzf - -C "$TMP" 2>/dev/null || true
fi
if [ -n "$(ls -A "$TMP" 2>/dev/null || true)" ]; then
    echo "[$TS] syncing payload to $APP_DIR"
    # -a: archive mode; --exclude='.env': non sovrascrivere il .env host-side
    # (contiene JWT_SECRET reale generato al provisioning).
    # NO --delete-excluded: cancellerebbe proprio il .env che vogliamo preservare.
    # Usa --delete con --filter per rimuovere solo file non-esclusi obsoleti.
    rsync -a --delete --exclude='.env' "$TMP"/ "$APP_DIR"/
fi
if [ -f "$APP_DIR/tms-deploy.env" ]; then
    # shellcheck disable=SC1091
    source "$APP_DIR/tms-deploy.env"
    if [ -n "${IMAGE_TAG:-}" ] && [ -f "$APP_DIR/.env" ]; then
        sed -i -E "s|^IMAGE_TAG=.*$|IMAGE_TAG=$IMAGE_TAG|" "$APP_DIR/.env"
        echo "[$TS] IMAGE_TAG updated to $IMAGE_TAG"
    fi
    rm -f "$APP_DIR/tms-deploy.env"
fi
cd "$APP_DIR"
timeout 600 docker compose -f docker-compose.yml -f docker-compose.prod.yml pull
# Run pending MongoDB migrations inside a throwaway backend container.
# depends_on con condition=service_healthy avvia mongo se serve e lo
# lascia up per il successivo `up -d`. `--rm` scarta il container a fine corsa.
timeout 180 docker compose -f docker-compose.yml -f docker-compose.prod.yml run --rm backend python scripts/migrate.py
timeout 300 docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
docker image prune -f
echo "[$TS] deploy DONE"
DEPLOY_EOF
chmod 755 /usr/local/bin/tms-deploy

cat > /etc/logrotate.d/tms-deploy <<'EOF'
/var/log/tms-deploy.log {
    weekly
    rotate 4
    compress
    delaycompress
    missingok
    notifempty
    # logrotate gira come root, ma ricrea il file con ownership deploy così il
    # wrapper /usr/local/bin/tms-deploy (eseguito come deploy via SSH restricted)
    # può scriverci. Senza, ogni rotazione rompe il deploy successivo.
    su root root
    create 644 deploy deploy
}
EOF
# Inizializza il log con la giusta ownership (logrotate altrimenti crea root:root
# al primo run prima della prima rotazione settimanale).
touch /var/log/tms-deploy.log
chown deploy:deploy /var/log/tms-deploy.log
chmod 644 /var/log/tms-deploy.log

DEPLOY_KEY="__DEPLOY_SSH_PUBKEY__"
if [ -n "$DEPLOY_KEY" ]; then
    cat > /home/deploy/.ssh/authorized_keys <<EOF
command="/usr/local/bin/tms-deploy",no-port-forwarding,no-agent-forwarding,no-X11-forwarding,restrict $DEPLOY_KEY
EOF
    chmod 600 /home/deploy/.ssh/authorized_keys
fi
chown -R deploy:deploy /home/deploy/.ssh

# ── 5. UFW firewall ───────────────────────────────────────────────────────
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

# ── 6. fail2ban SSH jail ─────────────────────────────────────────────────
cat > /etc/fail2ban/jail.d/ssh.local <<'EOF'
[sshd]
enabled = true
port = ssh
maxretry = 5
findtime = 600
bantime = 3600
EOF
systemctl enable --now fail2ban

# ── 7. certbot via snap ───────────────────────────────────────────────────
snap install core
snap refresh core
snap install --classic certbot
ln -sf /snap/bin/certbot /usr/local/bin/certbot

mkdir -p /etc/letsencrypt/renewal-hooks/deploy
cat > /etc/letsencrypt/renewal-hooks/deploy/reload-nginx.sh <<'EOF'
#!/bin/bash
if command -v docker >/dev/null 2>&1; then
    cd /opt/tms 2>/dev/null && docker compose exec -T nginx nginx -s reload 2>/dev/null || true
fi
EOF
chmod +x /etc/letsencrypt/renewal-hooks/deploy/reload-nginx.sh

# ── 8. .env iniziale con placeholder JWT_SECRET ─────────────────────────
# Il backend (Pydantic Settings) accetta solo `development|staging|production`.
# L'infra usa `prod` come tag breve: traduci al volo prima di scrivere .env.
APP_ENVIRONMENT="__ENVIRONMENT__"
if [ "$APP_ENVIRONMENT" = "prod" ]; then APP_ENVIRONMENT="production"; fi
if [ ! -f /opt/tms/.env ]; then
    cat > /opt/tms/.env <<EOF
ENVIRONMENT=$APP_ENVIRONMENT
DB_NAME=loginbusiness
MONGO_URL=mongodb://mongo:27017
REDIS_URL=redis://redis:6379/0
RATE_LIMIT_STORAGE_URI=redis://redis:6379/0
JWT_SECRET=PLACEHOLDER-CHANGE-ME-BEFORE-STARTING-GENERATE-WITH-openssl-rand-hex-32
JWT_ACCESS_TTL_MINUTES=15
JWT_REFRESH_TTL_DAYS=7
CORS_ORIGINS=https://__APP_FQDN__,https://__API_FQDN__
SECURE_COOKIES=true
TRUSTED_PROXIES=127.0.0.1,::1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16
ENABLE_GPS_SIMULATION=false
LOG_LEVEL=INFO
IMAGE_TAG=prod-latest
BACKUPS_BUCKET=__BACKUPS_BUCKET__
S3_INVOICES_BUCKET=__INVOICES_BUCKET__
S3_INVOICES_RETENTION_YEARS=10
S3_PRESIGNED_TTL_SECONDS=900
AWS_REGION=__AWS_REGION__
EOF
    chmod 640 /opt/tms/.env
    chown deploy:deploy /opt/tms/.env
fi

# ── 9. Cron dump MongoDB weekly ─────────────────────────────────────────
cat > /etc/cron.d/tms-backup-mongo <<'EOF'
# Domenica 02:00 UTC
0 2 * * 0 root /opt/tms/scripts/backup_mongo.sh >> /var/log/tms-backup-mongo.log 2>&1
EOF
chmod 644 /etc/cron.d/tms-backup-mongo

cat > /etc/logrotate.d/tms-backup-mongo <<'EOF'
/var/log/tms-backup-mongo.log {
    weekly
    rotate 8
    compress
    delaycompress
    missingok
    notifempty
    create 644 root root
}
EOF

# ── 10. CloudWatch Agent ─────────────────────────────────────────────────
CW_AGENT_DEB=/tmp/amazon-cloudwatch-agent.deb
curl -fsSL -o "$CW_AGENT_DEB" \
    "https://s3.__AWS_REGION__.amazonaws.com/amazoncloudwatch-agent-__AWS_REGION__/ubuntu/amd64/latest/amazon-cloudwatch-agent.deb"
dpkg -i -E "$CW_AGENT_DEB"
rm -f "$CW_AGENT_DEB"

/opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl \
    -a fetch-config \
    -m ec2 \
    -c "ssm:/tms/__ENVIRONMENT__/cloudwatch-agent-config" \
    -s || echo "[$(date -Iseconds)] CW Agent config fetch failed (SSM param creato da provision.sh?)"

# ── 11. motd con istruzioni per il primo SSH admin ──────────────────────
cat > /etc/motd <<'EOF'
╔═══════════════════════════════════════════════════════════════╗
║   LoginBusiness TMS — host EC2                                ║
║                                                               ║
║   Primo setup una tantum (dopo aver configurato DNS):         ║
║     sudo certbot certonly --standalone \                      ║
║       -d <app_fqdn> -d <api_fqdn> \                           ║
║       -m <admin_email> --agree-tos -n                         ║
║                                                               ║
║   Deploy (normalmente via CI/CD):                             ║
║     cd /opt/tms                                               ║
║     # .env va completato con JWT_SECRET reale                 ║
║     docker compose -f docker-compose.yml \                    ║
║                    -f docker-compose.prod.yml pull            ║
║     docker compose -f docker-compose.yml \                    ║
║                    -f docker-compose.prod.yml up -d           ║
║                                                               ║
║   Log user_data: sudo cat /var/log/tms-user-data.log          ║
║   Status:        ls -la /var/log/tms-user-data.done           ║
╚═══════════════════════════════════════════════════════════════╝
EOF

touch /var/log/tms-user-data.done
echo "[$(date -Iseconds)] cloud-init DONE successfully"
