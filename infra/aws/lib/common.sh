#!/bin/bash
# infra/aws/lib/common.sh — helper condivisi per provision/teardown.
#
# `source` questo file dagli script che usano aws-cli. Esporta:
#   log, die, require, tag_name_filter, resource_id_by_tag,
#   wait_aws, json, check_aws_cli.
#
# Le funzioni non presumono variabili globali fuori da quelle documentate
# (ENV, PROJECT_TAG, AWS_REGION): così lo stesso helper può essere usato
# sia per prod che staging.

# shellcheck shell=bash
set -o pipefail

# ── Colori output ─────────────────────────────────────────────────────────
if [ -t 1 ]; then
    C_BLUE='\033[1;34m'
    C_GREEN='\033[1;32m'
    C_YELLOW='\033[1;33m'
    C_RED='\033[1;31m'
    C_DIM='\033[2m'
    C_RESET='\033[0m'
else
    C_BLUE='' C_GREEN='' C_YELLOW='' C_RED='' C_DIM='' C_RESET=''
fi

# ── Logging ──────────────────────────────────────────────────────────────
log()  { printf "${C_BLUE}[%s]${C_RESET} %s\n" "$(date +%H:%M:%S)" "$*"; }
ok()   { printf "${C_GREEN}[✓]${C_RESET} %s\n" "$*"; }
warn() { printf "${C_YELLOW}[!]${C_RESET} %s\n" "$*" >&2; }
dim()  { printf "${C_DIM}%s${C_RESET}\n" "$*"; }

die() {
    printf "${C_RED}[✗] %s${C_RESET}\n" "$*" >&2
    exit 1
}

# ── Precondizioni ────────────────────────────────────────────────────────
require() {
    # require <cmd>           — fallisce se il comando non è nel PATH
    # require_var <VAR_NAME>  — fallisce se la variabile è vuota
    local cmd="$1"
    command -v "$cmd" >/dev/null 2>&1 || die "missing command: $cmd"
}

require_var() {
    local name="$1"
    local value="${!name-}"
    [ -n "$value" ] || die "missing env var: $name"
}

check_aws_cli() {
    require aws
    aws sts get-caller-identity >/dev/null 2>&1 \
        || die "AWS CLI non configurata. Esegui prima: aws configure"
    ok "aws-cli OK, account: $(aws sts get-caller-identity --query Account --output text)"
}

# ── Tag-based identification ──────────────────────────────────────────────
# Ogni risorsa è taggata con Project=LoginBusiness + Environment=<env>.
# Le funzioni seguenti ritornano l'ID (o "None") della prima risorsa che
# matcha <Name> tag all'interno dell'environment corrente.

tag_filters() {
    # Genera i filtri standard per `aws ec2 describe-*` / `aws ec2 wait ...`
    local name="$1"
    printf "Name=tag:Project,Values=%s Name=tag:Environment,Values=%s Name=tag:Name,Values=%s" \
        "LoginBusiness" "$ENV" "$name"
}

# ── Path portability (MSYS/Git Bash su Windows) ──────────────────────────
# aws-cli su Windows è un binary nativo e non comprende i path MSYS
# (es. `/tmp/foo`, `/c/Users/...`). Usiamo `cygpath -w` quando disponibile
# per convertirli in path Windows classici (`C:\...`). Su Linux/Mac la
# funzione è un no-op e ritorna il path invariato.
win_path() {
    if command -v cygpath >/dev/null 2>&1; then
        cygpath -w "$1"
    else
        printf '%s' "$1"
    fi
}

# ── Retry helper per eventuali throttling o eventual consistency ─────────
retry() {
    # retry <max_attempts> <cmd...>
    local max="$1"; shift
    local attempt=1 rc=0
    while [ "$attempt" -le "$max" ]; do
        if "$@"; then
            return 0
        fi
        rc=$?
        warn "tentativo $attempt/$max fallito (rc=$rc), riprovo tra 3s..."
        sleep 3
        attempt=$((attempt + 1))
    done
    return "$rc"
}

# ── Wait idempotente su state (uso describe loop) ────────────────────────
wait_until() {
    # wait_until <max_seconds> <check_cmd...>
    # Il check_cmd deve uscire 0 quando la condizione è soddisfatta.
    local max="$1"; shift
    local waited=0
    while [ "$waited" -lt "$max" ]; do
        if "$@" >/dev/null 2>&1; then
            return 0
        fi
        sleep 3
        waited=$((waited + 3))
    done
    return 1
}

# ── JSON helpers (evita dipendenza da jq) ─────────────────────────────────
# aws-cli ha --query nativo, preferiamolo. jq è opzionale.
# Scriviamo le funzioni che lavorano con --query + --output text.

# ── State file locale (gitignored) ───────────────────────────────────────
# Salviamo gli ID delle risorse create in `.aws-state/<env>.env` per
# riferimento rapido e per il teardown. Il file è un .env Bash-sourceable.
STATE_DIR="$REPO_ROOT/infra/aws/.aws-state"
state_file() {
    mkdir -p "$STATE_DIR"
    echo "$STATE_DIR/${ENV}.env"
}

state_set() {
    # state_set KEY value   — persiste una coppia KEY=VALUE, deduplicata per key
    local key="$1" val="$2"
    local file
    file="$(state_file)"
    touch "$file"
    # Rimuovi la vecchia entry per key (se c'è), aggiungi la nuova.
    grep -v "^${key}=" "$file" > "${file}.tmp" 2>/dev/null || true
    mv "${file}.tmp" "$file"
    echo "${key}=${val}" >> "$file"
}

state_get() {
    local key="$1"
    local file
    file="$(state_file)"
    [ -f "$file" ] || return 1
    grep "^${key}=" "$file" 2>/dev/null | tail -1 | cut -d= -f2-
}

state_clear() {
    local file
    file="$(state_file)"
    [ -f "$file" ] && rm -f "$file"
}
