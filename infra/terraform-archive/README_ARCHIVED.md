# ⚠️ ARCHIVIO — Terraform non più usato

Questa cartella contiene la versione Terraform dell'infrastruttura, sviluppata
nelle PR #63, #64, #65, #66, #67 e successivamente sostituita dalla versione
basata su AWS CLI in `../aws/` (preferenza dell'utente, PR feat/aws-cli-provisioning).

**NON eseguire `terraform apply` da qui** — crea risorse parallele a quelle
provisionate da `infra/aws/provision.sh`.

Se in futuro si decidesse di tornare a Terraform:
1. Verificare che nessuna risorsa sia stata creata da `infra/aws/provision.sh`
   (o fare teardown prima)
2. Rinominare la cartella in `infra/terraform/`
3. Rimuovere `infra/aws/`
4. Aggiornare CI e docs
