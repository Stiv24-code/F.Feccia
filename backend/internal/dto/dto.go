package dto

import (
	"time"

	"github.com/google/uuid"
)

// LoginRequest uses "email" as the wire field name to match the frontend's
// existing contract (built against the Python backend, which authenticates
// by email) even though the underlying User model's column is just "login" —
// a generic string identifier, not validated as an email format.
type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// AuthUserResponse mirrors Python's UserOut — used both in the login/refresh
// response and as the response shape for /auth/register, GET /admin/users
// and PATCH /admin/users/{id}. ProfileID is always nil — the RBAC "profiles"
// concept (backend/routers/admin.py) was deliberately not ported (see plan),
// so there is nothing to populate it with.
type AuthUserResponse struct {
	ID        int64   `json:"id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	Role      string  `json:"role"`
	ProfileID *string `json:"profile_id"`
	// CustomerID is only non-nil for RoleCliente accounts (self-registered
	// via /auth/register-cliente) — the Customer/anagrafica they're scoped
	// to. Distinct from the legacy unused ProfileID above.
	CustomerID *string `json:"customer_id"`
	Active     bool    `json:"active"`
}

// RegisterRequest mirrors Python's UserCreate — the password policy
// (min 12 chars) matches backend/models.py's UserCreate.password field
// exactly, since this is admin-facing user provisioning, not self-signup.
// CustomerID is required exactly when Role == "cliente": an admin-provisioned
// client-portal account is scoped to an EXISTING Customer/anagrafica the
// admin picks (unlike self-registration, which has no existing Customer to
// pick from and creates one instead) — checked in AuthService.Register, not
// here (validator "required_if" can't express "one of several role values").
type RegisterRequest struct {
	Email      string  `json:"email" validate:"required,email"`
	Name       string  `json:"name" validate:"required,min=1"`
	Password   string  `json:"password" validate:"required,min=12"`
	Role       string  `json:"role" validate:"required,oneof=admin amministrazione planner operatore cliente"`
	CustomerID *string `json:"customer_id,omitempty"`
}

// ClientRegisterRequest is the public, unauthenticated self-registration
// form (POST /auth/register-cliente): creates a Customer (anagrafica) and a
// RoleCliente User atomically (see AuthService.RegisterClient) — unlike
// RegisterRequest, Role is never accepted from the caller.
type ClientRegisterRequest struct {
	RagioneSociale string `json:"ragione_sociale" validate:"required"`
	Indirizzo      string `json:"indirizzo"`
	Citta          string `json:"citta"`
	Cap            string `json:"cap"`
	Provincia      string `json:"provincia"`
	// Lat/Lng are optional, filled in from the Indirizzo geocoding search when
	// the address matched — never required, so a signup never fails just
	// because the address wasn't found on the map.
	Lat           *float64 `json:"lat"`
	Lng           *float64 `json:"lng"`
	PartitaIva    string   `json:"partita_iva"`
	CodiceFiscale string   `json:"codice_fiscale"`
	Telefono      string   `json:"telefono"`
	Name          string   `json:"name" validate:"required,min=1"`
	Email         string   `json:"email" validate:"required,email"`
	Password      string   `json:"password" validate:"required,min=12"`
}

// PatchUserRequest mirrors Python's UserUpdate (PATCH /admin/users/{id}):
// all fields optional, only provided ones are applied. ProfileID is accepted
// on the wire (the frontend always sends it) but any non-empty value is
// rejected — there is no profiles table to validate it against yet.
type PatchUserRequest struct {
	Name      *string `json:"name"`
	ProfileID *string `json:"profile_id"`
	Active    *bool   `json:"active"`
}

// LoginResult is returned by both /auth/login and /auth/refresh. RefreshToken
// is deliberately excluded from the JSON body (json:"-") — it only ever
// travels as an httpOnly cookie, mirroring backend/services.py's
// build_login_response (access token in body, refresh token in cookie).
type LoginResult struct {
	AccessToken  string           `json:"access_token"`
	RefreshToken string           `json:"-"`
	RefreshTTL   time.Duration    `json:"-"`
	TokenType    string           `json:"token_type"`
	ExpiresIn    int64            `json:"expires_in"`
	User         AuthUserResponse `json:"user"`
}

// RegisterClientResult is returned by POST /auth/register-cliente instead of
// LoginResult whenever a confirmation email was actually sent — Verified is
// false and there is no token, the caller must complete VerifyEmailRequest
// first. When SMTP isn't configured (no way to verify anything) the endpoint
// falls back to the old immediate-login behaviour and still returns a plain
// LoginResult, so this type's absence from a 200 response is meaningful too.
type RegisterClientResult struct {
	Verified bool   `json:"verified"`
	Message  string `json:"message"`
	Email    string `json:"email"`
}

// VerifyEmailRequest confirms a registration link's token (POST
// /auth/verify-email) — on success the account behaves exactly like a fresh
// Login, returning a normal LoginResult.
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

// ResendVerificationRequest asks for a fresh confirmation link (POST
// /auth/resend-verification) without re-submitting the whole
// ClientRegisterRequest form again.
type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type CreateUserRequest struct {
	Login    string `json:"login" validate:"required,min=3,max=150"`
	Name     string `json:"name"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=admin amministrazione planner operatore"`
}

type UpdateUserRequest struct {
	Login    string  `json:"login" validate:"required,min=3,max=150"`
	Name     string  `json:"name"`
	Password *string `json:"password"`
	// "cliente" is accepted here only so re-saving a cliente account's
	// name/active status (role unchanged) doesn't fail validation — the
	// admin UI never offers switching a role TO cliente via this edit path
	// (that always goes through /auth/register instead, which also asks for
	// the Customer to link).
	Role string `json:"role" validate:"required,oneof=admin amministrazione planner operatore cliente"`
}

