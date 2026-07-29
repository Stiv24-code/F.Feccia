# LoginBusiness (FECCIA F.lli) — Development Plan (POC → V1 → V1.1 → V1.2 → **V1.3**) — **Updated (V1.3 Completed)**

## 1) Obiettivi
- Consegnare un **Transport Management System (TMS)** completo in italiano, con ciclo operativo end-to-end: **Ordine → Pianificazione → Viaggio → Chiusura → Fatturazione**.
- Offrire una UI data-heavy e operator-friendly coerente con il design system:
  - **sidebar scura (#0B1220)** + workspace chiaro
  - accento teal (HSL **195 92% 28%**)
  - **colori stato**: giallo (PIANIFICABILE), rosso (VIAGGIO), arancione (CHIUSO), verde (FATTURATO)
- Backend modulare con CRUD stabile, esportazioni e basi per integrazioni.
- Estendere l’operatività con visualizzazioni “planner-first” e strumenti di controllo per l’operatore.

**Estensioni implementate:**
- **V1.1 (COMPLETATA)**
  1) **Listini: Rule Editor** + **lookup tariffa** e precompilazione tariffa in creazione ordine
  2) **Disponibilità mezzi/autisti** + **gestione indisponibilità autisti** (ferie/malattia/permesso)
- **V1.2 (COMPLETATA)**
  - **Feature A:** Listini — **editing regole** + `item_id` stabile (no delete per indice)
  - **Feature B:** Planner V2 — **griglie duali** per Partenze/Rientri (mockup pag. 16)
- **V1.3 (COMPLETATA)**
  - **Feature C:** **Mappa Viaggi** interattiva (percorsi, mezzi, filtri, dettaglio) per visualizzazione intuitiva e innovativa.

**Stato attuale:**
- ✅ V1 consegnata e stabile
- ✅ V1.1 consegnata e integrata end-to-end
- ✅ V1.2 consegnata e verificata
- ✅ V1.3 consegnata e verificata (mappa con dati reali dimostrativi)

---

## 2) Step di implementazione

### Phase 1 — Core Flow POC (isolation, no auth)
**Goal:** Validare flusso e persistenza su DB.

User stories:
1. Creare un ordine con cliente, carico/scarico, date, tariffa.
2. Passare da **PIANIFICABILE** a **VIAGGIO** assegnando mezzo/autista/vettore.
3. Chiudere ordine/viaggio (**CHIUSO**).
4. Creare proforma e finalizzare (**FATTURATO**).

**Steps (COMPLETATA):**
- Modelli ed entità dati:
  - `customers`, `destinations`, `vehicles`, `drivers`, `carriers`, `products`, `garages`
  - `pricelists`, `orders`, `trips`, `invoices`, `counters`
- Endpoint core: CRUD + lifecycle ordine/viaggio/fattura
- Seed demo data

**Deliverable (COMPLETATA):**
- ✅ API + DB persistence per flusso core

---

### Phase 2 — V1 App Development (frontend shell + core pages)
**Goal:** UI operativa completa e connessa al backend.

**Steps (COMPLETATA):**
- AppShell (sidebar + topbar + routing)
- Pagine core:
  - ✅ Dashboard
  - ✅ Raccolta Ordini
  - ✅ Planner (vista classica)
  - ✅ Gestione Viaggi
  - ✅ Fatturazione
- UX: toast, skeleton, badge stati, `data-testid`

**Deliverable (COMPLETATA):**
- ✅ UI V1 completa, navigabile e operativa

---

### Phase 3 — Auth + Stabilizzazione CRUD
**Goal:** Autenticazione e hardening.

**Steps (COMPLETATA):**
- Auth token-based (login + me) e protezione route
- CRUD Anagrafiche completo (Clienti/Destinazioni/Veicoli/Autisti/Vettori/Prodotti/Garage)
- Fix unicode escape UI (€, à, →, •)
- Fix filtri “all”

**Deliverable (COMPLETATA):**
- ✅ App autenticata e CRUD stabile

---

### Phase 4 — Export / Hardening
**Goal:** Export essenziali e test.

