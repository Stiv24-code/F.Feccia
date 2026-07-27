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
	Active    bool    `json:"active"`
}

// RegisterRequest mirrors Python's UserCreate — the password policy
// (min 12 chars) matches backend/models.py's UserCreate.password field
// exactly, since this is admin-facing user provisioning, not self-signup.
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=1"`
	Password string `json:"password" validate:"required,min=12"`
	Role     string `json:"role" validate:"required,oneof=admin amministrazione planner operatore"`
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
	Role     string  `json:"role" validate:"required,oneof=admin amministrazione planner operatore"`
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
	RagioneSociale      string `json:"ragione_sociale" validate:"required"`
	Indirizzo           string `json:"indirizzo"`
	Citta               string `json:"citta"`
	Cap                 string `json:"cap"`
	Provincia           string `json:"provincia"`
	Nazione             string `json:"nazione"`
	PartitaIva          string `json:"partita_iva"`
	CodiceFiscale       string `json:"codice_fiscale"`
	Telefono            string `json:"telefono"`
	Email               string `json:"email"`
	Pec                 string `json:"pec"`
	CondizioniPagamento string `json:"condizioni_pagamento"`
	Note                string `json:"note"`
	RichiedeRifOrdine   bool   `json:"richiede_rif_ordine"`
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
	Note      string    `json:"note"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DriverRequest struct {
	Nome            string  `json:"nome" validate:"required"`
	Cognome         string  `json:"cognome" validate:"required"`
	CodiceFiscale   string  `json:"codice_fiscale"`
	Patente         string  `json:"patente"`
	ScadenzaPatente *string `json:"scadenza_patente"`
	Telefono        string  `json:"telefono"`
	Email           string  `json:"email"`
	Note            string  `json:"note"`
}

type DriverResponse struct {
	ID              uuid.UUID `json:"id"`
	Nome            string    `json:"nome"`
	Cognome         string    `json:"cognome"`
	CodiceFiscale   string    `json:"codice_fiscale"`
	Patente         string    `json:"patente"`
	ScadenzaPatente *string   `json:"scadenza_patente"`
	Telefono        string    `json:"telefono"`
	Email           string    `json:"email"`
	Note            string    `json:"note"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
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
	Motivo      string    `json:"motivo"`
	Note        string    `json:"note"`
}

