// Package seeddemo ports backend/seed_demo.py: a one-shot, dev/demo-only
// seed with realistic FECCIA F.lli data (customers, destinations, fleet,
// drivers, pricelists, orders, trips, invoices). NOT for production use.
//
// Shared between cmd/seed (manual `make seed-demo`) and internal/app's
// startup hook (automatic on an empty DB, gated behind IS_LOCAL=true — see
// internal/app/seed.go) so both entry points stay in sync.
package seeddemo

import (
	crand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"
)

// Seed mirrors seed_demo.py's `seed()`. Exits early with an error if a user
// with SEED_ADMIN_EMAIL already exists — refuses to reseed a non-empty DB.
func Seed(db *gorm.DB) error {
	adminEmail := getEnv("SEED_ADMIN_EMAIL", "admin@feccia.it")

	var existing models.User
	if err := db.Where("login = ?", adminEmail).First(&existing).Error; err == nil {
		return fmt.Errorf("un utente con email %s esiste gia' nel DB. Il seed e' destinato solo a un DB vuoto", adminEmail)
	}

	// ─────────────────────────── UTENTI ───────────────────────────
	adminPassword := getEnv("SEED_ADMIN_PASSWORD", "")
	if adminPassword == "" {
		adminPassword = randomToken()
	}
	plannerPassword := randomToken()
	operatorePassword := randomToken()
	amminPassword := randomToken()

	users := []models.User{
		mustUser(adminEmail, "Admin Feccia", utils.RoleAdmin, adminPassword),
		mustUser("m.rossi@feccia.it", "Marco Rossi", utils.RolePlanner, plannerPassword),
		mustUser("l.bianchi@feccia.it", "Laura Bianchi", utils.RoleOperatore, operatorePassword),
		mustUser("g.ferrari@feccia.it", "Giuseppe Ferrari", utils.RoleAmministrazione, amminPassword),
	}
	if err := db.Create(&users).Error; err != nil {
		return fmt.Errorf("utenti: %w", err)
	}
	fmt.Printf("✓ %d utenti\n", len(users))

	// ─────────────────────────── CLIENTI ───────────────────────────
	customers := []models.Customer{
		{RagioneSociale: "VOG PRODUCTS SOC. AGRIC. COOP.", Citta: "Laives", Provincia: "BZ", Nazione: "Italia", PartitaIva: "IT00124290214", Telefono: "+39 0471 155689", Email: "ordini@vog.it", CondizioniPagamento: "BB 30gg DF FM"},
		{RagioneSociale: "PARMALAT S.p.A.", Citta: "Parma", Provincia: "PR", Nazione: "Italia", PartitaIva: "IT01234567890", Telefono: "+39 0521 803322", Email: "logistica@parmalat.it", CondizioniPagamento: "BB 60gg DF FM"},
		{RagioneSociale: "LIVERAS GROUP SPA", Citta: "Brescia", Provincia: "BS", Nazione: "Italia", PartitaIva: "IT09876543210", Telefono: "+39 030 2429801", Email: "trasporti@liveras.it", CondizioniPagamento: "BB 30gg DF FM"},
		{RagioneSociale: "BARILLA G. e R. FRATELLI S.p.A.", Citta: "Parma", Provincia: "PR", Nazione: "Italia", PartitaIva: "IT04578750232", Telefono: "+39 0521 2621", Email: "shipping@barilla.com", CondizioniPagamento: "BB 60gg DF FM"},
		{RagioneSociale: "NESTLÉ ITALIANA S.p.A.", Citta: "Milano", Provincia: "MI", Nazione: "Italia", PartitaIva: "IT03523490156", Telefono: "+39 02 81817111", Email: "supply.chain@nestle.it", CondizioniPagamento: "BB 90gg DF FM"},
		{RagioneSociale: "FERRERO S.p.A.", Citta: "Alba", Provincia: "CN", Nazione: "Italia", PartitaIva: "IT00170380048", Telefono: "+39 0173 3131", Email: "logistica@ferrero.com", CondizioniPagamento: "BB 60gg DF FM"},
		{RagioneSociale: "MUTTI S.p.A.", Citta: "Montechiarugolo", Provincia: "PR", Nazione: "Italia", PartitaIva: "IT00186420341", Telefono: "+39 0521 652511", Email: "trasporti@mutti.it", CondizioniPagamento: "BB 30gg DF FM"},
		{RagioneSociale: "GRANAROLO S.p.A.", Citta: "Bologna", Provincia: "BO", Nazione: "Italia", PartitaIva: "IT00549090371", Telefono: "+39 051 4170111", Email: "spedizioni@granarolo.it", CondizioniPagamento: "BB 30gg DF FM"},
		{RagioneSociale: "LACTALIS ITALIA S.r.l.", Citta: "Collecchio", Provincia: "PR", Nazione: "Italia", PartitaIva: "IT02109890340", Telefono: "+39 0521 301411", Email: "logistics@lactalis.it", CondizioniPagamento: "BB 60gg DF FM"},
		{RagioneSociale: "CONSERVE ITALIA SOC. COOP.", Citta: "San Lazzaro di Savena", Provincia: "BO", Nazione: "Italia", PartitaIva: "IT02858450584", Telefono: "+39 051 6228311", Email: "ordini@conserveitalia.it", CondizioniPagamento: "BB 30gg DF FM"},
		{RagioneSociale: "DE CECCO S.p.A.", Citta: "Fara San Martino", Provincia: "CH", Nazione: "Italia", PartitaIva: "IT00126960690", Telefono: "+39 0872 9891", Email: "export@dececco.it", CondizioniPagamento: "BB 60gg DF FM"},
		{RagioneSociale: "BONDUELLE ITALIA S.r.l.", Citta: "San Paolo d'Argon", Provincia: "BG", Nazione: "Italia", PartitaIva: "IT12345678901", Telefono: "+39 035 4252411", Email: "orders@bonduelle.it", CondizioniPagamento: "BB 30gg DF FM"},
		{RagioneSociale: "DANONE ITALIA S.p.A.", Citta: "Milano", Provincia: "MI", Nazione: "Italia", PartitaIva: "IT11345670155", Telefono: "+39 02 698991", Email: "logistics@danone.it", CondizioniPagamento: "BB 60gg DF FM"},
		{RagioneSociale: "HEINEKEN ITALIA S.p.A.", Citta: "Pollein", Provincia: "AO", Nazione: "Italia", PartitaIva: "IT05765750966", Telefono: "+39 0165 2361", Email: "distribuzione@heineken.it", CondizioniPagamento: "BB 30gg DF FM"},
		{RagioneSociale: "SAMMONTANA S.p.A.", Citta: "Empoli", Provincia: "FI", Nazione: "Italia", PartitaIva: "IT04261920482", Telefono: "+39 0571 70991", Email: "freddo@sammontana.it", CondizioniPagamento: "BB 30gg DF FM"},
	}
	for i := range customers {
		customers[i].ID = uuid.New()
		customers[i].RichiedeRifOrdine = randBelow(3) == 0
		customers[i].Active = true
	}
	if err := db.Create(&customers).Error; err != nil {
		return fmt.Errorf("clienti: %w", err)
	}
	fmt.Printf("✓ %d clienti\n", len(customers))

	// ─────────────────────────── DESTINAZIONI ───────────────────────────
	// Caricate da un export reale Visirun (POI Carico-Scarico) invece di una
	// manciata di città curate a mano — vedi poi_data.go.
	destinations, err := loadDestinationsPOI()
	if err != nil {
		return fmt.Errorf("destinazioni: %w", err)
	}
	if err := db.Create(&destinations).Error; err != nil {
		return fmt.Errorf("destinazioni: %w", err)
	}
	fmt.Printf("✓ %d destinazioni\n", len(destinations))

	// ─────────────────────────── VEICOLI ───────────────────────────
	vehicles := []models.Vehicle{
		{Targa: "KX300X", TipoVeicolo: "motrice", Marca: "IVECO", Modello: "S-Way 490", PortataKg: 26000},
		{Targa: "AB123CD", TipoVeicolo: "motrice", Marca: "MAN", Modello: "TGX 18.510", PortataKg: 24000},
		{Targa: "MN012OP", TipoVeicolo: "motrice", Marca: "SCANIA", Modello: "R500", PortataKg: 26000},
		{Targa: "FT456WE", TipoVeicolo: "motrice", Marca: "MERCEDES", Modello: "Actros 1845", PortataKg: 25000},
		{Targa: "GH789RT", TipoVeicolo: "motrice", Marca: "VOLVO", Modello: "FH 500", PortataKg: 26000},
		{Targa: "LP321QZ", TipoVeicolo: "motrice", Marca: "DAF", Modello: "XG+ 530", PortataKg: 25000},
		{Targa: "VB654SD", TipoVeicolo: "motrice", Marca: "IVECO", Modello: "S-Way 460", PortataKg: 24000},
		{Targa: "NM987FG", TipoVeicolo: "motrice", Marca: "RENAULT", Modello: "T 480", PortataKg: 24000},
		{Targa: "EF456GH", TipoVeicolo: "rimorchio_isotermico", Marca: "KRONE", Modello: "Cool Liner", PortataKg: 28000},
		{Targa: "IJ789KL", TipoVeicolo: "rimorchio", Marca: "SCHMITZ", Modello: "S.KO Cool", PortataKg: 25000},
		{Targa: "TY123WQ", TipoVeicolo: "rimorchio_isotermico", Marca: "LAMBERET", Modello: "SR2 Green", PortataKg: 27000},
		{Targa: "OP456ER", TipoVeicolo: "rimorchio_isotermico", Marca: "CARRIER", Modello: "Transicold Vector", PortataKg: 26000},
		{Targa: "HJ789UI", TipoVeicolo: "rimorchio", Marca: "KÖGEL", Modello: "Cargo Rail", PortataKg: 25000},
		{Targa: "ZX321CV", TipoVeicolo: "container", Marca: "MAERSK", Modello: "20ft Reefer", PortataKg: 22000},
	}
	for i := range vehicles {
		vehicles[i].ID = uuid.New()
		vehicles[i].Anno = randRange(2019, 2025)
		if vehicles[i].TipoVeicolo == "motrice" {
			vehicles[i].Scompartature = 1
		} else {
			vehicles[i].Scompartature = randRange(1, 3)
		}
		vehicles[i].Active = true
	}
	if err := db.Create(&vehicles).Error; err != nil {
		return fmt.Errorf("veicoli: %w", err)
	}
	fmt.Printf("✓ %d veicoli\n", len(vehicles))

	// ─────────────────────────── AUTISTI ───────────────────────────
	drivers := []models.Driver{
		{Nome: "Marco", Cognome: "Rossi", Patente: "CE", Telefono: "+39 333 1234567"},
		{Nome: "Luca", Cognome: "Bianchi", Patente: "CE", Telefono: "+39 334 7654321"},
		{Nome: "Giuseppe", Cognome: "Verdi", Patente: "CE", Telefono: "+39 335 1112233"},
		{Nome: "Antonio", Cognome: "Neri", Patente: "CE+ADR", Telefono: "+39 336 4445566"},
		{Nome: "Franco", Cognome: "Colombo", Patente: "CE", Telefono: "+39 337 7788990"},
		{Nome: "Roberto", Cognome: "Ricci", Patente: "CE+ADR", Telefono: "+39 338 1122334"},
		{Nome: "Alessandro", Cognome: "Moretti", Patente: "CE", Telefono: "+39 339 5566778"},
		{Nome: "Stefano", Cognome: "Barbieri", Patente: "CE", Telefono: "+39 340 9900112"},
		{Nome: "Davide", Cognome: "Fontana", Patente: "CE+ADR", Telefono: "+39 341 3344556"},
		{Nome: "Paolo", Cognome: "Esposito", Patente: "CE", Telefono: "+39 342 7788901"},
	}
	for i := range drivers {
		drivers[i].ID = uuid.New()
		scadenza := fmt.Sprintf("2027-%02d-%02d", randRange(1, 12), randRange(1, 28))
		drivers[i].ScadenzaPatente = &scadenza
		drivers[i].Active = true
	}
	if err := db.Create(&drivers).Error; err != nil {
		return fmt.Errorf("autisti: %w", err)
	}
	fmt.Printf("✓ %d autisti\n", len(drivers))

	// ─────────────────────────── VETTORI ───────────────────────────
	carriers := []models.Carrier{
		{RagioneSociale: "Trasporti Bianchi SRL", Citta: "Milano", Telefono: "+39 02 1234567"},
		{RagioneSociale: "Logistica Express S.p.A.", Citta: "Roma", Telefono: "+39 06 9876543"},
		{RagioneSociale: "EuroTrans GmbH", Citta: "Monaco di Baviera", Telefono: "+49 89 123456"},
		{RagioneSociale: "TransAlp Logistics AG", Citta: "Zurigo", Telefono: "+41 44 5556677"},
		{RagioneSociale: "BeneLux Transport B.V.", Citta: "Rotterdam", Telefono: "+31 10 1234567"},
	}
	for i := range carriers {
		carriers[i].ID = uuid.New()
		carriers[i].Active = true
	}
	if err := db.Create(&carriers).Error; err != nil {
		return fmt.Errorf("vettori: %w", err)
	}
	fmt.Printf("✓ %d vettori\n", len(carriers))

	// ─────────────────────────── PRODOTTI ───────────────────────────
	products := []models.Product{
		{Codice: "MELA", Descrizione: "Mela", UnitaMisura: "Kg"},
		{Codice: "COCOA", Descrizione: "Cocoa Butter", UnitaMisura: "Kg"},
		{Codice: "LATTE", Descrizione: "Latte UHT", UnitaMisura: "Lt"},
		{Codice: "PASTA", Descrizione: "Pasta alimentare", UnitaMisura: "Kg"},
		{Codice: "OLIO", Descrizione: "Olio d'oliva", UnitaMisura: "Lt"},
		{Codice: "POMOD", Descrizione: "Pomodori pelati", UnitaMisura: "Kg"},
		{Codice: "FORM", Descrizione: "Formaggio Parmigiano", UnitaMisura: "Kg"},
		{Codice: "BIRRA", Descrizione: "Birra in fusti", UnitaMisura: "Lt"},
		{Codice: "SURG", Descrizione: "Surgelati misti", UnitaMisura: "Kg"},
		{Codice: "YOGURT", Descrizione: "Yogurt fresco", UnitaMisura: "Kg"},
		{Codice: "CIOCCO", Descrizione: "Cioccolato / Praline", UnitaMisura: "Kg"},
		{Codice: "VERD", Descrizione: "Verdure fresche", UnitaMisura: "Kg"},
	}
	for i := range products {
		products[i].ID = uuid.New()
		products[i].Active = true
	}
	if err := db.Create(&products).Error; err != nil {
		return fmt.Errorf("prodotti: %w", err)
	}
	fmt.Printf("✓ %d prodotti\n", len(products))

	// ─────────────────────────── GARAGE, CATEGORIE, COSTI ───────────────────────────
	garages := []models.Garage{
		{ID: uuid.New(), Nome: "Garage Feccia F.lli - Lodi", Indirizzo: "Via Industriale 15", Citta: "Lodi", Lat: geoPtr(45.3138), Lng: geoPtr(9.5032), Note: "Sede principale", Active: true},
		{ID: uuid.New(), Nome: "Deposito Milano Rho", Indirizzo: "Via Sempione 220", Citta: "Milano", Lat: geoPtr(45.5306), Lng: geoPtr(9.0393), Note: "Deposito secondario", Active: true},
		{ID: uuid.New(), Nome: "Hub Verona Interporto", Indirizzo: "Viale del Lavoro 4", Citta: "Verona", Lat: geoPtr(45.3852), Lng: geoPtr(10.9296), Note: "Hub intermodale", Active: true},
	}
	if err := db.Create(&garages).Error; err != nil {
		return fmt.Errorf("garage: %w", err)
	}

	// Caricati dallo stesso export Visirun (POI Lavaggio) — vedi poi_data.go.
	washStations, err := loadWashStationsPOI()
	if err != nil {
		return fmt.Errorf("stazioni lavaggio: %w", err)
	}
	if err := db.Create(&washStations).Error; err != nil {
		return fmt.Errorf("stazioni lavaggio: %w", err)
	}

	categories := []models.TransportCategory{
		{ID: uuid.New(), Nome: "Standard", Descrizione: "Trasporto standard", Active: true},
		{ID: uuid.New(), Nome: "GMP", Descrizione: "Good Manufacturing Practice", Active: true},
		{ID: uuid.New(), Nome: "Kosher", Descrizione: "Trasporto Kosher certificato", Active: true},
		{ID: uuid.New(), Nome: "ADR", Descrizione: "Merci pericolose", Active: true},
		{ID: uuid.New(), Nome: "Temperatura controllata", Descrizione: "Trasporto refrigerato -18°C / +4°C", Active: true},
	}
	if err := db.Create(&categories).Error; err != nil {
		return fmt.Errorf("categorie trasporto: %w", err)
	}

	accCosts := []models.AccessoryCost{
		{ID: uuid.New(), Nome: "Sosta", Descrizione: "Costo sosta aggiuntiva (>2h)", CostoDefault: 150, Active: true},
		{ID: uuid.New(), Nome: "Lavaggio", Descrizione: "Lavaggio cisterna/rimorchio", CostoDefault: 200, Active: true},
		{ID: uuid.New(), Nome: "Fuel Surcharge", Descrizione: "Adeguamento carburante mensile", CostoDefault: 0, Active: true},
		{ID: uuid.New(), Nome: "Deviazione", Descrizione: "Deviazione percorso extra", CostoDefault: 100, Active: true},
		{ID: uuid.New(), Nome: "Consegna notturna", Descrizione: "Supplemento consegna 22:00-06:00", CostoDefault: 250, Active: true},
		{ID: uuid.New(), Nome: "Pedaggio", Descrizione: "Rimborso pedaggi autostradali", CostoDefault: 0, Active: true},
	}
	if err := db.Create(&accCosts).Error; err != nil {
		return fmt.Errorf("costi accessori: %w", err)
	}
	fmt.Printf("✓ %d garage, %d stazioni lavaggio, %d categorie, %d costi accessori\n", len(garages), len(washStations), len(categories), len(accCosts))

	// Counters are NOT pre-seeded: database.NextSequence upserts a counter
	// at 1 on first use, and the demo orders/invoices below use their own
	// explicit "25/xxxx" progressivi (a different year prefix than any real
	// order created afterwards), so there's no collision either way —
	// matches Python's net effect exactly.

	// ─────────────────────────── LISTINI ───────────────────────────
	// Con le destinazioni ora caricate dall'export Visirun (ordine non più
	// legato a città specifiche note), le regole prendono semplicemente le
	// prime N destinazioni distinte invece di indici con nome (laivesID ecc.).
	caricoA := &destinations[0].ID
	scaricoA := &destinations[1].ID
	scaricoB := &destinations[2].ID
	caricoB := &destinations[3].ID
	melaID := &products[0].ID
	pastaID := &products[3].ID
	cioccoID := &products[10].ID

	pricelists := []models.PriceList{
		{
			ID: uuid.New(), ClienteID: customers[0].ID,
			DataInizio: "2025-01-01", DataFine: "2025-12-31", Note: "Listino annuale 2025", InUso: true, Active: true,
			Items: []models.PriceListItem{
				{ID: uuid.New(), ProdottoID: melaID, DestinazioneCaricoID: caricoA, DestinazioneScaricoID: scaricoA, Tariffa: 2150, TipoTariffa: "forfait", UnitaPeso: "Kg", TipoTrasporto: "stradale", PercAdeguamentoCarburante: 4.18},
				{ID: uuid.New(), ProdottoID: melaID, DestinazioneCaricoID: caricoA, DestinazioneScaricoID: scaricoB, Tariffa: 1800, TipoTariffa: "forfait", UnitaPeso: "Kg", TipoTrasporto: "stradale", PercAdeguamentoCarburante: 4.18},
				{ID: uuid.New(), DestinazioneCaricoID: caricoA, Tariffa: 2.5, TipoTariffa: "euro_kg", RangePesoMin: 10000, RangePesoMax: 26000, UnitaPeso: "Kg", MinimoTassabile: 15000, TipoTrasporto: "stradale", PercAdeguamentoCarburante: 3.5},
			},
		},
		{
			ID: uuid.New(), ClienteID: customers[3].ID,
			DataInizio: "2025-01-01", DataFine: "2025-12-31", Note: "Listino Barilla 2025", InUso: true, Active: true,
			Items: []models.PriceListItem{
				{ID: uuid.New(), ProdottoID: pastaID, DestinazioneCaricoID: caricoB, DestinazioneScaricoID: scaricoA, Tariffa: 2400, TipoTariffa: "forfait", UnitaPeso: "Kg", TipoTrasporto: "stradale", PercAdeguamentoCarburante: 3.8},
				{ID: uuid.New(), ProdottoID: pastaID, DestinazioneCaricoID: caricoB, Tariffa: 1950, TipoTariffa: "forfait", UnitaPeso: "Kg", TipoTrasporto: "stradale", PercAdeguamentoCarburante: 3.8},
			},
		},
		{
			ID: uuid.New(), ClienteID: customers[5].ID,
			DataInizio: "2025-06-01", DataFine: "2026-05-31", Note: "Listino Ferrero semestrale", InUso: true, Active: true,
			Items: []models.PriceListItem{
				{ID: uuid.New(), ProdottoID: cioccoID, Tariffa: 3.2, TipoTariffa: "euro_kg", RangePesoMin: 5000, RangePesoMax: 24000, UnitaPeso: "Kg", MinimoTassabile: 12000, TipoTrasporto: "stradale", PercAdeguamentoCarburante: 4.0},
			},
		},
	}
	if err := db.Create(&pricelists).Error; err != nil {
		return fmt.Errorf("listini: %w", err)
	}
	fmt.Printf("✓ %d listini con regole\n", len(pricelists))

	// ─────────────────────────── ORDINI ───────────────────────────
	// Solo questi 5 stati esistono per un ordine: "fatturato" non è uno stato
	// a sé (vedi sezione FATTURE più sotto: è fattura_id != "" su un ordine
	// CHIUSO, non un valore di Stato separato).
	statiDist := weighted(map[string]int{"PIANIFICABILE": 10, "PIANIFICATO": 8, "VIAGGIO": 15, "CHIUSO": 23, "SCARTATO": 4})
	tipologie := weighted(map[string]int{"export": 6, "import": 5, "nazionale": 6, "solo_estero": 3})
	tariffe := []float64{950, 1200, 1500, 1800, 2000, 2100, 2150, 2200, 2350, 2500, 2600, 2800, 3000, 3200, 3500}
	tipoTariffe := weighted(map[string]int{"forfait": 4, "euro_kg": 1})
	categorieTrasporto := weighted(map[string]int{"Standard": 3, "GMP": 1, "Kosher": 1, "Temperatura controllata": 1})
	oraRitiroDa := []string{"05:00", "06:00", "07:00", "08:00"}
	oraRitiroA := []string{"08:00", "10:00", "12:00"}
	oraConsegnaDa := []string{"06:00", "08:00", "14:00"}
	oraConsegnaA := []string{"12:00", "16:00", "18:00", "20:00"}

	// Ancorato a "oggi" (non a un anno fisso) così gli ordini seminati
	// ricadono subito nella settimana corrente del calendario Planner,
	// invece di finire in un anno passato non più visibile di default.
	today := time.Now()

	var motrici []models.Vehicle
	for _, v := range vehicles {
		if v.TipoVeicolo == "motrice" {
			motrici = append(motrici, v)
		}
	}
	var trailers []models.Vehicle
	for _, v := range vehicles {
		if v.TipoVeicolo != "motrice" {
			trailers = append(trailers, v)
		}
	}

	orders := make([]models.Order, 0, 60)
	for i := 0; i < 60; i++ {
		seq := i + 1
		cust := pick(customers)
		carico := pick(destinations)
		scarico := pick(destinations)
		for scarico.ID == carico.ID {
			scarico = pick(destinations)
		}
		prod := pick(products)
		stato := pick(statiDist)
		// Spread: da ~3 settimane fa a ~7 settimane nel futuro, centrato su
		// oggi, così la settimana corrente del calendario Planner è sempre
		// popolata fin dal primo avvio.
		ritiro := today.AddDate(0, 0, randRange(-20, 45))
		consegna := ritiro.AddDate(0, 0, randRange(1, 3))

		rifOrdineCliente := ""
		if randBelow(3) == 0 {
			rifOrdineCliente = fmt.Sprintf("%d-%d", randRange(4500000, 4600000), randRange(10, 99))
		}

		order := models.Order{
			ID: uuid.New(), Progressivo: fmt.Sprintf("%s/%04d", today.Format("06"), seq),
			ClienteID:             cust.ID,
			DestinazioneCaricoID:  &carico.ID,
			DestinazioneScaricoID: &scarico.ID,
			DataRitiro:            ritiro.Format("2006-01-02"), OraRitiroDa: pick(oraRitiroDa), OraRitiroA: pick(oraRitiroA),
			DataConsegna: consegna.Format("2006-01-02"), OraConsegnaDa: pick(oraConsegnaDa), OraConsegnaA: pick(oraConsegnaA),
			Tariffa: pick(tariffe), TipoTariffa: pick(tipoTariffe), Tipologia: pick(tipologie),
			CategoriaTrasporto: pick(categorieTrasporto), RifOrdineCliente: rifOrdineCliente,
			AndataRitorno: randBelow(5) == 0,
			Items: []models.OrderItem{
				{ID: uuid.New(), ProdottoID: prod.ID, Quantita: 1, Peso: float64(randRange(8000, 25000))},
			},
			ServiziAccessori: []byte("[]"), CostiAccessori: []byte("[]"), Stato: models.OrderStato(stato),
		}

		if stato == "PIANIFICATO" || stato == "VIAGGIO" || stato == "CHIUSO" {
			m := pick(motrici)
			d := pick(drivers)
			order.TargaMotrice = m.Targa
			order.AutistaID = &d.ID
			// Punto di partenza: sempre presente su un ordine assegnato (come nel
			// form di assegnazione reale). Punto di lavaggio dopo lo scarico:
			// solo su una parte degli ordini, è opzionale anche a mano.
			g := pick(garages)
			order.GarageID = &g.ID
			if randBelow(2) == 0 {
				w := pick(washStations)
				order.WashStationID = &w.ID
			}
			if randBelow(3) == 0 {
				c := pick(carriers)
				order.VettoreID = &c.ID
			}
		}

		orders = append(orders, order)
	}
	if err := db.Create(&orders).Error; err != nil {
		return fmt.Errorf("ordini: %w", err)
	}
	fmt.Printf("✓ %d ordini\n", len(orders))

	// ─────────── ORDINI TEST: PERCORSI CORTI (<100KM), DA PIANIFICARE ───────────
	// Coppie carico/scarico reali (dall'export Visirun) a distanza nota <100km —
	// per testare le 3 alternative di percorso in AssignOrderForm: ORS rifiuta
	// alternative_routes oltre i 100km (vedi geo.GetRoadRouteAlternatives), e la
	// grande maggioranza degli ordini "normali" sopra è long-haul internazionale,
	// quindi non le esercita. Tutti PIANIFICABILE: pronti per essere assegnati a
	// mano dal tester, non assegnati automaticamente come gli ordini sopra.
	shortRoutePairs := []struct{ carico, scarico string }{
		{"3 B Latte Brignano Gera d'Adda", "Barry Verbania"},
		{"A&A CHOCOLATERIE LOKEREN", "REFRESCO FRANCE S.A. LE QUESNOY"},
		{"A&A Fratelli Parodi Campomorone", "FATTORIE OSELLA"},
		{"A-ware Dairy Trade B. V. Saturnus 21 8448 CC Heerenveen", "Bunge Loders Croklaan Wormerveer"},
		{"AarhusKarlshamn B.V", "DOHLER OOSTERHOUT"},
		{"Acetaia Monari Federzoni 1912 Solara", "Greenoleo Cremona"},
		{"ADEA BUSTO ARSIZIO", "OLFOOD BORGO SAN GIACOMO"},
		{"ADM Europoort Rotterdam", "PBI FRUIT JUICE COMPANY N.V. ZEEBRUGGE"},
		{"AGRICOLA ALIMENTARE-ITALIANA", "Innospec Castiglione delle Stiviere"},
		{"AGRICOLA GREINS ARRE (PD)", "Philip Morris Crespellano"},
		{"APHA TRADING CARBONARA SCRIVIA", "Balocco Fossano"},
		{"AZ.AGR.TROT. EREDE ROSSI SILVIO CASSOLNOVO", "Cavanna Olii Casella"},
		{"Azienda Agricola Ruffia", "BOTALLA FORMAGGI Mongrando"},
		{"Azienda Vinicola Carassanese", "CANTINA SOCIALE DI ARI (CH)"},
		{"Balconi Nerviano", "La Suissa Arquata Scrivia"},
		{"Barilla Castiglione delle Stiviere", "CEREAL DOCKS CAMISANO"},
		{"Barry Callebaut Wieze Lebbeke", "Delicia Tilburg"},
		{"BELGOMILK SCHOTEN", "Cargill Izegem"},
		{"Bimbo QSR Bomporto", "PESA PER WALCOR"},
		{"Biochem - Lohne", "MOLKEREI ELSDORF-ROTENBURG eG"},
		{"BIOLICA SAN GENESIO (PV)", "Ferrero Genova"},
		{"Biscottificio Baroni Albaredo d' Adige", "Montenegro SPA San Lazzaro di Savena"},
		{"Biscottificio Gandola Rudiano (BS)", "Cantine Riunite - Campegine"},
		{"Borgo Imperiale Motta Baluffi", "PELLEGRINI SPA RONCA'"},
		{"C.T. Logistics Serra Riccò", "Cleys Ozzero"},
	}
	destByNome := make(map[string]*models.Destination, len(destinations))
	for i := range destinations {
		destByNome[destinations[i].Nome] = &destinations[i]
	}
	shortRouteOrders := make([]models.Order, 0, len(shortRoutePairs))
	for i, pair := range shortRoutePairs {
		carico, okC := destByNome[pair.carico]
		scarico, okS := destByNome[pair.scarico]
		if !okC || !okS {
			continue
		}
		cust := pick(customers)
		prod := pick(products)
		ritiro := today.AddDate(0, 0, randRange(-5, 20))
		consegna := ritiro.AddDate(0, 0, randRange(1, 2))
		shortRouteOrders = append(shortRouteOrders, models.Order{
			ID: uuid.New(), Progressivo: fmt.Sprintf("%s/T%03d", today.Format("06"), i+1),
			ClienteID:             cust.ID,
			DestinazioneCaricoID:  &carico.ID,
			DestinazioneScaricoID: &scarico.ID,
			DataRitiro:            ritiro.Format("2006-01-02"), OraRitiroDa: pick(oraRitiroDa), OraRitiroA: pick(oraRitiroA),
			DataConsegna: consegna.Format("2006-01-02"), OraConsegnaDa: pick(oraConsegnaDa), OraConsegnaA: pick(oraConsegnaA),
			Tariffa: pick(tariffe), TipoTariffa: pick(tipoTariffe), Tipologia: "nazionale",
			CategoriaTrasporto: pick(categorieTrasporto),
			Items: []models.OrderItem{
				{ID: uuid.New(), ProdottoID: prod.ID, Quantita: 1, Peso: float64(randRange(8000, 25000))},
			},
			ServiziAccessori: []byte("[]"), CostiAccessori: []byte("[]"), Stato: models.OrderStatoPianificabile,
		})
	}
	if len(shortRouteOrders) > 0 {
		if err := db.Create(&shortRouteOrders).Error; err != nil {
			return fmt.Errorf("ordini test percorsi corti: %w", err)
		}
	}
	fmt.Printf("✓ %d ordini test percorsi corti (<100km, PIANIFICABILE)\n", len(shortRouteOrders))

	// ─────────────────────────── VIAGGI ───────────────────────────
	var viaggioOrders []models.Order
	for _, o := range orders {
		if o.Stato == "VIAGGIO" {
			viaggioOrders = append(viaggioOrders, o)
		}
	}

	var trips []models.Trip
	tripCount := min(10, len(viaggioOrders))
	for j := 0; j < tripCount; j++ {
		tripOrds := sliceClamp(viaggioOrders, j*2, j*2+randRange(1, 3))
		if len(tripOrds) == 0 {
			continue
		}
		m := pick(motrici)
		d := pick(drivers)
		trailer := pick(trailers)
		ordiniIds := make([]string, len(tripOrds))
		for i, o := range tripOrds {
			ordiniIds[i] = o.ID.String()
		}
		ordiniIdsJSON, err := json.Marshal(ordiniIds)
		if err != nil {
			return fmt.Errorf("trip ordini_ids: %w", err)
		}

		trip := models.Trip{
			ID: uuid.New(), OrdiniIds: ordiniIdsJSON,
			TargaMotrice: m.Targa, TargaRimorchio: trailer.Targa,
			AutistaID: &d.ID,
			GarageID:  &garages[0].ID,
			KmTotali:  float64(randRange(300, 2500)), CostoStimato: float64(randRange(800, 4000)),
			Stato: "IN_CORSO", DataPartenza: tripOrds[0].DataRitiro, DataArrivo: tripOrds[0].DataConsegna,
		}
		trips = append(trips, trip)

		for _, o := range tripOrds {
			db.Model(&models.Order{}).Where("id = ?", o.ID).Updates(map[string]interface{}{
				"viaggio_id": trip.ID, "targa_motrice": m.Targa, "autista_id": d.ID,
			})
		}
	}
	if len(trips) > 0 {
		if err := db.Create(&trips).Error; err != nil {
			return fmt.Errorf("viaggi: %w", err)
		}
	}
	fmt.Printf("✓ %d viaggi\n", len(trips))

	// ─────────────────────────── FATTURE ───────────────────────────
	// "Fatturato" non è un valore di Stato dell'ordine: un ordine CHIUSO che
	// è stato fatturato resta CHIUSO, e porta semplicemente un fattura_id
	// valorizzato (esattamente come fa InvoiceService.Finalize in
	// produzione). Dividiamo gli ordini CHIUSO in due gruppi: la prima metà
	// simula fatture già finalizzate (DEFINITIVA + fattura_id stampato sugli
	// ordini), la seconda genera fatture ancora in bozza (PROFORMA, nessun
	// fattura_id sugli ordini).
	var chiusoOrders []models.Order
	for _, o := range orders {
		if o.Stato == "CHIUSO" {
			chiusoOrders = append(chiusoOrders, o)
		}
	}
	mid := len(chiusoOrders) / 2
	definitivaPool, proformaPool := chiusoOrders[:mid], chiusoOrders[mid:]

	destByID := make(map[uuid.UUID]models.Destination, len(destinations))
	for _, d := range destinations {
		destByID[d.ID] = d
	}
	prodByID := make(map[uuid.UUID]models.Product, len(products))
	for _, p := range products {
		prodByID[p.ID] = p
	}

	var invoices []models.Invoice
	for k := 0; k < min(4, len(definitivaPool)); k++ {
		invOrders := sliceClamp(definitivaPool, k*3, k*3+randRange(2, 4))
		if len(invOrders) == 0 {
			continue
		}
		righe, totale := buildInvoiceLines(invOrders, destByID, prodByID)
		inv := models.Invoice{
			ID: uuid.New(), Numero: fmt.Sprintf("O/F-25/%04d", k+1),
			ClienteID:           invOrders[0].ClienteID,
			DataFattura:         fmt.Sprintf("2025-%02d-%02d", randRange(7, 11), randRange(1, 28)),
			DataScadenza:        fmt.Sprintf("2025-%02d-%02d", randRange(9, 12), randRange(1, 28)),
			CondizioniPagamento: "BB 30gg DF FM", Righe: righe, CostiAccessori: []byte("[]"),
			TotaleImponibile: totale, Totale: totale, Stato: "DEFINITIVA", Tipo: "ordine",
		}
		invoices = append(invoices, inv)
		ordIDs := make([]string, len(invOrders))
		for i, o := range invOrders {
			ordIDs[i] = o.ID.String()
		}
		db.Model(&models.Order{}).Where("id IN ?", ordIDs).Update("fattura_id", inv.ID)
	}
	for k := 0; k < min(4, len(proformaPool)); k++ {
		invOrders := sliceClamp(proformaPool, k*2, k*2+randRange(1, 3))
		if len(invOrders) == 0 {
			continue
		}
		righe, totale := buildInvoiceLines(invOrders, destByID, prodByID)
		invoices = append(invoices, models.Invoice{
			ID: uuid.New(), Numero: fmt.Sprintf("O/F-25/%04d", k+5),
			ClienteID:   invOrders[0].ClienteID,
			DataFattura: fmt.Sprintf("2025-%02d-%02d", randRange(9, 12), randRange(1, 28)),
			Righe:       righe, CostiAccessori: []byte("[]"),
			TotaleImponibile: totale, Totale: totale, Stato: "PROFORMA", Tipo: "ordine",
		})
	}
	if len(invoices) > 0 {
		if err := db.Create(&invoices).Error; err != nil {
			return fmt.Errorf("fatture: %w", err)
		}
	}
	fmt.Printf("✓ %d fatture\n", len(invoices))

	// ─────────────────────────── INDISPONIBILITÀ AUTISTI ───────────────────────────
	iso := func(d time.Time) string { return d.Format("2006-01-02") }
	unavails := []models.DriverUnavailability{
		{ID: uuid.New(), AutistaID: drivers[2].ID, AutistaNome: drivers[2].Nome + " " + drivers[2].Cognome, DataDa: iso(today.AddDate(0, 0, 14)), DataA: iso(today.AddDate(0, 0, 28)), Motivo: "ferie", Note: "Ferie estive"},
		{ID: uuid.New(), AutistaID: drivers[5].ID, AutistaNome: drivers[5].Nome + " " + drivers[5].Cognome, DataDa: iso(today.AddDate(0, 0, -5)), DataA: iso(today.AddDate(0, 0, -1)), Motivo: "malattia"},
		{ID: uuid.New(), AutistaID: drivers[7].ID, AutistaNome: drivers[7].Nome + " " + drivers[7].Cognome, DataDa: iso(today.AddDate(0, 0, 3)), DataA: iso(today.AddDate(0, 0, 5)), Motivo: "permesso", Note: "Permesso familiare"},
	}
	if err := db.Create(&unavails).Error; err != nil {
		return fmt.Errorf("indisponibilita' autisti: %w", err)
	}
	fmt.Printf("✓ %d indisponibilita' autisti\n", len(unavails))

	fmt.Println()
	fmt.Println("═══════════════════════════════════════")
	fmt.Println("  SEED COMPLETATO")
	fmt.Println("═══════════════════════════════════════")
	fmt.Println("  Salva queste credenziali in un password manager ORA:")
	fmt.Println("  (non verranno piu' mostrate)")
	fmt.Println()
	fmt.Printf("  admin            %-32s %s\n", adminEmail, adminPassword)
	fmt.Printf("  planner          m.rossi@feccia.it                %s\n", plannerPassword)
	fmt.Printf("  operatore        l.bianchi@feccia.it              %s\n", operatorePassword)
	fmt.Printf("  amministrazione  g.ferrari@feccia.it              %s\n", amminPassword)
	fmt.Println("═══════════════════════════════════════")
	return nil
}

