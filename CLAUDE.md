# CLAUDE.md — LoginBusiness TMS

> Non committato (`.git/info/exclude`). Istruzioni private per questo repo.

## Panoramica
**LoginBusiness** — TMS sviluppato da WhereTech per FECCIA F.lli. Rif. RC-FEC-25001.
Deriva da POC Emergent.sh, ora in porting verso produzione AWS minimale.
Prodotto pre-consegna: nessun utente reale, nessun vincolo di backward compatibility.

Contesto essenziale: `plan.md` (release POC), `design_guidelines.md` (UI),
`c:\tmp\analisi_tecnica.txt` (spec funzionale estratta dal PDF).

## Stack
- Backend: Go (modulo `fratelli-feccia`), `backend/` — Fiber + GORM + **PostgreSQL**. Porting da un
  precedente backend Python (FastAPI + Motor/MongoDB) completato: nginx instrada solo su questo backend,
  il codice Python è stato rimosso dal repo.
- Frontend: React 19 + shadcn/ui + Tailwind + Leaflet. CRA+craco (→ Vite in #25).
- DB: PostgreSQL (servizio `postgres` in `docker-compose.yml`). Nessun Mongo/Redis (rimossi col cutover —
  Redis serviva solo al rate-limiter Python multi-istanza, il limiter Go è in-memory single-instance).
- Seed dati demo: `backend/internal/seeddemo` (`make seed-demo` o automatico all'avvio su DB vuoto se
  `IS_LOCAL=true`, mai in produzione).

## Infrastruttura target
- AWS `eu-west-1`. **Una sola EC2 t3.small** + Elastic IP. Docker Compose host.
- Container: `nginx` (proxy + SPA + TLS) + `backend-go` + `postgres`.
- **Nginx**, non Caddy. TLS via `certbot` host. DNS esterno su **Google**, non Route53.
- Provisioning via **aws-cli script** in `infra/aws/provision.sh` (idempotente, tag-based).
  La versione Terraform sta in `infra/terraform-archive/` solo come riferimento storico.
- CI/CD via GitHub Actions + SSH deploy + GHCR.
- Stima: ~€25/mese prod.

## Working directory
`c:\Users\gcapo\OneDrive - WhereTech\AI\ClaudeLavoro\LoginBusiness\repo\`

## Strumenti
- **Repo**: https://github.com/gcaporossi-wheretech/LoginBusiness
- **Master tracking**: [issue #1](https://github.com/gcaporossi-wheretech/LoginBusiness/issues/1) — aggiornare solo a chiusura fase
- **Milestones**: Phase 0..5 (piano di porting)
- **Memory**: `C:\Users\gcapo\.claude\projects\c--Users-gcapo-OneDrive---WhereTech-AI-ClaudeLavoro-LoginBusiness\memory\`
  (`project_phase_status.md` = snapshot stato fasi/issue)

## Workflow
Il flusso issue→PR→merge è automatizzato dalla skill `close-issue-pr`.
Le convenzioni di codice, naming e commit sono nelle rules di `.claude/rules/`.

## Azioni destructive
AWS / DNS / segreti / force push → conferma obbligatoria. Vedi `.claude/rules/aws-destructive-actions.md`.

## Cosa NON fare (riassunto)
- Niente riferimenti Emergent → rule `no-emergent-references.md`
- Niente segreti in chiaro → rule `no-hardcoded-secrets.md`
- Niente firma `Co-Authored-By: Claude` sui commit
- Non reintrodurre codice/dipendenze Python nel repo (backend rimosso, vedi Stack)
- Non toccare `frontend/src/components/ui/*.jsx` (primitivi shadcn)
- PR > ~1000 righe solo dopo aver discusso lo scope
- Non cambiare decisioni architetturali già prese (`eu-west-1`, Nginx, DNS Google) senza discuterne
