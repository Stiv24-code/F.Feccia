# Allineamento al prototipo — note di UI

> Da Francesco (design), 11 settembre 2026.
> Versione con gli screenshot appaiati: **https://claude.ai/code/artifact/97553908-98f7-4a6e-b790-b5f6611fba7f**
> Questo file è la stessa cosa in forma operativa — file, valori, ordine — perché ti sia comoda
> se la dai in pasto a Claude invece di leggerla.

Ho confrontato schermata per schermata il prototipo (`ui/design/TMS Unificato.dc.html`, commit dbae463)
con il dev su `main`, 27 coppie catturate nello stesso stato. Il grosso è in piedi: vetro, tag,
cornici, **la lastra sopra le tabelle**, la sequenza del lavaggio, il tema scuro. Quello che segue è
il resto, più **otto cose che hai fatto meglio del prototipo e che allineo io** (in fondo).

Se qualcosa è stato fatto così per un motivo tecnico che non conosco, dimmelo e adeguo il prototipo:
diverse di queste scelte le ho fatte su schermate vuote, tu le hai fatte con i dati veri.

---

## Se usi Claude su questo file

Consigli pratici, per esperienza mia su questo repo:

- **Un passo per volta**, nell'ordine in cui sono scritti: i primi due sono token e si portano dietro
  metà delle altre note. Fare tutto in un colpo rende il diff illeggibile.
- **Dopo ogni passo, guarda la schermata** con il tema Glass attivo. Diversi di questi valori si
  vedono solo sul vetro.
- **Non toccare** quello che è nella lista «già allineato» qui sotto: sono valori che ho verificato
  uno per uno e combaciano. Un agente tende a «migliorare» anche quelli.
- I valori con `rgb(...)` e `px` sono misurati sul prototipo con `getComputedStyle`: sono quelli, non
  approssimazioni.
- Dove scrivo un nome di file l'ho verificato nel repo l'11/09. Se il file è stato spostato, cerca il
  componente per nome.

## Già allineato — da non toccare

| cosa | valore verificato |
|---|---|
| Sidebar: vetro, sfocatura, raggio, distacco | `rgba(242,248,255,.42)` · `blur(28px) saturate(1.6)` · `radius 18px` · offset 12px |
| Testata fissa | `rgba(255,255,255,.12)` · `blur(30px) saturate(1.6)` |
| Lastra sopra le tabelle (`PageSlab`) | c'è su Ordini, In arrivo, Template PDF, Anagrafiche, Planner |
| Fondale a mesh | `radial-gradient(620px 480px at 12% 8%, …)` identico |
| Colore dei tag Tipo e Stato | `rgb(91,107,134)` / `rgb(253,236,236)` + bordo `rgb(240,197,192)` |
| Cornici sulle icone di riga nel registro | chip con bordo, come nel prototipo |
| Binario sotto le tab | capsula con pill attiva chiara |
| Sequenza dell'itinerario | `partenza → lavaggio → carico → scarico`, con l'etichetta «prima del carico» |
| Colonna di oggi nel planner | presente (solo più tenue, vedi 4.4) |
| Puntino di stato sulle card del planner | presente |
| Font | Outfit su tutto, con `tabular-nums` sui numeri |

---

## Passo 1 — La tipografia delle tabelle

**Il punto da cui parte tutto.** Nel prototipo ogni cella ha la sua misura; nel dev è `14px` con
`line-height: 20px` su tutto. Da lì arrivano le righe più alte, le colonne che non bastano e le
ellissi su quasi ogni tratta: sistemato questo, quei sintomi rientrano da soli.

**File:** `frontend/src/components/shared/DataTable.tsx` — la tabella nasce da
`<Table className="text-xs md:text-sm">` (riga ~88) e le intestazioni da
`<TableHead className="py-2 text-xs font-medium">` (riga ~92). Le celle le compone ogni pagina nel
suo `renderRow`, quindi conviene definire le misure una volta come classi condivise.

| elemento | ora | atteso |
|---|---|---|
| cella cliente | `14px` · peso 400 · lh 20px | `12.5px` · peso **600** · lh normal |
| cella tratta | `14px` · peso 400 | `12px` · peso 400 · colore secondario |
| cella codice / progressivo | `14px` · peso 500 | `11.5px` · peso **600** · colore secondario |
| cella tariffa | `14px` · peso 400 | `11.5px` · peso **600** |
| testo primario | `rgb(24,39,67)` | `rgb(24,39,66)` |
| testo secondario | — (non esiste) | `rgb(52,67,95)` per tratta, codici, metadati |
| intestazione colonna | `12px` · 500 · spacing normale · Title Case | `10px` · **700** · `letter-spacing .8px` · MAIUSCOLO |
| colore intestazione | `rgb(87,101,127)` | già giusto ✓ |
| altezza riga | 45px | 56px (la riga ha due livelli, vedi passo 4) |