**Steps (COMPLETATA):**
- ✅ Export Excel ordini (`GET /export/orders`) con download da UI
- ✅ Regression test generale e fix

---

### Phase 5 — V1.1: Listini Rule Editor + Tariff Lookup
**Goal:** Gestire regole contrattuali e precompilare tariffa.

**Backend (COMPLETATA):**
- ✅ `GET /api/pricelists/{id}`
- ✅ `POST /api/pricelists/{id}/items`
- ✅ `GET /api/pricelists/lookup-tariff`

**Frontend (COMPLETATA):**
- ✅ Detail view listino con tabella regole
- ✅ Aggiunta/cancellazione regole
- ✅ Auto lookup tariffa in creazione ordine

> Nota: la cancellazione per indice (V1.1) è stata superata in V1.2 con cancellazione per `item_id`.

---

### Phase 6 — V1.1: Disponibilità Mezzi/Autisti + Indisponibilità Autisti
**Goal:** Migliorare qualità pianificazione.

**Backend (COMPLETATA):**
- ✅ `GET /api/availability/vehicles` (available/busy)
- ✅ `GET /api/availability/drivers` (available/busy/unavailable)
- ✅ `POST/GET/DELETE /api/driver-unavailability`

**Frontend (COMPLETATA):**
- ✅ Planner: pulsanti “Disponibilità” e “Indisponibilità”
- ✅ Pannello disponibilità con status dot
- ✅ Gestione indisponibilità (crea/elimina)
- ✅ Indicatori disponibilità in dialog “Assegna”

---

### Phase 7 — V1.2 Feature A: Listini — Editing regole + `item_id` stabile
**Goal:** Rendere le regole modificabili e referenziabili in modo stabile.

#### 7.1 Backend
**Modifiche dati (COMPLETATA):**
- ✅ Aggiunto `item_id: str` (UUID) a `PriceListItemBase` con `default_factory`.
- ✅ Retrocompatibilità: in `GET /api/pricelists/{id}` vengono assegnati `item_id` alle regole legacy prive del campo.

**Endpoint aggiornati/nuovi (COMPLETATA):**
- ✅ `POST /api/pricelists/{id}/items` → restituisce `item_id`
- ✅ `PUT /api/pricelists/{id}/items/{item_id}` → modifica regola per `item_id`
- ✅ `DELETE /api/pricelists/{id}/items/{item_id}` → elimina per `item_id` (fallback per indice numerico per retrocompatibilità)
- ✅ `GET /api/pricelists/lookup-tariff` → include `item_id` quando `found=true`

#### 7.2 Frontend (PriceListsPage)
**UI (COMPLETATA):**
- ✅ Azione **Modifica** (matita) per riga regola, dialog precompilato
- ✅ CTA “Salva Modifiche” su edit
- ✅ Delete per `item_id`

**Deliverable (COMPLETATA):**
- ✅ Rule editing completo + `item_id` stabile end-to-end

---

### Phase 8 — V1.2 Feature B: Planner V2 — Griglie duali Partenze/Rientri (mockup pag. 16)
**Goal:** Avvicinare il Planner al mockup: due griglie per pagina (top/bottom) con segmentazione per flussi.

**Implementazione (COMPLETATA):**
- ✅ 3 tab principali: **Tutti**, **Partenze**, **Rientri**.
- ✅ Tab **Partenze**: Export (sopra) + Nazionale (sotto)
- ✅ Tab **Rientri**: Import (sopra) + Solo Estero (sotto)
- ✅ Sub-filtro stato (Tutti/Da pianificare/In viaggio/Chiusi) applicato a tutte le viste
- ✅ Azioni **Assegna/Chiudi** funzionanti da qualsiasi griglia

**Deliverable (COMPLETATA):**
- ✅ Planner V2 con griglie duali per mockup

---

### Phase 9 — **V1.3 Feature C: Mappa Viaggi Interattiva (Innovativa)**
**Goal:** Pianificare e visualizzare i viaggi su mappa in modo intuitivo, con dati dimostrativi immediatamente utili.

