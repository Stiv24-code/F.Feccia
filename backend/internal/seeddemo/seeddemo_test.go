package seeddemo

import (
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/database"
	"fratelli-feccia/pkg/utils"
)

// seededDB runs the whole demo seed against an in-memory SQLite, the same
// setup the service tests use. database.Migrate is dialect-aware (the
// Postgres-only dedup index is skipped here), so this exercises the seed as
// written without needing a real Postgres.
func seededDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("apertura db di test: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := Seed(db); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return db
}

func inbound(t *testing.T, db *gorm.DB) []models.InboundOrder {
	t.Helper()
	var out []models.InboundOrder
	if err := db.Find(&out).Error; err != nil {
		t.Fatalf("lettura ordini in arrivo: %v", err)
	}
	return out
}

// TestSeedInboundMatrix is the regression guard for the reason the matrix
// exists at all: before, canale e stato erano estratti a caso e un seed
// sfortunato poteva non produrre mai un draft da portale o uno in modifica.
func TestSeedInboundMatrix(t *testing.T) {
	db := seededDB(t)
	orders := inbound(t, db)

	// Chiave di stato come la legge la dashboard: status da solo non basta,
	// il flag portal sdoppia "pending" e order_id sdoppia "accepted".
	key := func(o models.InboundOrder) string {
		state := string(o.Status)
		if o.Portal {
			state += "+portal"
		}
		if o.OrderID != nil {
			state += "+converted"
		}
		return o.Source + "/" + state
	}
	seen := map[string]int{}
	for _, o := range orders {
		seen[key(o)]++
	}

	want := []string{
		"pdf/pending", "pdf/pending+portal", "pdf/accepted", "pdf/accepted+converted", "pdf/modify",
		"mail/pending", "mail/pending+portal", "mail/accepted", "mail/accepted+converted", "mail/modify",
		// Il canale portale non ha il caso "+portal": quel flag significa
		// "da confermare sul portale del cliente", non sul nostro.
		"portal/pending", "portal/accepted", "portal/accepted+converted", "portal/modify",
	}
	for _, w := range want {
		if seen[w] < 3 {
			t.Errorf("combinazione %s: %d righe, attese 3 (una per cliente del canale)", w, seen[w])
		}
	}
	if got, exp := len(seen), len(want); got != exp {
		t.Errorf("combinazioni canale/stato presenti: %d, attese %d (%v)", got, exp, seen)
	}
	if seen["portal/pending+portal"] != 0 {
		t.Errorf("il canale portale non deve avere il flag portal: %d righe", seen["portal/pending+portal"])
	}
}

// TestSeedInboundChannelPayloads pins what each channel is allowed to carry:
// è la differenza che rende utile la matrice, e quella su cui poggia la
// regola di sicurezza di Convert (cliente_id solo da fonte attendibile).
func TestSeedInboundChannelPayloads(t *testing.T) {
	db := seededDB(t)

	var templates []models.PdfTemplate
	if err := db.Find(&templates).Error; err != nil {
		t.Fatalf("lettura template: %v", err)
	}
	tplIDs := map[string]bool{}
	for _, tpl := range templates {
		tplIDs[tpl.ID.String()] = true
	}
	if len(templates) < 3 {
		t.Fatalf("template seminati: %d, attesi almeno 3 (uno per cliente del canale pdf)", len(templates))
	}
	tplUsed := map[string]int{}

	for _, o := range inbound(t, db) {
		switch o.Source {
		case models.InboundOrderSourcePDF:
			if o.TemplateID == nil {
				t.Errorf("%s: draft pdf senza template_id", o.Ref)
				continue
			}
			if !tplIDs[o.TemplateID.String()] {
				t.Errorf("%s: template_id %s non corrisponde a nessun template seminato", o.Ref, o.TemplateID)
			}
			tplUsed[o.TemplateID.String()]++
			if o.ClienteID != nil {
				t.Errorf("%s: un draft pdf non deve portare cliente_id (il mittente non e' attendibile)", o.Ref)
			}
		case models.InboundOrderSourceMail:
			if o.TemplateID != nil {
				t.Errorf("%s: draft mail con template_id", o.Ref)
			}
			if o.ClienteID != nil {
				t.Errorf("%s: un draft mail non deve portare cliente_id (il mittente non e' attendibile)", o.Ref)
			}
		case models.InboundOrderSourcePortal:
			if o.ClienteID == nil {
				t.Errorf("%s: draft da portale senza cliente_id", o.Ref)
			}
			if o.DestinazioneCaricoID == nil || o.DestinazioneScaricoID == nil {
				t.Errorf("%s: draft da portale senza destinazioni strutturate", o.Ref)
			}
			if o.CommittenteID == nil {
				t.Errorf("%s: draft da portale senza committente", o.Ref)
			}
			if o.TariffaProposta <= 0 {
				t.Errorf("%s: draft da portale senza tariffa proposta", o.Ref)
			}
			if o.OraRitiroDa == "" || o.OraConsegnaDa == "" {
				t.Errorf("%s: draft da portale senza orari", o.Ref)
			}
		default:
			t.Errorf("%s: canale inatteso %q", o.Ref, o.Source)
		}
	}

	for id, n := range tplUsed {
		if n == 0 {
			t.Errorf("template %s senza richieste importate", id)
		}
	}
	if len(tplUsed) != len(templates) {
		t.Errorf("template con richieste dietro: %d su %d", len(tplUsed), len(templates))
	}
}