type DriverUnavailabilityResponse struct {
	ID          uuid.UUID `json:"id"`
	AutistaID   uuid.UUID `json:"autista_id"`
	AutistaNome string    `json:"autista_nome"`
	DataDa      string    `json:"data_da"`
	DataA       string    `json:"data_a"`
	Motivo      string    `json:"motivo"`
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
// fields (stato, targa_motrice, autista_id, vettore_id, viaggio_id, fattura_id,
// progressivo) are intentionally absent, exactly like the Python schema, so a
// PUT can never touch them. Reference fields are ids only — the server no
// longer stores a client-submitted denormalized name, it's always derived
// from the live associated row via Preload.
type OrderRequest struct {
	ClienteID             string                   `json:"cliente_id" validate:"required"`
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
	AndataRitorno         bool                     `json:"andata_ritorno"`
	Note                  string                   `json:"note"`
	Items                 []OrderItemRequestDTO    `json:"items"`
	ServiziAccessori      []string                 `json:"servizi_accessori"`
	CostiAccessori        []map[string]interface{} `json:"costi_accessori"`
}

type OrderAssignRequest struct {
	GarageID       string `json:"garage_id"`
	TargaMotrice   string `json:"targa_motrice"`
	TargaRimorchio string `json:"targa_rimorchio"`
	AutistaID      string `json:"autista_id"`
	VettoreID      string `json:"vettore_id"`
	WashStationID  string `json:"wash_station_id"`
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
	AndataRitorno         bool                     `json:"andata_ritorno"`
	Note                  string                   `json:"note"`
	Items                 []OrderItemResponseDTO   `json:"items"`
	ServiziAccessori      []string                 `json:"servizi_accessori"`
	CostiAccessori        []map[string]interface{} `json:"costi_accessori"`
	Stato                 string                   `json:"stato" enums:"PIANIFICABILE,PIANIFICATO,VIAGGIO,CHIUSO,SCARTATO"`
	GarageID              string                   `json:"garage_id"`
	Garage                *GarageResponse          `json:"garage"`
	TargaMotrice          string                   `json:"targa_motrice"`
	TargaRimorchio        string                   `json:"targa_rimorchio"`
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

// VehicleRequest mirrors VehicleCreate — telemetry fields (GPS/temperature)
// are intentionally absent, exactly like the Python schema, so Create/Update
// can never touch them (they're written by separate GPS/temperature endpoints).
type VehicleRequest struct {
	Targa          string  `json:"targa" validate:"required"`
	TipoVeicolo    string  `json:"tipo_veicolo"`
	Marca          string  `json:"marca"`
	Modello        string  `json:"modello"`
	Anno           int     `json:"anno"`
	Scompartature  int     `json:"scompartature"`
	PortataKg      float64 `json:"portata_kg"`
	Note           string  `json:"note"`
	GpsTrackerUrl  string  `json:"gps_tracker_url"`
	GpsTrackerTipo string  `json:"gps_tracker_tipo"`
	GpsApiKey      string  `json:"gps_api_key"`
}

type VehicleResponse struct {
	ID             uuid.UUID `json:"id"`
	Targa          string    `json:"targa"`
	TipoVeicolo    string    `json:"tipo_veicolo"`
	Marca          string    `json:"marca"`
	Modello        string    `json:"modello"`
	Anno           int       `json:"anno"`
	Scompartature  int       `json:"scompartature"`
	PortataKg      float64   `json:"portata_kg"`
	Note           string    `json:"note"`
	GpsTrackerUrl  string    `json:"gps_tracker_url"`
	GpsTrackerTipo string    `json:"gps_tracker_tipo"`
	GpsApiKey      string    `json:"gps_api_key"`

	LastLat       float64 `json:"last_lat"`
	LastLng       float64 `json:"last_lng"`
	LastSpeedKmh  float64 `json:"last_speed_kmh"`
	LastHeading   float64 `json:"last_heading"`
	LastGpsUpdate string  `json:"last_gps_update"`
	GpsActive     bool    `json:"gps_active"`
	GpsSource     string  `json:"gps_source"`

	TempMin          *float64 `json:"temp_min"`
	TempMax          *float64 `json:"temp_max"`
	LastTempCelsius  *float64 `json:"last_temp_celsius"`
	LastTempTs       string   `json:"last_temp_ts"`
	LastTempSensorID string   `json:"last_temp_sensor_id"`
	LastTempAlert    bool     `json:"last_temp_alert"`

	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type VehicleGPSUpdateRequest struct {
	Lat       float64 `json:"lat" validate:"required"`
	Lng       float64 `json:"lng" validate:"required"`
	SpeedKmh  float64 `json:"speed_kmh"`
	Heading   float64 `json:"heading"`
	Timestamp string  `json:"timestamp"`
}

type VehicleAvailabilityResponse struct {
	VehicleResponse
	Disponibilita string `json:"disponibilita"`
}

type DriverAvailabilityResponse struct {
	DriverResponse
	Disponibilita         string `json:"disponibilita"`
	MotivoIndisponibilita string `json:"motivo_indisponibilita"`
}

type GPSUpdateResult struct {
	OK        bool             `json:"ok"`
	Targa     string           `json:"targa"`
	GpsSource string           `json:"gps_source,omitempty"`
	Position  GPSPositionShort `json:"position"`
}

type GPSPositionShort struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type GPSHistoryResponse struct {
	VehicleID uuid.UUID `json:"vehicle_id"`
	Targa     string    `json:"targa"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	SpeedKmh  float64   `json:"speed_kmh"`
	Heading   float64   `json:"heading"`
	Timestamp string    `json:"timestamp"`
	GpsSource string    `json:"gps_source"`
}

type GPSLiveVehicle struct {
	ID            uuid.UUID `json:"id"`
	Targa         string    `json:"targa"`
	Marca         string    `json:"marca"`
	Modello       string    `json:"modello"`
	TipoVeicolo   string    `json:"tipo_veicolo"`
	LastLat       float64   `json:"last_lat"`
	LastLng       float64   `json:"last_lng"`
	LastSpeedKmh  float64   `json:"last_speed_kmh"`
	LastHeading   float64   `json:"last_heading"`
	LastGpsUpdate string    `json:"last_gps_update"`
	GpsActive     bool      `json:"gps_active"`
	GpsTrackerUrl string    `json:"gps_tracker_url"`
	GpsSource     string    `json:"gps_source"`
}

// GPSWebhookPayload mirrors the normalized V1 webhook payload accepted by
// POST /api/v1/webhooks/gps/{vendor}.
type GPSWebhookPayload struct {
	Targa     string  `json:"targa"`
	VehicleID string  `json:"vehicle_id"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	SpeedKmh  float64 `json:"speed_kmh"`
	Heading   float64 `json:"heading"`
	Timestamp string  `json:"timestamp"`
}

type TemperatureReadingResponse struct {
	VehicleID   uuid.UUID `json:"vehicle_id"`
	Targa       string    `json:"targa"`
	TempCelsius float64   `json:"temp_celsius"`
	SensorID    string    `json:"sensor_id"`
	Ts          string    `json:"ts"`
	Source      string    `json:"source"`
	OutOfRange  bool      `json:"out_of_range"`
}

type TemperatureWebhookRequest struct {
	Targa       string  `json:"targa"`
	VehicleID   string  `json:"vehicle_id"`
	SensorID    string  `json:"sensor_id"`
	TempCelsius float64 `json:"temp_celsius" validate:"required"`
	Ts          string  `json:"ts"`
}

type TemperatureWebhookResult struct {
	OK         bool `json:"ok"`
	OutOfRange bool `json:"out_of_range"`
	Alert      bool `json:"alert"`
}

type TemperatureThresholdsRequest struct {
	TempMin *float64 `json:"temp_min"`
	TempMax *float64 `json:"temp_max"`
}

type TemperatureThresholdsResult struct {
	OK      bool     `json:"ok"`
	TempMin *float64 `json:"temp_min"`
	TempMax *float64 `json:"temp_max"`
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
	OrdiniIds      []string `json:"ordini_ids"`
	TargaMotrice   string   `json:"targa_motrice"`
	TargaRimorchio string   `json:"targa_rimorchio"`
	AutistaID      string   `json:"autista_id"`
	VettoreID      string   `json:"vettore_id"`
	GarageID       string   `json:"garage_id"`
	Note           string   `json:"note"`
	DataPartenza   string   `json:"data_partenza"`
	DataArrivo     string   `json:"data_arrivo"`
}

type TripResponse struct {
	ID             uuid.UUID        `json:"id"`
	OrdiniIds      []string         `json:"ordini_ids"`
	TargaMotrice   string           `json:"targa_motrice"`
	TargaRimorchio string           `json:"targa_rimorchio"`
	AutistaID      string           `json:"autista_id"`
	Autista        *DriverResponse  `json:"autista"`
	VettoreID      string           `json:"vettore_id"`
	Vettore        *CarrierResponse `json:"vettore"`
	GarageID       string           `json:"garage_id"`
	Garage         *GarageResponse  `json:"garage"`
	Segmenti       []TripSegmentDTO `json:"segmenti"`
	KmTotali       float64          `json:"km_totali"`
	CostoStimato   float64          `json:"costo_stimato"`
	Stato          string           `json:"stato"`
	Note           string           `json:"note"`
	DataPartenza   string           `json:"data_partenza"`
	DataArrivo     string           `json:"data_arrivo"`
	CreatedAt      time.Time        `json:"created_at"`
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
	TotalOrders    int64               `json:"total_orders"`
	Pianificabili  int64               `json:"pianificabili"`
	InViaggio      int64               `json:"in_viaggio"`
	Chiusi         int64               `json:"chiusi"`
	Fatturati      int64               `json:"fatturati"`
	TotalCustomers int64               `json:"total_customers"`
	TotalVehicles  int64               `json:"total_vehicles"`
	TotalDrivers   int64               `json:"total_drivers"`
	TotalRevenue   float64             `json:"total_revenue"`
	MonthlyTrend   []MonthlyOrderTrend `json:"monthly_trend"`
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
	TargaMotrice    string            `json:"targa_motrice"`
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
	GpsLive         bool              `json:"gps_live"`
	GpsSpeedKmh     float64           `json:"gps_speed_kmh"`
	GpsHeading      float64           `json:"gps_heading"`
	GpsTrackerUrl   string            `json:"gps_tracker_url"`
	GpsLastUpdate   string            `json:"gps_last_update"`
	GpsSource       string            `json:"gps_source"`
	LastTempCelsius *float64          `json:"last_temp_celsius"`
	LastTempAlert   bool              `json:"last_temp_alert"`
	Garage          *MapNamedPoint    `json:"garage"`
	WashStation     *MapNamedPoint    `json:"wash_station"`
}

type MapStats struct {
	InViaggio     int `json:"in_viaggio"`
	Pianificabili int `json:"pianificabili"`
	Chiusi        int `json:"chiusi"`
	GpsLive       int `json:"gps_live"`
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
