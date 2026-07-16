# LoginBusiness — Transport Management System

Transport Management System (TMS) sviluppato da **WhereTech S.r.l.** per il cliente **FECCIA F.lli**.

Gestisce il ciclo operativo end-to-end di un'azienda di autotrasporti:
**Raccolta Ordini → Pianificazione → Viaggi → Chiusura → Fatturazione**, con moduli per anagrafiche, listini tariffari, mappa interattiva e tracking GPS.

Riferimento funzionale: *Analisi Tecnica RC-FEC-25001 del 4/8/2025* (documento interno, non pubblicato nel repo).

---

## Stack tecnologico

### Backend
- **Python 3.12** + **FastAPI 0.110** (async)
- **MongoDB 7** via driver **Motor**
- **Redis 7** (cache, rate limiting)
- **Pydantic v2** per modelli e settings
- **PyJWT** + **bcrypt** per autenticazione (in corso di migrazione — vedi milestone *Phase 1 — Sicurezza critica*)

### Frontend
- **React 19** + **shadcn/ui** (Radix UI)
- **Tailwind CSS 3**
- **react-router-dom 7**
- **Leaflet** + **react-leaflet** per la mappa viaggi
- **recharts** per i grafici KPI
- **react-hook-form** + **zod** per i form

### Infrastruttura target
- **AWS eu-west-1 (Dublino)**, MVP minimal: 1 EC2 `t3.small` con Docker Compose
- **Nginx** come reverse proxy + TLS termination (certificati Let's Encrypt via `certbot` sull'host)
- Provisioning via **aws-cli script** (`infra/aws/provision.sh`) idempotente (VPC, EC2, EBS, Elastic IP, Security Group, S3, AWS Backup, CloudWatch). Versione Terraform archiviata in `infra/terraform-archive/`.
- DNS esterno su Google DNS (`wheretech.it`)
- CI/CD via **GitHub Actions** con deploy SSH

Dettagli completi: vedi la master issue [#1](../../issues/1) e le milestone *Phase 0 → Phase 5*.

---

## Quick start (sviluppo locale)

### Prerequisiti
- Docker Desktop 24+
- (opzionale per hot reload) Python 3.12, Node 20, Yarn 1.22

### Avvio dello stack completo con Docker Compose

```bash
cp .env.example .env
# Modifica .env — in particolare JWT_SECRET (genera con: openssl rand -hex 32)
docker compose up -d
```

Servizi esposti:
- **Frontend (SPA + reverse proxy)**: http://localhost
- **API health**: http://localhost/api/health
- **MongoDB**: interno al network `tms`, non esposto sull'host
- **Redis**: interno al network `tms`, non esposto sull'host

Per popolare un DB fresco con dati demo:
```bash
docker compose exec backend python seed_demo.py
# Stampa le credenziali admin generate (salvarle nel password manager)
```

### Deploy produzione (EC2)

Sulla EC2 provisionata (issue #15 + #16):
```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```
Le immagini vengono pullate da GHCR, i certificati Let's Encrypt montati dall'host, il volume MongoDB puntato sul disco EBS dedicato `/data/mongo`.

### Sviluppo senza Docker (hot reload)

#### Backend
```bash
cd backend
python -m venv .venv
.venv\Scripts\activate          # Windows
# source .venv/bin/activate     # macOS/Linux
pip install -r requirements.txt
cp ../.env.example .env
uvicorn app:app --reload --port 8000
```

#### Frontend
```bash
cd frontend
yarn install
yarn start
```

---

## Struttura del repository

```
LoginBusiness/
├── backend/                   # API FastAPI + MongoDB
│   ├── app.py                 # composition root (FastAPI + middleware + routers)
│   ├── database.py            # Motor client + DB helpers
│   ├── models.py              # Pydantic models
│   ├── services.py            # pricelist match / OSRM / GPS sim
│   ├── routers/               # un file per dominio (auth, orders, trips, ...)
│   ├── migrations/            # migrazioni Mongo idempotenti (#21)
│   ├── scripts/migrate.py     # runner delle migrazioni
│   ├── requirements.txt
│   └── seed_demo.py           # script standalone per dati demo
├── frontend/                  # SPA React
│   ├── public/index.html
│   ├── src/
│   │   ├── App.js
│   │   ├── pages/             # DashboardPage, OrdersPage, PlannerPage, ...
│   │   ├── components/        # ui/ (shadcn) + shared/ + layout/
│   │   └── lib/               # api.js, auth-context.js, ...
│   └── package.json
├── tests/                     # (placeholder) pytest suite
├── design_guidelines.md       # design system (palette, font, spacing, status colors)
├── plan.md                    # storico release V1 → V1.3 (POC)
├── backend_test.py            # smoke test API via requests (BACKEND_URL env)
└── README.md
```

---

## Deploy

Il deploy su AWS è tracciato nella milestone *Phase 2 — Infrastruttura AWS minimal* (issue [#14](../../issues/14) → [#19](../../issues/19)).

In breve, una volta completata la Phase 2:
- push su branch `develop` → deploy automatico su ambiente staging
- push su branch `main` → deploy produzione (con approval gate)

Entrambi via workflow GitHub Actions → SSH sulla EC2 target → `docker compose pull && up -d`.

---

## Licenza

Software proprietario — © 2025-2026 **WhereTech S.r.l.** — Tutti i diritti riservati.
Vedi il file [`LICENSE`](./LICENSE) per i dettagli.

## Contatti

- Referente tecnico: Gianluca Caporossi — gcaporossi@wheretech.it
- Cliente: FECCIA F.lli