Il senso dei pesi: nel prototipo il peso dice cosa conta. Cliente e tariffa a 600, la tratta a 400.
Con tutto a 400 la riga si legge come un elenco piatto e l'occhio non trova dove appoggiarsi.

## Passo 2 — I token che restano

Sei valori, tutti misurati. I primi due in `frontend/src/index.css`, blocco `.glass` (righe 238–247).

| # | cosa | ora | atteso |
|---|---|---|---|
| 0.6 | `--glass-panel-blur` | `blur(14px) saturate(1.5)` | `blur(14px) saturate(1.6)` |
| 0.6 | `--glass-row-blur` | `blur(10px) saturate(1.4)` | `blur(10px) saturate(1.6)` |
| 0.6 | raggio del pannello | `12px` | `14px` |
| 0.5 | larghezza sidebar | `260px` | `214px` |
| 0.5 | ombra sidebar | `0 14px 40px rgba(31,61,122,.16)` | `0 14px 40px rgba(30,60,120,.14)` |
| 0.7 | badge contatore nelle tab | `bg rgb(244,247,251)` · `color rgb(87,101,127)` | `bg rgb(53,101,175)` · `color #fff` |
| 0.8 | hover di riga | `rgba(244,247,251,.6)` | `rgba(53,101,175,.09)` |

La saturazione a `1.6` è quella che dà al vetro il colore vivo invece del grigio: è lo stesso valore
su tutte le superfici tranne il fondale.

I 46px della sidebar non sono un dettaglio: sono lo spazio che manca alla colonna della tratta in
Ordini.

**File:** `index.css` per i token del vetro, `components/layout/AppShell.js` per la larghezza della
sidebar e i badge.

### 0.13 — L'intestazione della tabella esce dall'angolo arrotondato

Difetto visibile su tutte le tabelle: il contenitore (`div.bg-card`) ha `border-radius: 12px` ma
`overflow: visible`, e il `thead` ha un fondo proprio (`rgba(255,255,255,.45)`) con gli angoli
quadrati. Risultato: lo spigolo dell'intestazione sporge oltre la curva e in alto a sinistra si vede
uno scalino. Sul vetro si nota di più, perché i due fondi hanno trasparenze diverse.

```diff
- <Card className="rounded-xl border shadow-sm" data-testid={testId || 'data-table'}>
+ <Card className="rounded-[14px] border shadow-sm overflow-hidden" data-testid={testId || 'data-table'}>
```

`overflow-hidden` sistema anche l'angolo in basso dell'ultima riga, che ha lo stesso problema.
**File:** `components/shared/DataTable.tsx`, riga ~86.

### 0.9 — Lo scalino fra tag Tipo e tag Stato

I colori sono giusti ✓. Nel prototipo lo stato è volutamente più grande del tipo, perché è il dato
che si cerca per primo:

| | ora | atteso |
|---|---|---|
| tag Tipo | `10px` · padding `3px 9px` · h 28 | `10px` · padding `3px 9px` · **h 21** |
| tag Stato | `10px` · padding `3px 9px` · h 28 | **`11px`** · padding **`5px 12px`** · h 26 |

L'altezza dipende dal `line-height`: a 28px entrambe le pill sembrano lo stesso oggetto.
**File:** `components/shared/StatusBadge.tsx`, `components/shared/TypeBadge.tsx`.

### 0.10 — L'ordine delle voci di menu

Oggi: `Dashboard · Anagrafiche · Listini · Ordini · Planner · Mappa Viaggi · Fatturazione · Utenti`.
Nel prototipo le voci di lavoro quotidiano stanno insieme in alto e le anagrafiche in fondo, perché
si consultano:

```
Dashboard · Ordini · Planner · Mappa
poi: Listini · Fatturazione · Utenti · Anagrafiche
```

Con Anagrafiche in seconda posizione, aprendosi mostra undici sottovoci e separa le tre voci che si
usano tutti i giorni. Nella stessa colonna: la voce **Ordini** nel prototipo porta un puntino blu
quando c'è qualcosa di nuovo in arrivo (0.11), e l'etichetta è **Mappa**, non «Mappa Viaggi» (0.12).
**File:** `components/layout/AppShell.js`.

## Passo 3 — La zona di testa: due righe invece di tre

