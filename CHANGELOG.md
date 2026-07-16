# Changelog

Tutte le modifiche rilevanti a **LoginBusiness TMS** sono documentate in questo file.

Il formato segue [Keep a Changelog](https://keepachangelog.com/) e il progetto adotta
[Semantic Versioning](https://semver.org/). Fino al go-live su AWS (Phase 5) il
progetto resta in `0.x`.

## [Unreleased]

In corso: **Phase 0 — Bonifica Emergent** (milestone 1) e pianificazione delle
successive fasi di sicurezza, infrastruttura AWS, refactor, feature gap e go-live.

### Changed
- `frontend/public/index.html` ricostruito da zero: rimossi script, badge "Made
  with Emergent" e snippet PostHog; aggiunti i font del design system
  (Space Grotesk, IBM Plex Sans, IBM Plex Mono). (#2)
- `frontend/craco.config.js` semplificato: rimossi i blocchi di caricamento
  condizionale per i plugin `visual-edits` e `health-check`; mantenuto solo
  l'alias `@ → src/`, le regole ESLint e le `watchOptions`. (#3)
- `backend/requirements.txt`: eliminati `emergentintegrations`, `openai`,
  `stripe`, `litellm`, `google-genai`, `google-generativeai` e le loro
  dipendenze transitive non più referenziate. (#4)
- `backend_test.py`: il `base_url` del client di test è ora configurabile
  tramite la variabile d'ambiente `BACKEND_URL` (default
  `http://localhost:8000/api`). (#5)
- `.gitignore`: aggiunte le voci `.emergent/` e `.gitconfig` per prevenire
  reintroduzione accidentale. (#5)

### Removed
- Cartelle `frontend/plugins/visual-edits/` e `frontend/plugins/health-check/`
  (runtime Emergent.sh per editor visuale e health endpoints). (#3)
- Cartella `.emergent/` con `emergent.yml` e `emergent_todos.json`. (#5)
- File `.gitconfig` top-level con identità `emergent-agent-e1`. (#5)

### Added
- File `README.md` con descrizione progetto, stack, quick start, struttura
  repo e riferimenti alle milestone. (#6)
- File `LICENSE` proprietario WhereTech S.r.l. (#6)
- Questo `CHANGELOG.md`. (#6)
- Moduli backend `config.py` (Pydantic Settings con fail-fast in produzione
  su CORS wildcard, cookie non-secure, JWT placeholder) e `dependencies.py`
  (`get_current_user` + `require_roles(*roles)` factory). (#7 #9 #11)
- Endpoint `POST /api/auth/refresh` per rotazione access token via cookie
  httpOnly. (#7)

### Security
- Trust `X-Forwarded-For` solo da reverse proxy fidati (CIDR whitelist
  via env var `TRUSTED_PROXIES`, default RFC1918 + localhost). Il backend
  dietro al container nginx non vede più l'IP interno di Docker ma il
  vero IP del browser, ripristinando il rate limit per-IP e l'accuratezza
  dei record audit. Se la request non arriva da un IP fidato, l'header
  XFF viene ignorato (anti-spoof). Nuovo modulo `backend/request_ip.py`;
  applicato anche `Annotated[list[str], NoDecode]` a `cors_origins` e
  `trusted_proxies` per evitare parse JSON errato delle CSV. (#59)
- Rimosso endpoint pubblico `POST /api/seed` e le credenziali admin
  hardcoded `admin@feccia.it / admin123` (CRITICAL). Seeding ora solo via
  script standalone `backend/seed_demo.py` con password autogenerata. (#10)
- Sostituzione hashing password da SHA256 custom a bcrypt con cost
  factor 12. Nuovo modulo `backend/security.py` condiviso tra `server.py`
  e `seed_demo.py`. (#8)
- Sostituito lo schema auth `token = user.id` in query string con JWT
  firmato (HS256, `algorithms=[HS256]` pinnato). Access token 15 min in
  memoria lato client, refresh token 7 giorni in cookie httpOnly con
  `path=/api/auth`. Rimosso uso di `localStorage` per il token. (#7)
- RBAC hardcoded su 79 endpoint (48 con `require_roles(...)` secondo
  matrice admin/amministrazione/planner/operatore, 31 con solo
  `get_current_user`). Voci menu sidebar filtrate per ruolo a livello UX.
  Il refactor a profili configurabili via UI Key User è tracciato in #54
  (Phase 4). (#9)
- CORS whitelist esplicita da `CORS_ORIGINS`, metodi ed header limitati,
  fail-fast su `*` + `credentials=True` in produzione. Security headers
  (X-Content-Type-Options, X-Frame-Options: DENY, Referrer-Policy,
  Permissions-Policy) su ogni response; HSTS in staging/production.
  Cookie refresh con `httpOnly + Secure (dinamico) + SameSite=Lax`. (#11)
- `UserCreate` richiede password di almeno 12 caratteri e role tra i
  quattro valori noti (`Literal`). (#7 #9 #11)
- Audit log automatico per ogni mutazione su `/api/*` più logging
  esplicito su `/auth/login|refresh|logout` (con discriminazione degli
  outcome: credenziali invalide, utente disattivato, cookie mancante,
  ecc.). Nuova collection `audit_logs` con indice TTL a 10 anni e
  indici per-user e per-resource. Nessuna password/token/hash mai
  persistito nei record. (#12)
- Escape `$regex` su tutti gli input utente negli endpoint di ricerca
  (13 call site in customers/destinations/vehicles/drivers/carriers/
  products/orders) tramite nuovo helper `_safe_regex` che fa
  `re.escape()` + truncation a 100 caratteri. Neutralizza ReDoS e
  regex injection. (#12)
- Rate limit su endpoint auth e export via `slowapi`. Limiti: login
  5 tentativi / 15 min per IP, refresh 20 / 15 min, register 3 / ora,
  export ordini 10 / ora. 429 con header `Retry-After`; frontend
  mostra toast dedicato. Storage memory in dev, Redis in
  staging/production (forzato da `assert_production_safe`). (#13)

### Changed
- **Infrastructure-as-Code sostituita: da Terraform a script aws-cli.**
  Nuovo `infra/aws/provision.sh` idempotente (identificazione risorse via
  tag `Project=LoginBusiness + Environment + Name`). `teardown.sh` con
  conferma interattiva. `user_data.sh` template con placeholder sostituiti
  via sed. Terraform archiviato in `infra/terraform-archive/` (non
  eliminato, riferimento storico). CI aggiornata: job `terraform` → `infra`
  con `bash -n` + shellcheck. (#70)
- Basemap della pagina Mappa Viaggi: da CARTO light a 4 basemap Esri (AGOL)
  con selettore — default National Geographic (stilizzata e colorata),
  alternative World Topographic / World Street Map / World Imagery. Nessuna
  API key richiesta (servizio pubblico `server.arcgisonline.com`). (#69)

### Changed (Phase 3 — Refactor & Tech Debt)

- **Decomposizione `backend/server.py` (2266 righe) in moduli per dominio** (#20).
  Nuova struttura:
  - `app.py` — composition root (FastAPI app + CORS + middleware + include dei router)
  - `database.py` — Motor client + helpers (`now_utc`, `new_id`, `safe_regex`, `get_next_sequence`)
  - `models.py` — tutti i Pydantic model (UserBase, CustomerBase, OrderBase, ecc.)
  - `services.py` — logica di dominio condivisa (auth cookies, pricelist scoring, OSRM + route cache, GPS simulation, map helpers, coords mapping)
  - `routers/{auth,dashboard,customers,destinations,vehicles,drivers,carriers,products,garages,masterdata,pricelists,orders,trips,invoices,driver_unavailability,availability,map,export,meta}.py` — un `APIRouter` per dominio, ognuno sotto le 200 righe
  
  `Dockerfile` CMD passa da `server:app` ad `app:app`. CI `py_compile` esteso a tutti i nuovi moduli. Path degli endpoint invariati (ancora tutti sotto `/api`), quindi nessuna modifica lato frontend.

### Added (Phase 3 — Refactor & Tech Debt)

- Logging strutturato JSON via **structlog** + middleware **correlation ID** (#22).
  Nuovo modulo `backend/logging_setup.py` che:
  - configura stdlib logging + structlog con renderer JSON in staging/production
    e `ConsoleRenderer` colorato in development;
  - espone `get_logger(name)` usato dai moduli (`auth`, `services`, `app`, ecc.);
  - inietta automaticamente `timestamp`, `level`, `request_id`, `user_id`,
    `path`, `method` su ogni record tramite `contextvars`.
  
  Nuovo middleware `correlation_id_middleware` in `app.py` (registrato come
  outermost, quindi audit + security headers + CORS ereditano i contextvars):
  legge l'header `X-Request-ID` o genera un UUID4, lo binda in
  `structlog.contextvars` e lo riflette nella response.
  
  `dependencies.get_current_user` binda `user_id` + `user_role` dopo la
  decodifica del JWT, così ogni log nel resto della request include chi
  ha invocato l'endpoint. Aggiunta dipendenza `structlog==24.4.0` in
  `requirements.txt`.
- Sistema di migration MongoDB idempotente (#21). Nuovo package
  `backend/migrations/` con discovery automatica per prefisso numerico
  e `backend/scripts/migrate.py` che registra ogni file applicato nella
  collection `_migrations`. Prima migrazione `001_create_indexes.py`
  crea gli indici su `users`, `customers`, `destinations`, `vehicles`,
  `drivers`, `carriers`, `products`, `garages`, `orders`, `pricelists`,
  `trips`, `invoices`, `gps_history`, `driver_unavailability`,
  `route_cache`, `counters` e `audit_logs` (quest'ultimo via
  `ensure_audit_indexes`). Il wrapper `tms-deploy` esegue
  `docker compose run --rm backend python scripts/migrate.py` prima di
  `up -d`, così la pipeline CD non pubblica mai codice che legge prima
  di avere gli indici. Rimosso l'`@app.on_event("startup")` che creava
  l'indice di `audit_logs` a runtime.

### Infrastruttura (Phase 2 — solo codice, no apply ancora)

- Docker stack per dev locale + deploy EC2: Dockerfile multi-stage backend
  (Python 3.12 slim + gunicorn + tini, non-root) e frontend
  (build Node 20 → runtime nginx:1.27-alpine). `docker-compose.yml` con
  nginx + backend + mongo:7 + redis:7 sul network `tms`. Override
  `docker-compose.prod.yml` con mount Let's Encrypt + resource limits
  per t3.small + logging rotato. `frontend/nginx/default.conf` unico per
  SPA + reverse proxy `/api` + TLS ready. (#14)
- Terraform base in `infra/terraform/` per provisioning eu-west-1 Dublino:
  VPC + subnet pubblica + Security Group + EC2 t3.small Ubuntu 22.04 +
  EBS data 20 GB encrypted + Elastic IP. State remoto su S3 con lock
  DynamoDB. DNS su Google (no Route53). Output `dns_instructions` con
  record A da configurare a mano. (#15)
- Cloud-init completo (`user_data.sh.tftpl`) che al primo avvio configura
  l'host: Docker CE + compose plugin, utente `deploy` per CI/CD con
  restriction `command="/usr/local/bin/tms-deploy"`, UFW firewall,
  fail2ban, certbot via snap con renewal hook, mount EBS dati
  idempotente, unattended-upgrades. Wrapper di deploy accetta tar
  stream da SSH, estrae in `/opt/tms`, aggiorna `IMAGE_TAG` e fa
  `docker compose pull && up -d`. (#16, #18)
- AWS Backup: vault + plan con rule daily 7gg + weekly 28gg sul volume
  EBS taggato `Backup=true`. S3 bucket `tms-{env}-invoices-{acc}` con
  **Object Lock COMPLIANCE 10 anni** (obbligo fiscale italiano). S3
  `tms-{env}-backups-{acc}` per dump MongoDB weekly. IAM Instance
  Profile con `CloudWatchAgentServerPolicy` + policy inline S3 sui
  due bucket, nessuna access key statica. `scripts/backup_mongo.sh`
  installato dalla CI/CD in `/opt/tms/scripts/` ed eseguito da cron
  ogni domenica 02:00 UTC. (#17)
- GitHub Actions CI (lint + build + validate su PR e push) e CD (build
  Docker su push main/develop → push GHCR con tag sha e env-latest →
  SSH deploy via wrapper restricted → smoke test `GET /api/health`).
  Concurrency group per cancellare run obsolete. Secrets richiesti per
  GitHub Environment: `DEPLOY_HOST`, `DEPLOY_SSH_KEY`, `API_FQDN`,
  `REACT_APP_BACKEND_URL`. (#18)
- CloudWatch: SSM Parameter con config agent (CPU, mem, disk, swap,
  log di syslog/auth/tms-\*), log groups `/tms/{env}/host` e `/app`,
  SNS topic + subscription email opzionale, 5 alarm (EC2 status check,
  CPU sostenuta > 80%, disk root > 85%, disk `/data` > 85%, memoria >
  85%) con OK actions, dashboard CloudWatch con 3 widget (CPU+mem,
  disk, network). Agent installato + avviato dal cloud-init. (#19)

---

## [0.0.0] — 2026-04-23

Fork iniziale del progetto dal POC generato con Emergent.sh.
Stato: backend FastAPI + MongoDB e frontend React + shadcn/ui funzionanti
(V1.3 del POC, mappa viaggi con percorsi Bézier e GPS simulato).