// buildInvoiceLines snapshots destinazione/prodotto names onto the invoice
// line at billing time (InvoiceLine.Descrizione/Prodotto are historical
// document text, not live FK references — see models/invoice.go) — destByID/
// prodByID resolve the in-memory orders' FK ids since Order no longer
// carries a denormalized name column to read directly.
func buildInvoiceLines(orders []models.Order, destByID map[uuid.UUID]models.Destination, prodByID map[uuid.UUID]models.Product) ([]models.InvoiceLine, float64) {
	righe := make([]models.InvoiceLine, len(orders))
	var totale float64
	for i, o := range orders {
		prodotto, peso := "", 0.0
		if len(o.Items) > 0 {
			if p, ok := prodByID[o.Items[0].ProdottoID]; ok {
				prodotto = p.Descrizione
			}
			peso = o.Items[0].Peso
		}
		caricoNome, scaricoNome := "", ""
		if o.DestinazioneCaricoID != nil {
			caricoNome = destByID[*o.DestinazioneCaricoID].Nome
		}
		if o.DestinazioneScaricoID != nil {
			scaricoNome = destByID[*o.DestinazioneScaricoID].Nome
		}
		righe[i] = models.InvoiceLine{
			ID: uuid.New(), OrdineID: o.ID.String(),
			Descrizione: caricoNome + " - " + scaricoNome,
			Prodotto:    prodotto, Peso: peso, Quantita: 1, Tariffa: o.Tariffa, Totale: o.Tariffa, IvaCodice: "N8",
		}
		totale += o.Tariffa
	}
	return righe, totale
}