La lastra c'è e i valori sono quelli giusti ✓ — quello che resta è dentro. Oggi in Ordini la testa
occupa **tre righe più un orfano**: i chip di stato non stanno su una riga (`Scartati 2` va a capo da
solo: gli altri a y=137, lui a y=169) e `Esporta Excel` resta isolato sotto la ricerca.

Due cose:

1. **I chip a 14px non ci stanno.** Con il passo 1 rientrano.
2. **Nel prototipo l'export sta sulla riga delle tab**, in alto a destra accanto alla CTA, non sotto
   la ricerca. Spostandolo lì la barra scende di una riga.

In `DataTable.tsx` l'export è già dentro `<SlabToolbar>` insieme ai filtri (riga ~63): il pezzo da
spostare è quello, nello stesso portale che usa `PageHeaderActions` per la CTA.
**File:** `components/shared/DataTable.tsx`, `components/layout/OrdiniTabsLayout.tsx`.

## Passo 4 — Le colonne del registro

- **Manca la colonna Consegna** (2.2). Nel prototipo Ritiro e Consegna sono affiancate: sono le due
  date su cui si decide cosa pianificare.
- **Manca il prodotto sotto il cliente** (2.3). Nel prototipo la cella cliente ha due livelli: nome e
  sotto, più chiaro, il prodotto — «Crema di latte», «Margarina sfusa». Sul trasporto di liquidi
  alimentari è quello che decide lavaggio e compatibilità della cisterna, per questo sta in tabella e
  non solo nel dettaglio. È anche il motivo dei 56px di riga.

```
ORDINE · CLIENTE (+prodotto) · TRATTA · RITIRO · CONSEGNA · TARIFFA · TIPO · STATO · azioni
```

- **Manca il filtro sul periodo** (2.4): nel prototipo «Ritiro: 1 lug → 31 lug» accanto alla ricerca.
  Con 94 ordini serve quasi sempre.
- **Il filtro mostra il nome del campo invece del valore** (2.5): dice «Tipologia», nel prototipo dice
  «Tutti i tipi». Con il valore si capisce da fuori se un filtro è acceso. Stessa cosa per il
  placeholder della ricerca: `Cerca ordine, cliente, tratta…` invece di `Cerca ordini…`.

### In arrivo (2.6, 2.7)

- Le date occupano **quattro righe per cella**: `12/09/26 15:00 - 13/09/26 04:00 CEST` porta la riga
  oltre gli 80px. Nel prototipo è `31/07 07–10` — giorno e finestra, senza anno e senza fuso: l'anno
  serve solo quando cambia, il fuso in una lista interna non aggiunge niente. Nella stessa tabella la
  pill «Da confermare su portale» va a capo su tre righe.
- Il **chip del canale è giallo ambra** e si legge come un avviso: nel prototipo è neutro (fondo
  grigio chiaro e bordo), perché in quella tabella il colore è riservato allo stato.

## Passo 5 — Le azioni nel dettaglio ordine

Oggi dalla schermata di dettaglio non si riesce a pianificare l'ordine né a saltare al suo viaggio:
bisogna tornare in elenco. Nel prototipo la testata cambia con lo stato:

| stato | azioni |
|---|---|
| Da pianificare | `+ Pianifica` · `Elimina` |
| Pianificato | `Apri viaggio VG-111 →` |
| In viaggio | `Apri viaggio VG-114 →` · `Export` |
| Consegnato | `Apri viaggio VG-107 →` · `Export` |

Piccola cosa nella stessa riga: `export` e `import` sono minuscoli, nel prototipo `Export`.
**File:** `frontend/src/pages/OrderDetailPage.tsx`.

### 3.3 — Le date dell'itinerario

Nel registro sono già `gg/mm` ✓. Nell'itinerario del dettaglio compare `2026-09-15`, e lo stesso nei
dialog dell'autista: sembra che lì la data arrivi stampata così com'è, senza passare da
`toLocaleDateString('it-IT')`. Nel prototipo è `06/07/2026` con la finestra orario sotto.
**File:** `components/shared/RouteItinerary.tsx`.

### 3.4 — Prodotto e peso

Finiscono in una riga sotto «NOTE»:
`Prodotto: Fully refined palm oil/PALM OIL MB | Kg: 25000 | Da richiesta pdf…`. Nel prototipo c'è una
card «DETTAGLI ORDINE» con Prodotto e Peso come campi affiancati e la nota del cliente separata: sono
i due dati che l'autista chiede al telefono.

## Passo 6 — I due interventi grossi

Conviene affrontarli insieme, perché il primo serve al secondo.

### 3.1 — Il cruscotto cliente nel dettaglio ordine

Nel prototipo accanto al dettaglio c'è un pannello navy che risponde a «con chi sto trattando»:

