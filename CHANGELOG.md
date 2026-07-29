# Changelog

Tutte le modifiche rilevanti a **LoginBusiness TMS** sono documentate in questo file.

Il formato segue [Keep a Changelog](https://keepachangelog.com/) e il progetto adotta
[Semantic Versioning](https://semver.org/). Fino al go-live in produzione il progetto
resta in `0.x`.

## [Unreleased]

### Changed
- `frontend/public/index.html` ricostruito da zero: rimossi script, badge "Made
  with Emergent" e snippet PostHog; aggiunti i font del design system
  (Space Grotesk, IBM Plex Sans, IBM Plex Mono). (#2)
- `frontend/craco.config.js` semplificato: rimossi i blocchi di caricamento
  condizionale per i plugin `visual-edits` e `health-check`; mantenuto solo
  l'alias `@ → src/`, le regole ESLint e le `watchOptions`. (#3)
- `.gitignore`: aggiunte le voci `.emergent/` e `.gitconfig` per prevenire
  reintroduzione accidentale. (#5)
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

### Removed
- Cartelle `frontend/plugins/visual-edits/` e `frontend/plugins/health-check/`
  (runtime Emergent.sh per editor visuale e health endpoints). (#3)
- Cartella `.emergent/` con `emergent.yml` e `emergent_todos.json`. (#5)
- File `.gitconfig` top-level con identità `emergent-agent-e1`. (#5)

### Added
- File `README.md` con descrizione progetto, stack, quick start, struttura
  repo e riferimenti alle milestone. (#6)
- File `LICENSE` proprietario. (#6)
- Questo `CHANGELOG.md`. (#6)

---

## [0.0.0] — storico pre-porting Go (Python/FastAPI/MongoDB)

Fork iniziale del progetto dal POC generato con Emergent.sh (2026-04-23),
backend Python (FastAPI + MongoDB via Motor), frontend React + shadcn/ui
(V1.3 del POC, mappa viaggi con percorsi Bézier e GPS simulato).

In questa fase, sul backend Python: hardening di sicurezza (RBAC su 79
endpoint, JWT HS256 con refresh cookie httpOnly, CORS whitelist esplicita,
security headers, rate limiting via slowapi, audit log con TTL 10 anni,
escape regex anti-ReDoS, trust `X-Forwarded-For` solo da proxy fidati),
decomposizione di `server.py` (2266 righe) in moduli per dominio, sistema
di migrazioni MongoDB idempotenti, logging strutturato JSON con
correlation ID. Infrastruttura: Docker Compose (nginx + backend + mongo +
redis), Terraform per EC2 + VPC + EBS + Elastic IP su AWS eu-west-1,
cloud-init con hardening host, AWS Backup su volumi/S3 con Object Lock,
CI/CD GitHub Actions verso GHCR, CloudWatch monitoring. Issue coinvolte:
#7–#22, #59.

Stack sostituito integralmente da Go + PostgreSQL nel porting successivo
(vedi `CLAUDE.md`) — i dettagli implementativi Python di questa fase non
sono più rilevanti per il codice attuale.