type UserResponse struct {
	ID        int64     `json:"id"`
	Login     string    `json:"login"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CustomerRequest is used for both create and update — Python's customers.py
// accepts the same CustomerCreate schema for POST and PUT.
type CustomerRequest struct {
	RagioneSociale string `json:"ragione_sociale" validate:"required"`
	Indirizzo      string `json:"indirizzo"`
	Citta          string `json:"citta"`
	Cap            string `json:"cap"`
	Provincia      string `json:"provincia"`
	Nazione        string `json:"nazione"`
	// Lat/Lng optional — see models.Customer for why this differs from
	// Destination/Garage/WashStation's mandatory Posizione.
	Lat                 *float64 `json:"lat"`
	Lng                 *float64 `json:"lng"`
	PartitaIva          string   `json:"partita_iva"`
	CodiceFiscale       string   `json:"codice_fiscale"`
	Telefono            string   `json:"telefono"`
	Email               string   `json:"email"`
	Pec                 string   `json:"pec"`
	CondizioniPagamento string   `json:"condizioni_pagamento"`
	Note                string   `json:"note"`
	RichiedeRifOrdine   bool     `json:"richiede_rif_ordine"`
}

type DestinationRequest struct {
	Nome           string   `json:"nome" validate:"required"`
	Indirizzo      string   `json:"indirizzo"`
	Citta          string   `json:"citta"`
	Cap            string   `json:"cap"`
	Provincia      string   `json:"provincia"`
	Nazione        string   `json:"nazione"`
	Lat            *float64 `json:"lat" validate:"required"`
	Lng            *float64 `json:"lng" validate:"required"`
	VincoliScarico string   `json:"vincoli_scarico"`
	Note           string   `json:"note"`
	Active         bool     `json:"active"`
}

type DestinationResponse struct {
	ID             uuid.UUID `json:"id"`
	Nome           string    `json:"nome"`
	Indirizzo      string    `json:"indirizzo"`
	Citta          string    `json:"citta"`
	Cap            string    `json:"cap"`
	Provincia      string    `json:"provincia"`
	Nazione        string    `json:"nazione"`
	Lat            *float64  `json:"lat"`
	Lng            *float64  `json:"lng"`
	VincoliScarico string    `json:"vincoli_scarico"`
	Note           string    `json:"note"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CarrierRequest struct {
	RagioneSociale string `json:"ragione_sociale" validate:"required"`
	PartitaIva     string `json:"partita_iva"`
	Indirizzo      string `json:"indirizzo"`
	Citta          string `json:"citta"`
	Telefono       string `json:"telefono"`
	Email          string `json:"email"`
	Note           string `json:"note"`
}

type CarrierResponse struct {
	ID             uuid.UUID `json:"id"`
	RagioneSociale string    `json:"ragione_sociale"`
	PartitaIva     string    `json:"partita_iva"`
	Indirizzo      string    `json:"indirizzo"`
	Citta          string    `json:"citta"`
	Telefono       string    `json:"telefono"`
	Email          string    `json:"email"`
	Note           string    `json:"note"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type GarageRequest struct {
	Nome      string   `json:"nome" validate:"required"`
	Indirizzo string   `json:"indirizzo"`
	Citta     string   `json:"citta"`
	Lat       *float64 `json:"lat" validate:"required"`
	Lng       *float64 `json:"lng" validate:"required"`
	Note      string   `json:"note"`
	Active    bool     `json:"active"`
}

type GarageResponse struct {
	ID        uuid.UUID `json:"id"`
	Nome      string    `json:"nome"`
	Indirizzo string    `json:"indirizzo"`
	Citta     string    `json:"citta"`
	Lat       *float64  `json:"lat"`
	Lng       *float64  `json:"lng"`
	Note      string    `json:"note"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WashStationRequest struct {
	Nome      string   `json:"nome" validate:"required"`
	Tipo      string   `json:"tipo"`
	Indirizzo string   `json:"indirizzo"`
	Citta     string   `json:"citta"`
	Lat       *float64 `json:"lat" validate:"required"`
	Lng       *float64 `json:"lng" validate:"required"`
	OrarioDa  string   `json:"orario_da"`
	OrarioA   string   `json:"orario_a"`
	Note      string   `json:"note"`
	Active    bool     `json:"active"`
}

type WashStationResponse struct {
	ID        uuid.UUID `json:"id"`
	Nome      string    `json:"nome"`
	Tipo      string    `json:"tipo"`
	Indirizzo string    `json:"indirizzo"`
	Citta     string    `json:"citta"`
	Lat       *float64  `json:"lat"`
	Lng       *float64  `json:"lng"`
	OrarioDa  string    `json:"orario_da"`
	OrarioA   string    `json:"orario_a"`
	Note      string    `json:"note"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DriverRequest struct {
	Nome            string   `json:"nome" validate:"required"`
	Cognome         string   `json:"cognome" validate:"required"`
	CodiceFiscale   string   `json:"codice_fiscale"`
	Patente         []string `json:"patente" enums:"AM,A1,A2,A,B1,B,BE,C1,C1E,C,CE,D1,D1E,D,DE,CQC,ADR" validate:"omitempty,dive,oneof=AM A1 A2 A B1 B BE C1 C1E C CE D1 D1E D DE CQC ADR"`
	ScadenzaPatente *string  `json:"scadenza_patente"`
	Telefono        string   `json:"telefono"`
	Email           string   `json:"email"`
	Note            string   `json:"note"`
}

type DriverResponse struct {
	ID              uuid.UUID `json:"id"`
	Nome            string    `json:"nome"`
	Cognome         string    `json:"cognome"`
	CodiceFiscale   string    `json:"codice_fiscale"`
	Patente         []string  `json:"patente" enums:"AM,A1,A2,A,B1,B,BE,C1,C1E,C,CE,D1,D1E,D,DE,CQC,ADR"`
	ScadenzaPatente *string   `json:"scadenza_patente"`
	Telefono        string    `json:"telefono"`
	Email           string    `json:"email"`
	Note            string    `json:"note"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	// ProssimeFerieDa/A are the next upcoming motivo=ferie unavailability
	// window (nil if none), computed on read only by DriverService.List —
	// not persisted, not populated by Create/Update.
	ProssimeFerieDa *string `json:"prossime_ferie_da"`
	ProssimeFerieA  *string `json:"prossime_ferie_a"`
}

type ProductRequest struct {
	Codice      string `json:"codice" validate:"required"`
	Descrizione string `json:"descrizione" validate:"required"`
	UnitaMisura string `json:"unita_misura"`
	Note        string `json:"note"`
}

type ProductResponse struct {
	ID          uuid.UUID `json:"id"`
	Codice      string    `json:"codice"`
	Descrizione string    `json:"descrizione"`
	UnitaMisura string    `json:"unita_misura"`
	Note        string    `json:"note"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type VehicleTypeRequest struct {
	Nome        string `json:"nome" validate:"required"`
	Descrizione string `json:"descrizione"`
}

type VehicleTypeResponse struct {
	ID          uuid.UUID `json:"id"`
	Nome        string    `json:"nome"`
	Descrizione string    `json:"descrizione"`
	Active      bool      `json:"active"`
}

type AccessoryCostRequest struct {
	Nome         string  `json:"nome" validate:"required"`
	Descrizione  string  `json:"descrizione"`
	CostoDefault float64 `json:"costo_default"`
}

type AccessoryCostResponse struct {
	ID           uuid.UUID `json:"id"`
	Nome         string    `json:"nome"`
	Descrizione  string    `json:"descrizione"`
	CostoDefault float64   `json:"costo_default"`
	Active       bool      `json:"active"`
}

type TransportCategoryRequest struct {
	Nome        string `json:"nome" validate:"required"`
	Descrizione string `json:"descrizione"`
}

type TransportCategoryResponse struct {
	ID          uuid.UUID `json:"id"`
	Nome        string    `json:"nome"`
	Descrizione string    `json:"descrizione"`
	Active      bool      `json:"active"`
}

type CountryRequest struct {
	Iso2   string `json:"iso2" validate:"required"`
	Iso3   string `json:"iso3"`
	Nome   string `json:"nome" validate:"required"`
	Eu     bool   `json:"eu"`
	Valuta string `json:"valuta"`
}

type CountryResponse struct {
	ID        uuid.UUID `json:"id"`
	Iso2      string    `json:"iso2"`
	Iso3      string    `json:"iso3"`
	Nome      string    `json:"nome"`
	Eu        bool      `json:"eu"`
	Valuta    string    `json:"valuta"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BankRequest struct {
	Nome       string `json:"nome" validate:"required"`
	BicSwift   string `json:"bic_swift"`
	IbanPrefix string `json:"iban_prefix"`
	Indirizzo  string `json:"indirizzo"`
	Citta      string `json:"citta"`
	Note       string `json:"note"`
}

type BankResponse struct {
	ID         uuid.UUID `json:"id"`
	Nome       string    `json:"nome"`
	BicSwift   string    `json:"bic_swift"`
	IbanPrefix string    `json:"iban_prefix"`
	Indirizzo  string    `json:"indirizzo"`
	Citta      string    `json:"citta"`
	Note       string    `json:"note"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type AccountingEntryRequest struct {
	Codice         string `json:"codice" validate:"required"`
	Descrizione    string `json:"descrizione" validate:"required"`
	Tipo           string `json:"tipo" validate:"omitempty,oneof=ricavo costo"`
	ContoContabile string `json:"conto_contabile"`
	IvaCodice      string `json:"iva_codice"`
}

type AccountingEntryResponse struct {
	ID             uuid.UUID `json:"id"`
	Codice         string    `json:"codice"`
	Descrizione    string    `json:"descrizione"`
	Tipo           string    `json:"tipo"`
	ContoContabile string    `json:"conto_contabile"`
	IvaCodice      string    `json:"iva_codice"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type DriverUnavailabilityRequest struct {
	AutistaID   uuid.UUID `json:"autista_id" validate:"required"`
	AutistaNome string    `json:"autista_nome"`
	DataDa      string    `json:"data_da" validate:"required"`
	DataA       string    `json:"data_a" validate:"required"`
	Motivo      string    `json:"motivo" enums:"ferie,malattia,permesso,altro" validate:"omitempty,oneof=ferie malattia permesso altro"`
	Note        string    `json:"note"`
}

type DriverUnavailabilityResponse struct {
	ID          uuid.UUID `json:"id"`
	AutistaID   uuid.UUID `json:"autista_id"`
	AutistaNome string    `json:"autista_nome"`
	DataDa      string    `json:"data_da"`
	DataA       string    `json:"data_a"`
	Motivo      string    `json:"motivo" enums:"ferie,malattia,permesso,altro"`
	Note        string    `json:"note"`
	CreatedAt   time.Time `json:"created_at"`
}

// OrderItemRequestDTO is the write-side shape of an order line — only the
// product reference id, the server resolves/validates the rest via Preload.
type OrderItemRequestDTO struct {
	ProdottoID string  `json:"prodotto_id"`
	Quantita   float64 `json:"quantita"`
	Peso       float64 `json:"peso"`
}

// OrderItemResponseDTO is the read-side shape — Prodotto is the full nested
// product (nil if the product reference is somehow unresolved), replacing
// the old denormalized ProdottoCodice/ProdottoDescrizione snapshot fields.
type OrderItemResponseDTO struct {
	Prodotto *ProductResponse `json:"prodotto"`
	Quantita float64          `json:"quantita"`
	Peso     float64          `json:"peso"`
}

// OrderRequest mirrors OrderCreate — used for both POST and PUT; state-machine
// fields (stato, motrice_id, semirimorchio_id, autista_id, vettore_id, viaggio_id,
// fattura_id, progressivo) are intentionally absent, exactly like the Python schema, so a
// PUT can never touch them. Reference fields are ids only — the server no
// longer stores a client-submitted denormalized name, it's always derived
// from the live associated row via Preload.
type OrderRequest struct {
	ClienteID string `json:"cliente_id" validate:"required"`
	// CommittenteID: parte ordinante se diversa dal cliente fatturato (vuoto
	// = coincide con Cliente).
	CommittenteID         string                   `json:"committente_id"`
	DestinazioneCaricoID  string                   `json:"destinazione_carico_id"`
	DestinazioneScaricoID string                   `json:"destinazione_scarico_id"`
	DataRitiro            string                   `json:"data_ritiro"`
	OraRitiroDa           string                   `json:"ora_ritiro_da"`
	OraRitiroA            string                   `json:"ora_ritiro_a"`
	DataConsegna          string                   `json:"data_consegna"`
	OraConsegnaDa         string                   `json:"ora_consegna_da"`
	OraConsegnaA          string                   `json:"ora_consegna_a"`
	Tariffa               float64                  `json:"tariffa"`
	TipoTariffa           string                   `json:"tipo_tariffa"`
	Tipologia             string                   `json:"tipologia"`
	CategoriaTrasporto    string                   `json:"categoria_trasporto"`
	RifOrdineCliente      string                   `json:"rif_ordine_cliente"`
	RifCarico             string                   `json:"rif_carico"`
	NoteCarico            string                   `json:"note_carico"`
	RifScarico            string                   `json:"rif_scarico"`
	NoteScarico           string                   `json:"note_scarico"`
	AndataRitorno         bool                     `json:"andata_ritorno"`
	Provvisorio           bool                     `json:"provvisorio"`
	Note                  string                   `json:"note"`
	Items                 []OrderItemRequestDTO    `json:"items"`
	ServiziAccessori      []string                 `json:"servizi_accessori"`
	CostiAccessori        []map[string]interface{} `json:"costi_accessori"`
}

// ClientInboundOrderRequest is the body of POST /me/inbound-orders — the
// same fields as OrderRequest (destinazioni/date/tariffa/note), plus the two
// free-text fields InboundOrder itself needs (product/kg) that OrderRequest
// has no equivalent for (real Order captures product+weight per line via
// Items/ProductID, but the client-portal form is a plain draft, not a
// priced order — matches how mail/PDF-sourced drafts already carry these as
// plain strings, see InboundOrder.Product/Kg).
type ClientInboundOrderRequest struct {
	OrderRequest
	Product string `json:"product"`
	Kg      int    `json:"kg"`
}

type OrderAssignRequest struct {
	GarageID        string `json:"garage_id"`
	MotriceID       string `json:"motrice_id"`
	SemirimorchioID string `json:"semirimorchio_id"`
	AutistaID       string `json:"autista_id"`
	VettoreID       string `json:"vettore_id"`
	WashStationID   string `json:"wash_station_id"`
	// RouteWaypoints: la sequenza scelta dal manager tra le alternative
	// proposte (POST /orders/{id}/route-alternatives). Opzionale — se
	// assente l'ordine viene comunque assegnato, semplicemente senza un
	// OrderRoute calcolato. Il server ricalcola sempre la geometria via ORS,
	// non si fida di quella eventualmente mandata dal client.
	RouteWaypoints []RouteWaypointDTO `json:"route_waypoints,omitempty"`
}

// RouteWaypointDTO identifica un punto per riferimento — usato in richiesta
// (assign, update-route) dove il client manda solo tipo+id, mai coordinate.
type RouteWaypointDTO struct {
	Tipo  string `json:"tipo" enums:"garage,destinazione,wash_station" validate:"required"`
	RefID string `json:"ref_id" validate:"required"`
}

// RouteWaypointResponseDTO è lo stesso waypoint risolto (nome+coordinate),
// per disegnare i marker sulla mappa senza un secondo giro di lookup lato client.
type RouteWaypointResponseDTO struct {
	Tipo  string  `json:"tipo"`
	RefID string  `json:"ref_id"`
	Nome  string  `json:"nome"`
	Lat   float64 `json:"lat"`
	Lng   float64 `json:"lng"`
}

type RouteResponseDTO struct {
	ID             uuid.UUID                  `json:"id"`
	Waypoints      []RouteWaypointResponseDTO `json:"waypoints"`
	Points         [][2]float64               `json:"points"`
	DistanceKm     float64                    `json:"distance_km"`
	DurationMin    int                        `json:"duration_min"`
	EditedManually bool                       `json:"edited_manually"`
}

// RouteAlternativeDTO è una delle fino a 3 proposte effimere di
// POST /orders/{id}/route-alternatives — mai scritta su DB finché il manager
// non la sceglie (a quel punto diventa un RouteResponseDTO persistito).
type RouteAlternativeDTO struct {
	Waypoints   []RouteWaypointResponseDTO `json:"waypoints"`
	Points      [][2]float64               `json:"points"`
	DistanceKm  float64                    `json:"distance_km"`
	DurationMin int                        `json:"duration_min"`
}

// GeocodeResultDTO is a forward-geocoding candidate — match quality depends
// on ORS/OpenStreetMap data coverage for the searched address, so several
// candidates are returned for the user to pick from rather than just one.
type GeocodeResultDTO struct {
	Label     string  `json:"label"`
	Indirizzo string  `json:"indirizzo"`
	Citta     string  `json:"citta"`
	Cap       string  `json:"cap"`
	Provincia string  `json:"provincia"`
	Nazione   string  `json:"nazione"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
}

type OrderRouteAlternativesRequest struct {
	GarageID      string `json:"garage_id"`
	WashStationID string `json:"wash_station_id"`
}

type OrderRouteAlternativesResponse struct {
	Alternatives []RouteAlternativeDTO `json:"alternatives"`
}

type OrderRouteUpdateRequest struct {
	Waypoints []RouteWaypointDTO `json:"waypoints" validate:"required,min=2"`
}

type OrderResponse struct {
	ID                    uuid.UUID                `json:"id"`
	Progressivo           string                   `json:"progressivo"`
	ClienteID             string                   `json:"cliente_id"`
	Cliente               *CustomerResponse        `json:"cliente"`
	CommittenteID         string                   `json:"committente_id"`
	Committente           *CustomerResponse        `json:"committente"`
	DestinazioneCaricoID  string                   `json:"destinazione_carico_id"`
	DestinazioneCarico    *DestinationResponse     `json:"destinazione_carico"`
	DestinazioneScaricoID string                   `json:"destinazione_scarico_id"`
	DestinazioneScarico   *DestinationResponse     `json:"destinazione_scarico"`
	DataRitiro            string                   `json:"data_ritiro"`
	OraRitiroDa           string                   `json:"ora_ritiro_da"`
	OraRitiroA            string                   `json:"ora_ritiro_a"`
	DataConsegna          string                   `json:"data_consegna"`
	OraConsegnaDa         string                   `json:"ora_consegna_da"`
	OraConsegnaA          string                   `json:"ora_consegna_a"`
	Tariffa               float64                  `json:"tariffa"`
	TipoTariffa           string                   `json:"tipo_tariffa"`
	Tipologia             string                   `json:"tipologia"`
	CategoriaTrasporto    string                   `json:"categoria_trasporto"`
	RifOrdineCliente      string                   `json:"rif_ordine_cliente"`
	RifCarico             string                   `json:"rif_carico"`
	NoteCarico            string                   `json:"note_carico"`
	RifScarico            string                   `json:"rif_scarico"`
	NoteScarico           string                   `json:"note_scarico"`
	AndataRitorno         bool                     `json:"andata_ritorno"`
	Provvisorio           bool                     `json:"provvisorio"`
	Note                  string                   `json:"note"`
	Items                 []OrderItemResponseDTO   `json:"items"`
	ServiziAccessori      []string                 `json:"servizi_accessori"`
	CostiAccessori        []map[string]interface{} `json:"costi_accessori"`
	Stato                 string                   `json:"stato" enums:"PIANIFICABILE,PIANIFICATO,VIAGGIO,CHIUSO,SCARTATO"`
	GarageID              string                   `json:"garage_id"`
	Garage                *GarageResponse          `json:"garage"`
	MotriceID             string                   `json:"motrice_id"`
	Motrice               *MotriceResponse         `json:"motrice"`
	SemirimorchioID       string                   `json:"semirimorchio_id"`
	Semirimorchio         *SemirimorchioResponse   `json:"semirimorchio"`
	AutistaID             string                   `json:"autista_id"`
	Autista               *DriverResponse          `json:"autista"`
	VettoreID             string                   `json:"vettore_id"`
	Vettore               *CarrierResponse         `json:"vettore"`
	WashStationID         string                   `json:"wash_station_id"`
	WashStation           *WashStationResponse     `json:"wash_station"`
	RouteID               string                   `json:"route_id"`
	Route                 *RouteResponseDTO        `json:"route"`
	ViaggioID             string                   `json:"viaggio_id"`
	FatturaID             string                   `json:"fattura_id"`
	CreatedAt             time.Time                `json:"created_at"`
	UpdatedAt             time.Time                `json:"updated_at"`
}

type OrderReturnSuggestion struct {
	Order   OrderResponse `json:"order"`
	Score   int           `json:"score"`
	Reasons []string      `json:"reasons"`
}

type OrderReturnSuggestionsResponse struct {
	Count       int                     `json:"count"`
	Candidates  []OrderReturnSuggestion `json:"candidates"`
	SourceOrder OrderSourceSummary      `json:"source_order"`
}

type OrderSourceSummary struct {
	ID                  uuid.UUID            `json:"id"`
	Progressivo         string               `json:"progressivo"`
	Cliente             *CustomerResponse    `json:"cliente"`
	DestinazioneScarico *DestinationResponse `json:"destinazione_scarico"`
	DataConsegna        string               `json:"data_consegna"`
}

// MotriceRequest mirrors MotriceCreate — the tractor-unit half of the
// former single Vehicle table (see Semirimorchio for the trailer half).
type MotriceRequest struct {
	Targa     string  `json:"targa" validate:"required"`
	Marca     string  `json:"marca"`
	Modello   string  `json:"modello"`
	Anno      int     `json:"anno"`
	PortataKg float64 `json:"portata_kg"`
	Note      string  `json:"note"`
}

type MotriceResponse struct {
	ID        uuid.UUID `json:"id"`
	Targa     string    `json:"targa"`
	Marca     string    `json:"marca"`
	Modello   string    `json:"modello"`
	Anno      int       `json:"anno"`
	PortataKg float64   `json:"portata_kg"`
	Note      string    `json:"note"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SemirimorchioRequest mirrors SemirimorchioCreate — the trailer half (see
// Motrice for the tractor half).
type SemirimorchioRequest struct {
	Targa         string  `json:"targa" validate:"required"`
	Tipo          string  `json:"tipo"`
	Scompartature int     `json:"scompartature"`
	PortataKg     float64 `json:"portata_kg"`
	Note          string  `json:"note"`
}

type SemirimorchioResponse struct {
	ID            uuid.UUID `json:"id"`
	Targa         string    `json:"targa"`
	Tipo          string    `json:"tipo"`
	Scompartature int       `json:"scompartature"`
	PortataKg     float64   `json:"portata_kg"`
	Note          string    `json:"note"`
	Active        bool      `json:"active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type MotriceAvailabilityResponse struct {
	MotriceResponse
	Disponibilita string `json:"disponibilita"`
}

type SemirimorchioAvailabilityResponse struct {
	SemirimorchioResponse
	Disponibilita string `json:"disponibilita"`
}

type DriverAvailabilityResponse struct {
	DriverResponse
	Disponibilita         string `json:"disponibilita"`
	MotivoIndisponibilita string `json:"motivo_indisponibilita"`
}

type TripSegmentDTO struct {
	Ordine           int     `json:"ordine"`
	Tipo             string  `json:"tipo"`
	OrigineNome      string  `json:"origine_nome"`
	OrigineLat       float64 `json:"origine_lat"`
	OrigineLng       float64 `json:"origine_lng"`
	DestinazioneNome string  `json:"destinazione_nome"`
	DestinazioneLat  float64 `json:"destinazione_lat"`
	DestinazioneLng  float64 `json:"destinazione_lng"`
	Km               float64 `json:"km"`
	TempoStimatoMin  int     `json:"tempo_stimato_min"`
	OrdineID         *string `json:"ordine_id"`
}

// TripRequest mirrors TripCreate — `segmenti` and `km_totali` are declared in
// the Python schema but always overwritten by the server (build_segments +
// total_km run unconditionally in create_trip), so they're omitted here as
// dead-on-arrival fields rather than modeled and ignored.
type TripRequest struct {
	OrdiniIds       []string `json:"ordini_ids"`
	MotriceID       string   `json:"motrice_id"`
	SemirimorchioID string   `json:"semirimorchio_id"`
	AutistaID       string   `json:"autista_id"`
	VettoreID       string   `json:"vettore_id"`
	GarageID        string   `json:"garage_id"`
	Note            string   `json:"note"`
	DataPartenza    string   `json:"data_partenza"`
	DataArrivo      string   `json:"data_arrivo"`
}

type TripResponse struct {
	ID              uuid.UUID              `json:"id"`
	OrdiniIds       []string               `json:"ordini_ids"`
	MotriceID       string                 `json:"motrice_id"`
	Motrice         *MotriceResponse       `json:"motrice"`
	SemirimorchioID string                 `json:"semirimorchio_id"`
	Semirimorchio   *SemirimorchioResponse `json:"semirimorchio"`
	AutistaID       string                 `json:"autista_id"`
	Autista         *DriverResponse        `json:"autista"`
	VettoreID       string                 `json:"vettore_id"`
	Vettore         *CarrierResponse       `json:"vettore"`
	GarageID        string                 `json:"garage_id"`
	Garage          *GarageResponse        `json:"garage"`
	Segmenti        []TripSegmentDTO       `json:"segmenti"`
	KmTotali        float64                `json:"km_totali"`
	CostoStimato    float64                `json:"costo_stimato"`
	Stato           string                 `json:"stato"`
	Note            string                 `json:"note"`
	DataPartenza    string                 `json:"data_partenza"`
	DataArrivo      string                 `json:"data_arrivo"`
	CreatedAt       time.Time              `json:"created_at"`
}

// TripDetailResponse is returned only by GET /trips/{id}, which additionally
// joins the trip's orders (matching Python's get_trip appending `ordini`).
type TripDetailResponse struct {
	TripResponse
	Ordini []OrderResponse `json:"ordini"`
}

type RecomputeSegmentsResult struct {
	OK            bool    `json:"ok"`
	SegmentiCount int     `json:"segmenti_count"`
	KmTotali      float64 `json:"km_totali"`
}

type OKResult struct {
	OK bool `json:"ok"`
}

// PriceListItemRequestDTO is the write-side shape of a price list rule —
// only reference ids, the server resolves/validates the rest via Preload.
type PriceListItemRequestDTO struct {
	ItemID                    *uuid.UUID `json:"item_id,omitempty"`
	ProdottoID                string     `json:"prodotto_id"`
	DestinazioneCaricoID      string     `json:"destinazione_carico_id"`
	DestinazioneScaricoID     string     `json:"destinazione_scarico_id"`
	Tariffa                   float64    `json:"tariffa"`
	TipoTariffa               string     `json:"tipo_tariffa"`
	RangePesoMin              float64    `json:"range_peso_min"`
	RangePesoMax              float64    `json:"range_peso_max"`
	UnitaPeso                 string     `json:"unita_peso"`
	MinimoTassabile           float64    `json:"minimo_tassabile"`
	TipoTrasporto             string     `json:"tipo_trasporto"`
	PercAdeguamentoCarburante float64    `json:"perc_adeguamento_carburante"`
}

// PriceListItemResponseDTO is the read-side shape — Prodotto/DestinazioneCarico/
// DestinazioneScarico are the full nested rows (nil when the rule applies to
// "any", e.g. a blanket weight-based rate with no destination filter).
type PriceListItemResponseDTO struct {
	ItemID                    uuid.UUID            `json:"item_id"`
	Prodotto                  *ProductResponse     `json:"prodotto"`
	DestinazioneCarico        *DestinationResponse `json:"destinazione_carico"`
	DestinazioneScarico       *DestinationResponse `json:"destinazione_scarico"`
	Tariffa                   float64              `json:"tariffa"`
	TipoTariffa               string               `json:"tipo_tariffa"`
	RangePesoMin              float64              `json:"range_peso_min"`
	RangePesoMax              float64              `json:"range_peso_max"`
	UnitaPeso                 string               `json:"unita_peso"`
	MinimoTassabile           float64              `json:"minimo_tassabile"`
	TipoTrasporto             string               `json:"tipo_trasporto"`
	PercAdeguamentoCarburante float64              `json:"perc_adeguamento_carburante"`
}

// PriceListRequest mirrors PriceListCreate — used for both create and the
// (non-duplicating) branch of update.
type PriceListRequest struct {
	ClienteID  string                    `json:"cliente_id" validate:"required"`
	DataInizio string                    `json:"data_inizio"`
	DataFine   string                    `json:"data_fine"`
	Items      []PriceListItemRequestDTO `json:"items"`
	Note       string                    `json:"note"`
}

type PriceListResponse struct {
	ID         uuid.UUID                  `json:"id"`
	ClienteID  string                     `json:"cliente_id"`
	Cliente    *CustomerResponse          `json:"cliente"`
	DataInizio string                     `json:"data_inizio"`
	DataFine   string                     `json:"data_fine"`
	Items      []PriceListItemResponseDTO `json:"items"`
	Note       string                     `json:"note"`
	InUso      bool                       `json:"in_uso"`
	Active     bool                       `json:"active"`
	CreatedAt  time.Time                  `json:"created_at"`
}

// PriceListUpdateResult mirrors update_pricelist's response — `duplicated`
// tells the caller whether the update created a new version (because the
// original was in_uso) instead of updating in place.
type PriceListUpdateResult struct {
	OK         bool       `json:"ok"`
	NewID      *uuid.UUID `json:"new_id,omitempty"`
	Duplicated bool       `json:"duplicated"`
}

type PriceListItemAddResult struct {
	OK         bool      `json:"ok"`
	ItemID     uuid.UUID `json:"item_id"`
	ItemsCount int       `json:"items_count"`
}

type PriceListItemUpdateResult struct {
	OK     bool      `json:"ok"`
	ItemID uuid.UUID `json:"item_id"`
}

type PriceListItemDeleteResult struct {
	OK         bool `json:"ok"`
	ItemsCount int  `json:"items_count"`
}

type TariffLookupResult struct {
	Found                     bool      `json:"found"`
	Tariffa                   float64   `json:"tariffa"`
	TariffaBase               float64   `json:"tariffa_base,omitempty"`
	TipoTariffa               string    `json:"tipo_tariffa"`
	PercAdeguamentoCarburante float64   `json:"perc_adeguamento_carburante,omitempty"`
	MinimoTassabile           float64   `json:"minimo_tassabile,omitempty"`
	ListinoID                 uuid.UUID `json:"listino_id,omitempty"`
	ItemID                    uuid.UUID `json:"item_id,omitempty"`
	Score                     int       `json:"score,omitempty"`
}

type InvoiceLineDTO struct {
	OrdineID    string  `json:"ordine_id"`
	Descrizione string  `json:"descrizione"`
	Prodotto    string  `json:"prodotto"`
	Peso        float64 `json:"peso"`
	Quantita    float64 `json:"quantita"`
	Tariffa     float64 `json:"tariffa"`
	Totale      float64 `json:"totale"`
	IvaCodice   string  `json:"iva_codice"`
}

// InvoiceRequest mirrors InvoiceCreate. `numero`/`stato` are server-assigned
// (progressive sequence + "PROFORMA"), matching the Python schema which
// excludes them from the create-able fields.
type InvoiceRequest struct {
	ClienteID           string                   `json:"cliente_id" validate:"required"`
	DataFattura         string                   `json:"data_fattura"`
	DataScadenza        string                   `json:"data_scadenza"`
	CondizioniPagamento string                   `json:"condizioni_pagamento"`
	Righe               []InvoiceLineDTO         `json:"righe"`
	CostiAccessori      []map[string]interface{} `json:"costi_accessori"`
	TotaleImponibile    float64                  `json:"totale_imponibile"`
	TotaleIva           float64                  `json:"totale_iva"`
	Totale              float64                  `json:"totale"`
}

type InvoiceResponse struct {
	ID                  uuid.UUID                `json:"id"`
	Numero              string                   `json:"numero"`
	ClienteID           string                   `json:"cliente_id"`
	Cliente             *CustomerResponse        `json:"cliente"`
	DataFattura         string                   `json:"data_fattura"`
	DataScadenza        string                   `json:"data_scadenza"`
	CondizioniPagamento string                   `json:"condizioni_pagamento"`
	Righe               []InvoiceLineDTO         `json:"righe"`
	CostiAccessori      []map[string]interface{} `json:"costi_accessori"`
	TotaleImponibile    float64                  `json:"totale_imponibile"`
	TotaleIva           float64                  `json:"totale_iva"`
	Totale              float64                  `json:"totale"`
	Stato               string                   `json:"stato"`
	Tipo                string                   `json:"tipo"`
	Note                string                   `json:"note"`
	PdfS3Key            *string                  `json:"pdf_s3_key"`
	PdfUploadedAt       *string                  `json:"pdf_uploaded_at"`
	PdfRetainUntil      *string                  `json:"pdf_retain_until"`
	CreatedAt           time.Time                `json:"created_at"`
}

type InvoiceFinalizeResult struct {
	OK          bool    `json:"ok"`
	PdfArchived bool    `json:"pdf_archived"`
	PdfS3Key    *string `json:"pdf_s3_key"`
}

type InvoicePDFURLResult struct {
	URL         string  `json:"url"`
	ExpiresIn   *int    `json:"expires_in"`
	RetainUntil *string `json:"retain_until"`
}

type MonthlyOrderTrend struct {
	Mese   string  `json:"mese"`
	Ordini int64   `json:"ordini"`
	Totale float64 `json:"totale"`
}

type DashboardStatsResponse struct {
	TotalOrders       int64               `json:"total_orders"`
	Pianificabili     int64               `json:"pianificabili"`
	InViaggio         int64               `json:"in_viaggio"`
	Chiusi            int64               `json:"chiusi"`
	Fatturati         int64               `json:"fatturati"`
	TotalCustomers    int64               `json:"total_customers"`
	TotalMotrici      int64               `json:"total_motrici"`
	TotalSemirimorchi int64               `json:"total_semirimorchi"`
	TotalDrivers      int64               `json:"total_drivers"`
	TotalRevenue      float64             `json:"total_revenue"`
	MonthlyTrend      []MonthlyOrderTrend `json:"monthly_trend"`
}

type CustomerDashboardSummary struct {
	ID             uuid.UUID `json:"id"`
	RagioneSociale string    `json:"ragione_sociale"`
	Citta          string    `json:"citta"`
	PartitaIva     string    `json:"partita_iva"`
}

type CustomerDashboardKPI struct {
	OrdiniTotali        int64   `json:"ordini_totali"`
	OrdiniFatturati     int64   `json:"ordini_fatturati"`
	OrdiniChiusi        int64   `json:"ordini_chiusi"`
	OrdiniInViaggio     int64   `json:"ordini_in_viaggio"`
	OrdiniPianificabili int64   `json:"ordini_pianificabili"`
	FatturatoNetto      float64 `json:"fatturato_netto"`
	TariffaMedia        float64 `json:"tariffa_media"`
}

type CustomerDashboardMonthly struct {
	Mese      string  `json:"mese"`
	Ordini    int64   `json:"ordini"`
	Fatturato float64 `json:"fatturato"`
}

type CustomerDashboardDestination struct {
	Nome      string  `json:"nome"`
	Ordini    int64   `json:"ordini"`
	Fatturato float64 `json:"fatturato"`
}

type CustomerDashboardTipologia struct {
	Tipologia string `json:"tipologia"`
	Ordini    int64  `json:"ordini"`
}

type CustomerDashboardCategoria struct {
	Categoria string `json:"categoria"`
	Ordini    int64  `json:"ordini"`
}

type CustomerDashboardResponse struct {
	Customer        CustomerDashboardSummary       `json:"customer"`
	KPI             CustomerDashboardKPI           `json:"kpi"`
	MonthlyTrend    []CustomerDashboardMonthly     `json:"monthly_trend"`
	TopDestinazioni []CustomerDashboardDestination `json:"top_destinazioni"`
	PerTipologia    []CustomerDashboardTipologia   `json:"per_tipologia"`
	PerCategoria    []CustomerDashboardCategoria   `json:"per_categoria"`
}

type MapPoint struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type MapPOI struct {
	ID   string  `json:"id"`
	Nome string  `json:"nome"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

// MapNamedPoint is a named location on the map with no other data attached —
// used for both garages and wash stations (see MapTripsResponse).
type MapNamedPoint struct {
	Nome string  `json:"nome"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

type MapRoute struct {
	ID              uuid.UUID         `json:"id"`
	Progressivo     string            `json:"progressivo"`
	Cliente         *CustomerResponse `json:"cliente"`
	Stato           string            `json:"stato"`
	Tipologia       string            `json:"tipologia"`
	Motrice         *MotriceResponse  `json:"motrice"`
	Autista         *DriverResponse   `json:"autista"`
	DataRitiro      string            `json:"data_ritiro"`
	DataConsegna    string            `json:"data_consegna"`
	Tariffa         float64           `json:"tariffa"`
	Carico          MapPoint          `json:"carico"`
	Scarico         MapPoint          `json:"scarico"`
	CurrentPosition MapPoint          `json:"current_position"`
	Progress        float64           `json:"progress"`
	RoadPoints      []MapPoint        `json:"road_points"`
	DistanceKm      float64           `json:"distance_km"`
	DurationHours   float64           `json:"duration_hours"`
	RemainingKm     float64           `json:"remaining_km"`
	EtaHours        float64           `json:"eta_hours"`
	Garage          *MapNamedPoint    `json:"garage"`
	WashStation     *MapNamedPoint    `json:"wash_station"`
}

type MapStats struct {
	InViaggio     int `json:"in_viaggio"`
	Pianificabili int `json:"pianificabili"`
	Chiusi        int `json:"chiusi"`
}

type MapTripsResponse struct {
	Routes       []MapRoute      `json:"routes"`
	POI          []MapPOI        `json:"poi"`
	Garages      []MapNamedPoint `json:"garages"`
	WashStations []MapNamedPoint `json:"wash_stations"`
	Stats        MapStats        `json:"stats"`
}

type CustomerResponse struct {
	ID                  uuid.UUID `json:"id"`
	RagioneSociale      string    `json:"ragione_sociale"`
	Indirizzo           string    `json:"indirizzo"`
	Citta               string    `json:"citta"`
	Cap                 string    `json:"cap"`
	Provincia           string    `json:"provincia"`
	Nazione             string    `json:"nazione"`
	Lat                 *float64  `json:"lat"`
	Lng                 *float64  `json:"lng"`
	PartitaIva          string    `json:"partita_iva"`
	CodiceFiscale       string    `json:"codice_fiscale"`
	Telefono            string    `json:"telefono"`
	Email               string    `json:"email"`
	Pec                 string    `json:"pec"`
	CondizioniPagamento string    `json:"condizioni_pagamento"`
	Note                string    `json:"note"`
	RichiedeRifOrdine   bool      `json:"richiede_rif_ordine"`
	Active              bool      `json:"active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// ── PDF templates + import ordini da PDF (porting OrderMesh) ──────────────

// PdfTemplateFieldDTO maps a rectangular zone of the PDF onto one inbound
// order field. Bounds are normalized 0..1 relative to the page, independent
// from render resolution. Target must be one of
// models.InboundOrderFieldTargets.
type PdfTemplateFieldDTO struct {
	ID     string  `json:"id"`
	Target string  `json:"target" validate:"required"`
	Label  string  `json:"label"`
	Page   int     `json:"page"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	W      float64 `json:"w"`
	H      float64 `json:"h"`
}

type PdfTemplateRequest struct {
	Name string `json:"name" validate:"required"`
	// Client is the default client name for orders imported with this template.
	Client string `json:"client"`
	// Senders holds full addresses or "@domain" patterns used to preselect
	// the template from a mail sender.
	Senders []string              `json:"senders"`
	Fields  []PdfTemplateFieldDTO `json:"fields"`
}

type PdfTemplateResponse struct {
	ID        uuid.UUID             `json:"id"`
	Name      string                `json:"name"`
	Client    string                `json:"client"`
	Senders   []string              `json:"senders"`
	Fields    []PdfTemplateFieldDTO `json:"fields"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

// PdfTemplateMatchResponse wraps the best template for a sender; Match is
// null when nothing matches (the UI then asks for a manual choice).
type PdfTemplateMatchResponse struct {
	Match *PdfTemplateResponse `json:"match"`
}

// PdfRenderBlockDTO is one detected text block with normalized bounds — the
// clickable suggestion in the template editor.
type PdfRenderBlockDTO struct {
	Text       string             `json:"text"`
	BoundsNorm map[string]float64 `json:"bounds_norm"`
}

type PdfRenderPageDTO struct {
	PageNum  int                 `json:"page_num"`
	ImageB64 string              `json:"image_b64"`
	Width    int                 `json:"width"`
	Height   int                 `json:"height"`
	Blocks   []PdfRenderBlockDTO `json:"blocks"`
}

type PdfRenderResponse struct {
	Filename  string             `json:"filename"`
	PageCount int                `json:"page_count"`
	Pages     []PdfRenderPageDTO `json:"pages"`
}

// PdfExtractedValueDTO is the outcome of reading one template zone: Method
// says how ("poppler-text", "claude-vision", "empty", "page-out-of-range",
// "render-error", "skipped-too-small") so the UI can flag uncertain zones.
type PdfExtractedValueDTO struct {
	Value      string  `json:"value"`
	Confidence float64 `json:"confidence"`
	Method     string  `json:"method"`
}

// InboundOrderDraftDTO is the order draft built from a PDF extraction —
// NOT persisted: the operator reviews it in the UI and confirms via
// POST /inbound-orders. Field names mirror models.InboundOrder.
type InboundOrderDraftDTO struct {
	Client        string     `json:"client"`
	SenderEmail   string     `json:"sender_email"`
	Ref           string     `json:"ref"`
	Product       string     `json:"product"`
	Kg            int        `json:"kg"`
	LoadDate      string     `json:"load_date"`
	LoadPlace     string     `json:"load_place"`
	DeliveryDate  string     `json:"delivery_date"`
	DeliveryPlace string     `json:"delivery_place"`
	Rate          string     `json:"rate"`
	Notes         string     `json:"notes"`
	Status        string     `json:"status"`
	Source        string     `json:"source"`
	TemplateID    *uuid.UUID `json:"template_id,omitempty"`
	ReceivedAt    time.Time  `json:"received_at"`
}

// PdfTemplateRef identifies which template produced an import result.
type PdfTemplateRef struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// PdfExtractionResponse is the shared response of POST /pdf/test and
// POST /pdf/import: the draft order, the merged value per target field and
// the raw per-zone extraction detail. Template is set on /pdf/import only.
type PdfExtractionResponse struct {
	Order      InboundOrderDraftDTO            `json:"order"`
	Values     map[string]string               `json:"values"`
	Extraction map[string]PdfExtractedValueDTO `json:"extraction"`
	Template   *PdfTemplateRef                 `json:"template,omitempty"`
}

// ── Inbound orders (dashboard di accettazione, porting OrderMesh) ─────────

// InboundOrderRequest confirms a draft (typically from /pdf/import) as an
// inbound order. Client plus at least one of ref/product are required —
// the service enforces the ref-or-product half. Status/Source/ReceivedAt
// are optional and default to pending/pdf/now, mirroring OrderMesh's
// POST /api/orders.
type InboundOrderRequest struct {
	Client        string     `json:"client" validate:"required"`
	SenderEmail   string     `json:"sender_email"`
	Ref           string     `json:"ref"`
	Product       string     `json:"product"`
	Kg            int        `json:"kg"`
	LoadDate      string     `json:"load_date"`
	LoadPlace     string     `json:"load_place"`
	DeliveryDate  string     `json:"delivery_date"`
	DeliveryPlace string     `json:"delivery_place"`
	Rate          string     `json:"rate"`
	Notes         string     `json:"notes"`
	Portal        bool       `json:"portal"`
	Status        string     `json:"status" validate:"omitempty,oneof=pending accepted modify"`
	Source        string     `json:"source" validate:"omitempty,oneof=seed mail pdf portal"`
	TemplateID    *uuid.UUID `json:"template_id"`
	ReceivedAt    *time.Time `json:"received_at"`
	// ClienteID is set internally by CreateMyInboundOrder (client portal) —
	// never trusted from external request bodies for the staff-facing
	// /inbound-orders endpoint, same posture as OrderRequest.ClienteID on /me/orders.
	ClienteID *uuid.UUID `json:"cliente_id,omitempty"`
	// Structured portal payload, like ClienteID set internally by
	// CreateMyInboundOrder from an authenticated submission — the ids the
	// client picked from its own destination list, kept so Convert can
	// rebuild the Order exactly instead of guessing FKs from the place
	// names. Nil for mail/pdf/seed drafts, which only ever have free text.
	CommittenteID         *uuid.UUID `json:"committente_id,omitempty"`
	DestinazioneCaricoID  *uuid.UUID `json:"destinazione_carico_id,omitempty"`
	DestinazioneScaricoID *uuid.UUID `json:"destinazione_scarico_id,omitempty"`
	OraRitiroDa           string     `json:"ora_ritiro_da"`
	OraRitiroA            string     `json:"ora_ritiro_a"`
	OraConsegnaDa         string     `json:"ora_consegna_da"`
	OraConsegnaA          string     `json:"ora_consegna_a"`
	// TariffaProposta: the client's "tariffa desiderata", a proposal only.
	TariffaProposta float64 `json:"tariffa_proposta"`
}

type InboundOrderResponse struct {
	ID            uuid.UUID  `json:"id"`
	Client        string     `json:"client"`
	SenderEmail   string     `json:"sender_email"`
	Ref           string     `json:"ref"`
	Product       string     `json:"product"`
	Kg            int        `json:"kg"`
	LoadDate      string     `json:"load_date"`
	LoadPlace     string     `json:"load_place"`
	DeliveryDate  string     `json:"delivery_date"`
	DeliveryPlace string     `json:"delivery_place"`
	Rate          string     `json:"rate"`
	Notes         string     `json:"notes"`
	Portal        bool       `json:"portal"`
	Status        string     `json:"status"`
	Source        string     `json:"source"`
	TemplateID    *uuid.UUID `json:"template_id,omitempty"`
	ReceivedAt    time.Time  `json:"received_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	ClienteID     *uuid.UUID `json:"cliente_id,omitempty"`

	CommittenteID         *uuid.UUID `json:"committente_id,omitempty"`
	DestinazioneCaricoID  *uuid.UUID `json:"destinazione_carico_id,omitempty"`
	DestinazioneScaricoID *uuid.UUID `json:"destinazione_scarico_id,omitempty"`
	OraRitiroDa           string     `json:"ora_ritiro_da"`
	OraRitiroA            string     `json:"ora_ritiro_a"`
	OraConsegnaDa         string     `json:"ora_consegna_da"`
	OraConsegnaA          string     `json:"ora_consegna_a"`
	TariffaProposta       float64    `json:"tariffa_proposta"`
	// OrderID: set once the draft has been converted into a real order, nil
	// while it is still awaiting conversion. The client portal uses it to
	// tell "richiesta in attesa" from "ordine confermato".
	OrderID *uuid.UUID `json:"order_id,omitempty"`
}

// InboundOrderActionResponse is returned by the accept action: the updated
// order plus a human-readable note about the confirmation mail ("inviata
// a ...", or why none was sent) — mirrors OrderMesh's {order, mail} shape.
type InboundOrderActionResponse struct {
	Order InboundOrderResponse `json:"order"`
	Mail  string               `json:"mail,omitempty"`
}

// InboundOrderConvertRequest is the operator's input to turn a draft into a
// real order. Every field is optional and overrides what the draft carries;
// ClienteID is the exception that is sometimes mandatory — see
// inboundorders.Convert. Nothing here is ever read from the draft's free
// text: a mail/pdf draft's Client is attacker-controlled, so the customer to
// bill is always either a trusted id stored at submission time (portal) or a
// deliberate choice made here by staff.
type InboundOrderConvertRequest struct {
	ClienteID             string `json:"cliente_id" validate:"omitempty,uuid4"`
	CommittenteID         string `json:"committente_id" validate:"omitempty,uuid4"`
	DestinazioneCaricoID  string `json:"destinazione_carico_id" validate:"omitempty,uuid4"`
	DestinazioneScaricoID string `json:"destinazione_scarico_id" validate:"omitempty,uuid4"`
	DataRitiro            string `json:"data_ritiro"`
	DataConsegna          string `json:"data_consegna"`
	// Tariffa: when omitted the draft's TariffaProposta is applied, so pass
	// it explicitly to price the order at anything other than what the
	// customer proposed. A pointer so an explicit 0 (free of charge) is
	// distinguishable from "field absent".
	Tariffa     *float64 `json:"tariffa"`
	TipoTariffa string   `json:"tipo_tariffa" validate:"omitempty,oneof=forfait tonnellata km"`
	Tipologia   string   `json:"tipologia" validate:"omitempty,oneof=nazionale internazionale"`
	Note        string   `json:"note"`
}

// InboundOrderConvertResponse reports both sides of the conversion: the draft
// now carrying OrderID, and the order that was created. TariffaFromCliente
// flags that the applied price is the customer's own proposal, unreviewed —
// the one number in the result that nobody on the FECCIA side chose.
type InboundOrderConvertResponse struct {
	InboundOrder      InboundOrderResponse `json:"inbound_order"`
	Order             OrderResponse        `json:"order"`
	TariffaFromClient bool                 `json:"tariffa_from_client"`
}

// InboundScrapeResponse reports one mailbox read: how many order mails were
// examined and how many new orders were stored.
type InboundScrapeResponse struct {
	Added   int `json:"added"`
	Scanned int `json:"scanned"`
}

// InboundConfigResponse mirrors OrderMesh's GET /api/config: the runtime
// readiness of every optional piece of the inbound pipeline, so the UI can
// enable/disable the scan and import actions accordingly.
type InboundConfigResponse struct {
	AcceptMode        string `json:"accept_mode"`
	TestRecipient     string `json:"test_recipient"`
	SmtpReady         bool   `json:"smtp_ready"`
	MailboxReady      bool   `json:"mailbox_ready"`
	Backend           string `json:"backend"`
	SubjectFilter     string `json:"subject_filter"`
	ScrapeIntervalMin int    `json:"scrape_interval_min"`
	PdfReady          bool   `json:"pdf_ready"`
	VisionReady       bool   `json:"vision_ready"`
}

// ─────────────────────── Liste paginate (Anagrafiche) ───────────────────────
// Envelope {data, total} per gli endpoint di elenco Anagrafiche con paginazione
// server-side (page/limit — vedi pkg/utils.PageParams). Un tipo concreto per
// entità (invece del generico utils.ListResult usato da ListUsers) così swag
// genera un tipo TS `data`/`total` corretto invece di `map[string]interface{}`.

type CustomerListResponse struct {
	Data  []CustomerResponse `json:"data"`
	Total int64              `json:"total"`
}

type MotriceListResponse struct {
	Data  []MotriceResponse `json:"data"`
	Total int64             `json:"total"`
}

type SemirimorchioListResponse struct {
	Data  []SemirimorchioResponse `json:"data"`
	Total int64                   `json:"total"`
}

type DriverListResponse struct {
	Data  []DriverResponse `json:"data"`
	Total int64            `json:"total"`
}

type CarrierListResponse struct {
	Data  []CarrierResponse `json:"data"`
	Total int64             `json:"total"`
}

type ProductListResponse struct {
	Data  []ProductResponse `json:"data"`
	Total int64             `json:"total"`
}

type GarageListResponse struct {
	Data  []GarageResponse `json:"data"`
	Total int64            `json:"total"`
}

type WashStationListResponse struct {
	Data  []WashStationResponse `json:"data"`
	Total int64                 `json:"total"`
}

type DestinationListResponse struct {
	Data  []DestinationResponse `json:"data"`
	Total int64                 `json:"total"`
}

type CountryListResponse struct {
	Data  []CountryResponse `json:"data"`
	Total int64             `json:"total"`
}

type BankListResponse struct {
	Data  []BankResponse `json:"data"`
	Total int64          `json:"total"`
}

type AccountingEntryListResponse struct {
	Data  []AccountingEntryResponse `json:"data"`
	Total int64                     `json:"total"`
}
