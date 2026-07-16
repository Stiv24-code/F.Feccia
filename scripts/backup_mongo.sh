#!/bin/bash
# Dump MongoDB settimanale caricato su S3.
# Installato dal cloud-init (issue #16) in /opt/tms/scripts/backup_mongo.sh.
# Eseguito da cron sull'host EC2 (vedi /etc/cron.d/tms-backup-mongo).
#
# Usa l'IAM Instance Profile agganciato alla EC2 (nessuna access key statica):
# `aws cli` erediterà le credenziali dal metadata service.
#
# Esempio di entry cron (installata da cloud-init):
#   0 2 * * 0 root /opt/tms/scripts/backup_mongo.sh >> /var/log/tms-backup-mongo.log 2>&1
#
# Variabili d'ambiente attese (settate dal cron o da .env sourced in cima):
#   BACKUPS_BUCKET  - nome bucket S3 destinazione (es. tms-prod-backups-123...)
#   MONGO_CONTAINER - nome container mongo (default: tms-mongo-1)

set -euo pipefail

# Se lanciato da cron senza shell interattiva carica l'ambiente dal .env
# dell'app (che contiene almeno MONGO_URL, BACKUPS_BUCKET).
if [ -f /opt/tms/.env ]; then
    set -a
    # shellcheck disable=SC1091
    . /opt/tms/.env
    set +a
fi

BACKUPS_BUCKET="${BACKUPS_BUCKET:-}"
MONGO_CONTAINER="${MONGO_CONTAINER:-tms-mongo-1}"

if [ -z "$BACKUPS_BUCKET" ]; then
    echo "[$(date -Iseconds)] ERROR: BACKUPS_BUCKET not set — salvare su .env" >&2
    exit 1
fi

TIMESTAMP=$(date -u +%Y%m%d_%H%M%S)
DUMP_FILE="/tmp/mongo_${TIMESTAMP}.gz"
S3_KEY="mongo/weekly/${TIMESTAMP}.gz"

echo "[$(date -Iseconds)] Starting dump from container ${MONGO_CONTAINER}..."

# mongodump --archive --gzip scrive un unico file binario compresso.
# `docker exec` legge stdout del container.
docker exec "$MONGO_CONTAINER" \
    mongodump --archive --gzip --quiet \
    > "$DUMP_FILE"

BYTES=$(stat -c%s "$DUMP_FILE" 2>/dev/null || stat -f%z "$DUMP_FILE")
echo "[$(date -Iseconds)] Dump size: ${BYTES} bytes"

echo "[$(date -Iseconds)] Uploading to s3://${BACKUPS_BUCKET}/${S3_KEY}..."
aws s3 cp "$DUMP_FILE" "s3://${BACKUPS_BUCKET}/${S3_KEY}" \
    --storage-class STANDARD_IA \
    --sse AES256 \
    --only-show-errors

rm -f "$DUMP_FILE"

echo "[$(date -Iseconds)] Backup completato: s3://${BACKUPS_BUCKET}/${S3_KEY}"