- la nota operativa del cliente — «Margarine e oli: scarico a caldo, serpentine richieste in inverno»
- tre numeri: ordini/anno, fatturato, ordini/mese
- il grafico degli ultimi 12 mesi
- **le tariffe pattuite**, con evidenziata quella di questa tratta
- gli ultimi viaggi del cliente con il loro stato

Serve a poter dire al telefono «su Conselice → Verona siamo a 480» senza aprire il listino. Il
collegamento `Vai al cruscotto cliente →` c'è, ma porta su un'altra pagina: mentre si guarda l'ordine
quei numeri non si vedono. È il motivo per cui nel prototipo il dettaglio è su due colonne.

### 4.6 — L'assegnazione a schermata intera

Oggi si assegna in un dialog di **512 × 1122px**, che essendo più alto della finestra va scorso. Nel
prototipo «Pianifica viaggio» è una schermata intera su due colonne:

- **sinistra**: ordine, tariffa, mappa grande, percorsi alternativi, itinerario, chi effettua il
  trasporto, ordini nel viaggio
- **destra**: il cruscotto cliente con tariffe pattuite e ultimi viaggi

Il motivo è che per assegnare si guardano insieme percorso, redditività e storico del cliente: in
512px di larghezza quelle tre cose non ci stanno. Nella modale infatti mancano la tariffa e la
redditività stimata (4.8) — si assegna senza vedere se il viaggio conviene.

Dentro la stessa schermata, due cose più piccole:

- **4.7 — i percorsi alternativi**: la scelta c'è, mostri più mappe. Nel prototipo è una mappa sola
  con due o tre card selezionabili sotto, con km, durata e differenza dal più breve — «Percorso A ·
  425 km · 2h15 · consigliato, più breve» accanto a «Percorso B · 462 km · +37 km». Così si
  confrontano i numeri guardando un disegno solo, e la scelta resta visibile anche dopo
  («Percorso A · scelto in pianificazione»).
- **4.9 — i selettori**: nel prototipo garage e lavaggio si scelgono da una lista con ricerca e righe
  ricche — nome, tipo di lavaggio o indirizzo, distanza dal carico, stalli: «CleanTank Piacenza ·
  Lavaggio alimentare EFTCO», «Garage Milano Opera · SP ex SS35 km 2 · 8 stalli». Con centinaia di
  stazioni in anagrafica, un elenco di soli nomi lascia la scelta al buio.

**File:** `components/planner/AssignOrderDialog.tsx` (contenitore),
`components/planner/AssignOrderForm.tsx` (corpo, 427 righe — il corpo si riusa quasi tutto).

## Passo 7 — Il planner

- **4.1 — le card hanno il fondo tinto per stato**: rosa, giallo, azzurro, verde. Con sette colonne
  piene la settimana diventa molto colorata e il colore perde forza di segnale. Nel prototipo la card
  resta bianca con bordo neutro e lo stato sta nel puntino in alto a destra — il puntino ce l'hai già
  ✓, si tratta di togliere la tinta al fondo.
- **4.2 — la card non dice il prodotto**: al suo posto c'è la targa. Sul liquido alimentare il
  prodotto è il vincolo; la targa si vede bene nel dettaglio.
- **4.3 — la colonna di oggi** è più tenue: nel prototipo l'intestazione è navy pieno con testo bianco
  e la colonna ha un fondo appena tinto, così si trova subito scorrendo la settimana.
- **4.4 — manca il contatore sulla freccia «indietro»**: quanti ordini sono rimasti non pianificati
  nelle settimane passate. Serve a non lasciarli indietro senza accorgersene.

**File:** `components/planner/PlannerCalendar.tsx`, `pages/PlannerPage.tsx`.

## Passo 8 — La mappa

La pagina c'è e il pannello di vetro a destra è al posto giusto ✓. Nello stato attuale però è
difficile leggerci qualcosa:

- **5.1** — viene disegnato un pin per ogni punto di ogni ordine, tutti uguali, su tutta l'Europa: si
  coprono a vicenda. Servirebbero il raggruppamento per zona a zoom basso e l'inquadratura sui viaggi
  attivi invece che sull'intero continente.
- **5.2** — manca il tracciato: nel prototipo ogni viaggio è una linea colorata che segue le strade,
  con il mezzo lungo il percorso e un'etichetta «VG-114 · Marco Bianchi · 62%». Senza linee la mappa
  dice dove sono i punti ma non dove stanno andando i mezzi.
- **5.3** — il pannello dice «Viaggi sulla mappa (0) — Nessun viaggio con i filtri attivi» mentre i
  chip sopra contano 11 da pianificare e 1 chiuso: sembra che filtri e lista non si parlino.
