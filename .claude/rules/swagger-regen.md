# swagger-regen — rigenerazione swagger + client API frontend

Quando si modifica qualcosa che impatta il contratto REST del backend Go — route
(`backend/internal/app/routes.go`), annotazioni swag su un handler (`// @Summary`, `// @Param`, `// @Success`, ecc.),
DTO di request/response, o un campo di un modello GORM esposto via API — eseguire sempre, in quest'ordine:

1. **Backend**: `cd backend && make swagger` (= `swag init`) → rigenera `backend/docs/swagger.json` e `swagger.yaml`.
2. **Frontend**: `cd frontend && yarn generate:api` → rigenera `frontend/src/api/**` da `backend/docs/swagger.json`
   (tool: `swagger-typescript-api`, vedi script `generate:api` in `frontend/package.json`).

Non considerare completa una modifica a modelli/DTO/route del backend finché questi due step non sono
stati eseguiti e i file generati (`backend/docs/swagger.*`, `frontend/src/api/**`) non sono coerenti con
il codice sorgente. Se lo step 2 produce errori TypeScript altrove nel frontend, è un segnale che il
contratto è cambiato in modo breaking: segnalarlo, non ignorarlo.

Non modificare a mano i file generati (`backend/docs/swagger.json`, `backend/docs/swagger.yaml`,
`backend/docs/docs.go`, `frontend/src/api/**`) — sono output di `swag init` / `swagger-typescript-api`.