func mustUser(login, name, role, password string) models.User {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return models.User{Login: login, Name: name, Role: role, Active: true, PasswordHash: string(hash)}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// randomToken mirrors Python's secrets.token_urlsafe(16): 16 random bytes,
// base64 URL-safe, no padding.
func randomToken() string {
	b := make([]byte, 16)
	if _, err := crand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// randBelow mirrors secrets.randbelow(n) — not security-critical here (only
// used to pick demo data), so math/rand/v2's auto-seeded global source is
// used instead of crypto/rand.
func randBelow(n int) int {
	return rand.IntN(n)
}

// randRange mirrors Python's `a + secrets.randbelow(b - a + 1)`: inclusive [a, b].
func randRange(a, b int) int {
	return a + rand.IntN(b-a+1)
}

func geoPtr(v float64) *float64 {
	return &v
}

func pick[T any](s []T) T {
	return s[rand.IntN(len(s))]
}

// weighted expands a {value: weight} map into a flat slice for pick(), same
// effect as Python's `["a"]*3 + ["b"]*1` weighted-list idiom. Map iteration
// order doesn't matter since every value is repeated `weight` times either way.
func weighted(counts map[string]int) []string {
	out := make([]string, 0, len(counts))
	for value, count := range counts {
		for i := 0; i < count; i++ {
			out = append(out, value)
		}
	}
	return out
}

// sliceClamp mirrors Python's forgiving slice semantics (s[start:end] never
// panics, just returns less than requested at the boundaries).
func sliceClamp(s []models.Order, start, end int) []models.Order {
	if start >= len(s) || start < 0 {
		return nil
	}
	if end > len(s) {
		end = len(s)
	}
	if end <= start {
		return nil
	}
	return s[start:end]
}