- **5.4** — i filtri sono «Da pianificare» e «Chiusi»; nel prototipo sono **In viaggio** e
  **Pianificati**, cioè quelli che ha senso guardare su una mappa: un ordine da pianificare non ha
  ancora un percorso da disegnare.

Quando il pannello avrà contenuto, nel prototipo ogni riga porta cliente, stato, tratta, autista,
barra di avanzamento, «in orario» o «rischio ritardo» ed ETA.
**File:** `frontend/src/pages/MapPage.tsx`.

## Passo 9 — Le ore di guida dell'autista

Qui mi ero sbagliato nella prima versione delle note: le informazioni **ci sono**. Dalle azioni di
riga si aprono «Viaggi assegnati» (targa, periodo, stato) e «Ferie e assenze» con la tabella dei
periodi e il pulsante per aggiungerne — quest'ultimo fa **più** del prototipo, dove le ferie si
leggono ma non si gestiscono.

Nel prototipo quelle stesse cose stanno in una scheda sola, con due aggiunte: patenti e abilitazioni,
e soprattutto **le ore di guida** — tre barre (oggi 6,5/9 h, settimana 41/56 h, ultime due settimane
88/120 h) con i vincoli normativi accanto.

Il pezzo che conta è quello: è la risposta a «posso dargli un altro viaggio?», e oggi per farsela
bisogna aprire due finestre e tenere i conti a mente. Se preferisci restare sui dialog invece di fare
la scheda per me va bene — l'importante è che le ore di guida si vedano da qualche parte.

---

## Dove sei più avanti del prototipo — queste le allineo io

1. **`tabular-nums`** su codici e tariffe: le colonne di numeri si allineano meglio delle mie.
2. **Tema scuro con token propri**, invece del filtro sul chiaro che uso io. La tua strada è quella giusta.
3. **Legenda dei colori nel planner**: chiarisce il codice colore a chi apre la pagina per la prima volta.
4. **«Scansiona casella»** in In arrivo: l'azione di andare a prendere le mail invece di aspettarle.
5. **Ctrl + K** per la ricerca globale, annunciato in testata.
6. **Portale clienti** dal login.
7. **Filtro Ritiro/Consegna** in In arrivo: nel prototipo il filtro periodo agisce solo sul ritiro.
8. **Gestione di ferie e assenze** con aggiunta dei periodi.

Nel prototipo adotto anche il **progressivo `26/R001`** al posto del mio `ORD-2630` (è quello che
finisce sui documenti) e la tua impostazione sulla dashboard senza `+ Nuovo ordine`.

Restano da parte, perché ci devo ancora ragionare io: le liste delle anagrafiche (colonne,
contenitore, ricerca), il commutatore lista/griglia nel registro, il login e il tema scuro.

## Una cosa da decidere insieme

**I breadcrumb.** Nel prototipo li ho tolti da tutte le schermate — la freccia indietro accanto al
codice basta e quella riga in più alza tutto il contenuto — tu li tieni. Scegliamone uno e lo teniamo
per l'intera applicazione: a me va bene anche rimetterli, purché sia una scelta sola.

## Checklist

- [ ] **1** Tipografia delle tabelle: misure per ruolo, pesi, colore secondario, intestazioni a etichetta
- [ ] **2** Token: saturazione 1.6, raggio 14, sidebar 214, ombra, badge blu, hover azzurro, scalino dei tag
- [ ] **2b** `overflow-hidden` sul contenitore della tabella (angolo dell'intestazione)
- [ ] **2c** Ordine delle voci di menu, puntino su Ordini, etichetta «Mappa»
- [ ] **3** Export sulla riga delle tab (la testa scende a due righe)
- [ ] **4** Colonne: Consegna, prodotto sotto il cliente, filtro periodo, valore nei filtri
- [ ] **4b** In arrivo: date compatte, pill su una riga, chip canale neutro
- [ ] **5** Dettaglio ordine: azioni per stato, date dell'itinerario, prodotto e peso come campi
- [ ] **6** Cruscotto cliente + assegnazione a schermata intera (percorsi come card, selettori con ricerca)
- [ ] **7** Planner: card neutre, prodotto, colonna di oggi, contatore sulla freccia
- [ ] **8** Mappa: raggruppamento, tracciati, pannello, filtri
- [ ] **9** Ore di guida dell'autista

---

*Le misure vengono da `getComputedStyle` sui due lati, catture del 10 settembre 2026 a 1440 × 900 con
tema Glass attivo. Se un valore non ti torna chiedimelo: ho i due screenshot appaiati per ogni punto.*