#### 9.1 Frontend (COMPLETATA)
**Tecnologia:** Leaflet + react-leaflet

**UX / Visualizzazione (COMPLETATA):**
- ✅ Pagina **“Mappa Viaggi”** (route `/mappa`) aggiunta in sidebar
- ✅ Layout “operatore”: **mappa grande + pannello laterale**
- ✅ Pannello laterale “Viaggi attivi” con card interattive:
  - progressivo ordine
  - stato con badge
  - tratta carico → scarico
  - targa + autista
  - barra progresso
  - tariffa
- ✅ Click su card → evidenzia rotta e mostra **pannello Dettaglio** (cliente, date, mezzo, autista, tariffa, tipologia)
- ✅ Filtri:
  - toggle Pianificabili on/off
  - toggle Chiusi on/off
  - filtro per mezzo specifico
  - pulsante Aggiorna
- ✅ **Auto fit-bounds** per centrare la mappa su tutti i punti visibili

**Mappa (COMPLETATA):**
- ✅ Cartografia CARTO light
- ✅ Percorsi curvi (Bézier) colorati per stato:
  - rosso = VIAGGIO
  - giallo = PIANIFICABILE
  - arancione = CHIUSO
- ✅ Marker camion custom per posizione attuale simulata (solo per VIAGGIO)
- ✅ Icona garage speciale con branding teal
- ✅ Marker POI (destinazioni) con tooltip
- ✅ Fix robustezza: filtro route senza coordinate valide (evita NaN LatLng)

#### 9.2 Backend (COMPLETATA)
**Endpoint:**
- ✅ `GET /api/map/trips`

**Dati esposti (COMPLETATA):**
- ✅ `routes[]`: ordini mappabili con carico/scarico, stato, progressivo, mezzo/autista, progress simulato, posizione attuale
- ✅ `poi[]`: destinazioni con coordinate
- ✅ `garages[]`: garage con coordinate
- ✅ `stats`: contatori (in_viaggio, pianificabili, chiusi)

**Coordinate reali dimostrative (COMPLETATA):**
- ✅ 12 destinazioni europee (Italia, Svizzera, Belgio, Olanda, Germania, Francia) + 2 garage

#### 9.3 Seed demo (COMPLETATA)
- ✅ Dati demo coerenti e immediatamente visualizzabili:
  - ordini in VIAGGIO con targa/autista assegnati
  - alcuni viaggi formali in `trips`
  - rotte europee realistiche (es. Lodi → Zurigo)

**Deliverable (COMPLETATA):**
- ✅ Visualizzazione mappa “innovativa” pronta per demo/uso operativo (con dati realistici)

---

## 3) Next Actions (immediate)
**Proposte per V1.4 / V2 (non ancora implementate):**
1. **Mappa (da demo → produzione)**
   - Integrazione tracking reale (GPS) + aggiornamento posizione live
   - Calcolo percorso reale con OSRM/GraphHopper (anziché linea Bézier)
   - Timeline eventi (arrivo carico, partenza, soste, arrivo scarico)
   - Cluster dei POI e layer (solo viaggi, solo POI, solo garage)
2. **Planner → Mappa**
   - CTA “Apri su mappa” dal Planner/Viaggio
   - Evidenziazione singolo mezzo/missione
3. **Export PDF**
   - Istruzioni Operative, CMR, Fattura
4. **Security**
   - JWT + RBAC per endpoint
5. **Dati**
   - Migrazione persistente lat/lng per destinazioni in anagrafica
   - Indici DB su `stato`, `data_ritiro`, `data_consegna`, `targa_motrice`

---

## 4) Criteri di successo
**Raggiunti (V1 + V1.1 + V1.2 + V1.3):**
- ✅ Flusso core completo e stabile
- ✅ Listini rule editor + lookup tariffa + editing regole con `item_id` stabile
- ✅ Disponibilità mezzi/autisti + gestione indisponibilità
- ✅ Planner V2 con griglie duali Partenze/Rientri
- ✅ **Mappa Viaggi** intuitiva e demo-ready con percorsi, mezzi, filtri e dettaglio
