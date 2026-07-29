# LoginBusiness — Transport Management System

Transport Management System (TMS) sviluppato da **res data** per il cliente **FECCIA F.lli**.

Gestisce il ciclo operativo end-to-end di un'azienda di autotrasporti:
**Raccolta Ordini → Pianificazione → Viaggi → Chiusura → Fatturazione**, con moduli per anagrafiche, listini tariffari, mappa interattiva e tracking GPS.

Riferimento funzionale: *Analisi Tecnica RC-FEC-25001 del 4/8/2025* (documento interno, non pubblicato nel repo).

---

## Stack tecnologico

### Backend
- **Go** — modulo `fratelli-feccia`, `backend/`
- **Fiber** (HTTP framework) + **GORM** (ORM)
- **PostgreSQL 16**
- JWT (access + refresh token)

### Frontend
- **React 19** + **shadcn/ui** (Radix UI)
- **Vite**
- **Tailwind CSS 3**
- **react-router-dom 7**
- **Leaflet** + **react-leaflet** per la mappa viaggi
- **recharts** per i grafici KPI
- **react-hook-form** + **zod** per i form

### Infrastruttura target
In fase di definizione (dominio e piattaforma di deploy da confermare). Non fare affidamento
su indicazioni AWS/GitHub Actions/dominio presenti in versioni precedenti di questo file o in
`infra/` — vanno riverificate prima dell'uso.

---

## Quick start (sviluppo locale)

### Prerequisiti
- Docker Desktop 24+
- (opzionale per hot reload) Go 1.25+, Node 20, Yarn 1.22

### Avvio dello stack completo con Docker Compose

```bash
# Crea un file .env nella root con le variabili richieste da docker-compose.yml
# (GO_DB_PASSWORD, GO_JWT_ACCESS_SECRET, GO_JWT_REFRESH_SECRET, SEED_ADMIN_PASSWORD, ...)
docker compose up -d --build
```

Servizi esposti:
- **Frontend (SPA + reverse proxy)**: http://localhost
- **API health**: http://localhost/api/v1/health
- **PostgreSQL**: interno al network `tms`, non esposto sull'host

Con `IS_LOCAL=true` (default nel compose di sviluppo), il backend semina automaticamente un
dataset demo completo al primo avvio su un DB vuoto (vedi `backend/internal/seeddemo`, mai in
produzione). In alternativa, da dentro `backend/`: `make seed-demo`.

### Sviluppo senza Docker (hot reload)

#### Backend
```bash
cd backend
go run ./main.go
# oppure, con hot reload: air (usa backend/.air.toml)
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
SBG.Feccia.TMS/
├── backend/                   # API Go (Fiber + GORM + PostgreSQL)
│   ├── main.go
│   ├── config/
│   ├── internal/              # handlers, models, services, seeddemo, ...
│   ├── pkg/                   # swagger, pdfgen, ...
│   ├── docs/                  # swagger.json / swagger.yaml (generati, non a mano)
│   └── Makefile                # run, build, test, swagger, seed-demo, ...
├── frontend/                  # SPA React + Vite
│   ├── src/
│   │   ├── pages/              # DashboardPage, OrdersPage, PlannerPage, ...
│   │   ├── components/         # ui/ (shadcn) + shared/ + layout/
│   │   └── api/                # client generato da swagger (yarn generate:api)
│   └── package.json
├── design_guidelines.md       # design system (palette, font, spacing, status colors)
├── plan.md                    # storico release V1 → V1.3 (POC)
└── README.md
```

---

## Licenza

Software proprietario — © 2025-2026 **res data** — Tutti i diritti riservati.
Vedi il file [`LICENSE`](./LICENSE) per i dettagli.

## Contatti

- Cliente: FECCIA F.lli