// TestSeedInboundConvertedLinks: order_id e' cio' che rende Convert
// idempotente, quindi un draft "convertito" deve puntare a un ordine che
// esiste davvero — altrimenti sarebbe uno stato che il codice non puo'
// produrre.
func TestSeedInboundConvertedLinks(t *testing.T) {
	db := seededDB(t)

	converted := 0
	for _, o := range inbound(t, db) {
		if o.OrderID == nil {
			continue
		}
		converted++
		if o.Status != models.InboundOrderStatusAccepted {
			t.Errorf("%s: draft convertito con stato %q", o.Ref, o.Status)
		}
		var ord models.Order
		if err := db.First(&ord, "id = ?", *o.OrderID).Error; err != nil {
			t.Errorf("%s: order_id %s non risolve a un ordine: %v", o.Ref, o.OrderID, err)
			continue
		}
		if ord.RifOrdineCliente != o.Ref {
			t.Errorf("%s: l'ordine collegato porta rif %q", o.Ref, ord.RifOrdineCliente)
		}
		var items int64
		db.Model(&models.OrderItem{}).Where("order_id = ?", ord.ID).Count(&items)
		if items == 0 {
			t.Errorf("%s: l'ordine collegato non ha righe prodotto", o.Ref)
		}
	}
	if converted == 0 {
		t.Fatal("nessun draft convertito seminato")
	}
}

// TestSeedInboundRefsUnique replicates inbound_orders_ref_client_key, che su
// SQLite non viene creato: senza questo il seed potrebbe passare in test e
// fallire all'insert su Postgres.
func TestSeedInboundRefsUnique(t *testing.T) {
	db := seededDB(t)

	seen := map[string]string{}
	for _, o := range inbound(t, db) {
		k := fmt.Sprintf("%s|%s", o.Ref, o.Client)
		if prev, dup := seen[k]; dup {
			t.Errorf("chiave (ref, client) duplicata: %q (id %s e %s)", k, prev, o.ID)
		}
		seen[k] = o.ID.String()
	}
}

// TestSeedPortalLoginsSeeCoda: il canale portale serve a poter fare login e
// vedere la propria coda, quindi ogni account cliente deve avere richieste
// non ancora convertite (ListForClient filtra su order_id IS NULL).
func TestSeedPortalLoginsSeeCoda(t *testing.T) {
	db := seededDB(t)

	var clients []models.User
	if err := db.Where("role = ?", utils.RoleCliente).Find(&clients).Error; err != nil {
		t.Fatalf("lettura utenti cliente: %v", err)
	}
	if len(clients) < 3 {
		t.Fatalf("utenti cliente seminati: %d, attesi almeno 3", len(clients))
	}
	for _, u := range clients {
		if u.CustomerID == nil {
			t.Errorf("utente cliente %s senza customer_id", u.Login)
			continue
		}
		var pending int64
		db.Model(&models.InboundOrder{}).
			Where("cliente_id = ? AND order_id IS NULL", *u.CustomerID).
			Count(&pending)
		if pending == 0 {
			t.Errorf("utente cliente %s: nessuna richiesta in coda sul portale", u.Login)
		}
	}
}
