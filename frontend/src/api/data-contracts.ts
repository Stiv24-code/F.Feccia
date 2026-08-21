/* eslint-disable */
/* tslint:disable */
// @ts-nocheck
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */

export interface DtoAccessoryCostRequest {
  costo_default?: number;
  descrizione?: string;
  nome: string;
}

export interface DtoAccessoryCostResponse {
  active?: boolean;
  costo_default?: number;
  descrizione?: string;
  id?: string;
  nome?: string;
}

export interface DtoAccountingEntryListResponse {
  data?: DtoAccountingEntryResponse[];
  total?: number;
}

export interface DtoAccountingEntryRequest {
  codice: string;
  conto_contabile?: string;
  descrizione: string;
  iva_codice?: string;
  tipo?: "ricavo" | "costo";
}

export interface DtoAccountingEntryResponse {
  active?: boolean;
  codice?: string;
  conto_contabile?: string;
  created_at?: string;
  descrizione?: string;
  id?: string;
  iva_codice?: string;
  tipo?: string;
  updated_at?: string;
}

export interface DtoAuthUserResponse {
  active?: boolean;
  /**
   * CustomerID is only non-nil for RoleCliente accounts (self-registered
   * via /auth/register-cliente) — the Customer/anagrafica they're scoped
   * to. Distinct from the legacy unused ProfileID above.
   */
  customer_id?: string;
  email?: string;
  id?: number;
  name?: string;
  profile_id?: string;
  role?: string;
}

export interface DtoBankListResponse {
  data?: DtoBankResponse[];
  total?: number;
}

export interface DtoBankRequest {
  bic_swift?: string;
  citta?: string;
  iban_prefix?: string;
  indirizzo?: string;
  nome: string;
  note?: string;
}

export interface DtoBankResponse {
  active?: boolean;
  bic_swift?: string;
  citta?: string;
  created_at?: string;
  iban_prefix?: string;
  id?: string;
  indirizzo?: string;
  nome?: string;
  note?: string;
  updated_at?: string;
}

export interface DtoCarrierListResponse {
  data?: DtoCarrierResponse[];
  total?: number;
}

export interface DtoCarrierRequest {
  citta?: string;
  email?: string;
  indirizzo?: string;
  note?: string;
  partita_iva?: string;
  ragione_sociale: string;
  telefono?: string;
}

export interface DtoCarrierResponse {
  active?: boolean;
  citta?: string;
  created_at?: string;
  email?: string;
  id?: string;
  indirizzo?: string;
  note?: string;
  partita_iva?: string;
  ragione_sociale?: string;
  telefono?: string;
  updated_at?: string;
}

export interface DtoClientInboundOrderRequest {
  andata_ritorno?: boolean;
  categoria_trasporto?: string;
  cliente_id: string;
  costi_accessori?: Record<string, any>[];
  data_consegna?: string;
  data_ritiro?: string;
  destinazione_carico_id?: string;
  destinazione_scarico_id?: string;
  items?: DtoOrderItemRequestDTO[];
  kg?: number;
  note?: string;
  ora_consegna_a?: string;
  ora_consegna_da?: string;
  ora_ritiro_a?: string;
  ora_ritiro_da?: string;
  product?: string;
  rif_ordine_cliente?: string;
  servizi_accessori?: string[];
  tariffa?: number;
  tipo_tariffa?: string;
  tipologia?: string;
}

export interface DtoClientRegisterRequest {
  cap?: string;
  citta?: string;
  codice_fiscale?: string;
  email: string;
  indirizzo?: string;
  /**
   * Lat/Lng are optional, filled in from the Indirizzo geocoding search when
   * the address matched — never required, so a signup never fails just
   * because the address wasn't found on the map.
   */
  lat?: number;
  lng?: number;
  /** @minLength 1 */
  name: string;
  partita_iva?: string;
  /** @minLength 12 */
  password: string;
  provincia?: string;
  ragione_sociale: string;
  telefono?: string;
}

export interface DtoCountryListResponse {
  data?: DtoCountryResponse[];
  total?: number;
}

export interface DtoCountryRequest {
  eu?: boolean;
  iso2: string;
  iso3?: string;
  nome: string;
  valuta?: string;
}

export interface DtoCountryResponse {
  active?: boolean;
  created_at?: string;
  eu?: boolean;
  id?: string;
  iso2?: string;
  iso3?: string;
  nome?: string;
  updated_at?: string;
  valuta?: string;
}

export interface DtoCreateUserRequest {
  /**
   * @minLength 3
   * @maxLength 150
   */
  login: string;
  name?: string;
  /** @minLength 6 */
  password: string;
  role: "admin" | "amministrazione" | "planner" | "operatore";
}

export interface DtoCustomerDashboardCategoria {
  categoria?: string;
  ordini?: number;
}

export interface DtoCustomerDashboardDestination {
  fatturato?: number;
  nome?: string;
  ordini?: number;
}

export interface DtoCustomerDashboardKPI {
  fatturato_netto?: number;
  ordini_chiusi?: number;
  ordini_fatturati?: number;
  ordini_in_viaggio?: number;
  ordini_pianificabili?: number;
  ordini_totali?: number;
  tariffa_media?: number;
}

export interface DtoCustomerDashboardMonthly {
  fatturato?: number;
  mese?: string;
  ordini?: number;
}

export interface DtoCustomerDashboardResponse {
  customer?: DtoCustomerDashboardSummary;
  kpi?: DtoCustomerDashboardKPI;
  monthly_trend?: DtoCustomerDashboardMonthly[];
  per_categoria?: DtoCustomerDashboardCategoria[];
  per_tipologia?: DtoCustomerDashboardTipologia[];
  top_destinazioni?: DtoCustomerDashboardDestination[];
}

export interface DtoCustomerDashboardSummary {
  citta?: string;
  id?: string;
  partita_iva?: string;
  ragione_sociale?: string;
}

export interface DtoCustomerDashboardTipologia {
  ordini?: number;
  tipologia?: string;
}

export interface DtoCustomerListResponse {
  data?: DtoCustomerResponse[];
  total?: number;
}

export interface DtoCustomerRequest {
  cap?: string;
  citta?: string;
  codice_fiscale?: string;
  condizioni_pagamento?: string;
  email?: string;
  indirizzo?: string;
  /**
   * Lat/Lng optional — see models.Customer for why this differs from
   * Destination/Garage/WashStation's mandatory Posizione.
   */
  lat?: number;
  lng?: number;
  nazione?: string;
  note?: string;
  partita_iva?: string;
  pec?: string;
  provincia?: string;
  ragione_sociale: string;
  richiede_rif_ordine?: boolean;
  telefono?: string;
}

export interface DtoCustomerResponse {
  active?: boolean;
  cap?: string;
  citta?: string;
  codice_fiscale?: string;
  condizioni_pagamento?: string;
  created_at?: string;
  email?: string;
  id?: string;
  indirizzo?: string;
  lat?: number;
  lng?: number;
  nazione?: string;
  note?: string;
  partita_iva?: string;
  pec?: string;
  provincia?: string;
  ragione_sociale?: string;
  richiede_rif_ordine?: boolean;
  telefono?: string;
  updated_at?: string;
}

export interface DtoDashboardStatsResponse {
  chiusi?: number;
  fatturati?: number;
  in_viaggio?: number;
  monthly_trend?: DtoMonthlyOrderTrend[];
  pianificabili?: number;
  total_customers?: number;
  total_drivers?: number;
  total_motrici?: number;
  total_orders?: number;
  total_revenue?: number;
  total_semirimorchi?: number;
}

export interface DtoDestinationListResponse {
  data?: DtoDestinationResponse[];
  total?: number;
}

export interface DtoDestinationRequest {
  active?: boolean;
  cap?: string;
  citta?: string;
  indirizzo?: string;
  lat: number;
  lng: number;
  nazione?: string;
  nome: string;
  note?: string;
  provincia?: string;
  vincoli_scarico?: string;
}

export interface DtoDestinationResponse {
  active?: boolean;
  cap?: string;
  citta?: string;
  created_at?: string;
  id?: string;
  indirizzo?: string;
  lat?: number;
  lng?: number;
  nazione?: string;
  nome?: string;
  note?: string;
  provincia?: string;
  updated_at?: string;
  vincoli_scarico?: string;
}

export interface DtoDriverAvailabilityResponse {
  active?: boolean;
  codice_fiscale?: string;
  cognome?: string;
  created_at?: string;
  disponibilita?: string;
  email?: string;
  id?: string;
  motivo_indisponibilita?: string;
  nome?: string;
  note?: string;
  patente?: (
    | "AM"
    | "A1"
    | "A2"
    | "A"
    | "B1"
    | "B"
    | "BE"
    | "C1"
    | "C1E"
    | "C"
    | "CE"
    | "D1"
    | "D1E"
    | "D"
    | "DE"
    | "CQC"
    | "ADR"
  )[];
  prossime_ferie_a?: string;
  /**
   * ProssimeFerieDa/A are the next upcoming motivo=ferie unavailability
   * window (nil if none), computed on read only by DriverService.List —
   * not persisted, not populated by Create/Update.
   */
  prossime_ferie_da?: string;
  scadenza_patente?: string;
  telefono?: string;
  updated_at?: string;
}

export interface DtoDriverListResponse {
  data?: DtoDriverResponse[];
  total?: number;
}

export interface DtoDriverRequest {
  codice_fiscale?: string;
  cognome: string;
  email?: string;
  nome: string;
  note?: string;
  patente?: (
    | "AM"
    | "A1"
    | "A2"
    | "A"
    | "B1"
    | "B"
    | "BE"
    | "C1"
    | "C1E"
    | "C"
    | "CE"
    | "D1"
    | "D1E"
    | "D"
    | "DE"
    | "CQC"
    | "ADR"
  )[];
  scadenza_patente?: string;
  telefono?: string;
}

export interface DtoDriverResponse {
  active?: boolean;
  codice_fiscale?: string;
  cognome?: string;
  created_at?: string;
  email?: string;
  id?: string;
  nome?: string;
  note?: string;
  patente?: (
    | "AM"
    | "A1"
    | "A2"
    | "A"
    | "B1"
    | "B"
    | "BE"
    | "C1"
    | "C1E"
    | "C"
    | "CE"
    | "D1"
    | "D1E"
    | "D"
    | "DE"
    | "CQC"
    | "ADR"
  )[];
  prossime_ferie_a?: string;
  /**
   * ProssimeFerieDa/A are the next upcoming motivo=ferie unavailability
   * window (nil if none), computed on read only by DriverService.List —
   * not persisted, not populated by Create/Update.
   */
  prossime_ferie_da?: string;
  scadenza_patente?: string;
  telefono?: string;
  updated_at?: string;
}

export interface DtoDriverUnavailabilityRequest {
  autista_id: string;
  autista_nome?: string;
  data_a: string;
  data_da: string;
  motivo?: "ferie" | "malattia" | "permesso" | "altro";
  note?: string;
}

export interface DtoDriverUnavailabilityResponse {
  autista_id?: string;
  autista_nome?: string;
  created_at?: string;
  data_a?: string;
  data_da?: string;
  id?: string;
  motivo?: "ferie" | "malattia" | "permesso" | "altro";
  note?: string;
}

export interface DtoGarageListResponse {
  data?: DtoGarageResponse[];
  total?: number;
}

export interface DtoGarageRequest {
  active?: boolean;
  citta?: string;
  indirizzo?: string;
  lat: number;
  lng: number;
  nome: string;
  note?: string;
}

export interface DtoGarageResponse {
  active?: boolean;
  citta?: string;
  created_at?: string;
  id?: string;
  indirizzo?: string;
  lat?: number;
  lng?: number;
  nome?: string;
  note?: string;
  updated_at?: string;
}

export interface DtoGeocodeResultDTO {
  cap?: string;
  citta?: string;
  indirizzo?: string;
  label?: string;
  lat?: number;
  lng?: number;
  nazione?: string;
  provincia?: string;
}

export interface DtoInboundConfigResponse {
  accept_mode?: string;
  backend?: string;
  mailbox_ready?: boolean;
  pdf_ready?: boolean;
  scrape_interval_min?: number;
  smtp_ready?: boolean;
  subject_filter?: string;
  test_recipient?: string;
  vision_ready?: boolean;
}

export interface DtoInboundOrderActionResponse {
  mail?: string;
  order?: DtoInboundOrderResponse;
}

export interface DtoInboundOrderDraftDTO {
  client?: string;
  delivery_date?: string;
  delivery_place?: string;
  kg?: number;
  load_date?: string;
  load_place?: string;
  notes?: string;
  product?: string;
  rate?: string;
  received_at?: string;
  ref?: string;
  sender_email?: string;
  source?: string;
  status?: string;
  template_id?: string;
}

export interface DtoInboundOrderRequest {
  client: string;
  /**
   * ClienteID is set internally by CreateMyInboundOrder (client portal) —
   * never trusted from external request bodies for the staff-facing
   * /inbound-orders endpoint, same posture as OrderRequest.ClienteID on /me/orders.
   */
  cliente_id?: string;
  delivery_date?: string;
  delivery_place?: string;
  kg?: number;
  load_date?: string;
  load_place?: string;
  notes?: string;
  portal?: boolean;
  product?: string;
  rate?: string;
  received_at?: string;
  ref?: string;
  sender_email?: string;
  source?: "seed" | "mail" | "pdf" | "portal";
  status?: "pending" | "accepted" | "modify";
  template_id?: string;
}

export interface DtoInboundOrderResponse {
  client?: string;
  cliente_id?: string;
  created_at?: string;
  delivery_date?: string;
  delivery_place?: string;
  id?: string;
  kg?: number;
  load_date?: string;
  load_place?: string;
  notes?: string;
  portal?: boolean;
  product?: string;
  rate?: string;
  received_at?: string;
  ref?: string;
  sender_email?: string;
  source?: string;
  status?: string;
  template_id?: string;
  updated_at?: string;
}

export interface DtoInboundScrapeResponse {
  added?: number;
  scanned?: number;
}

export interface DtoInvoiceFinalizeResult {
  ok?: boolean;
  pdf_archived?: boolean;
  pdf_s3_key?: string;
}

export interface DtoInvoiceLineDTO {
  descrizione?: string;
  iva_codice?: string;
  ordine_id?: string;
  peso?: number;
  prodotto?: string;
  quantita?: number;
  tariffa?: number;
  totale?: number;
}

export interface DtoInvoicePDFURLResult {
  expires_in?: number;
  retain_until?: string;
  url?: string;
}

export interface DtoInvoiceRequest {
  cliente_id: string;
  condizioni_pagamento?: string;
  costi_accessori?: Record<string, any>[];
  data_fattura?: string;
  data_scadenza?: string;
  righe?: DtoInvoiceLineDTO[];
  totale?: number;
  totale_imponibile?: number;
  totale_iva?: number;
}

export interface DtoInvoiceResponse {
  cliente?: DtoCustomerResponse;
  cliente_id?: string;
  condizioni_pagamento?: string;
  costi_accessori?: Record<string, any>[];
  created_at?: string;
  data_fattura?: string;
  data_scadenza?: string;
  id?: string;
  note?: string;
  numero?: string;
  pdf_retain_until?: string;
  pdf_s3_key?: string;
  pdf_uploaded_at?: string;
  righe?: DtoInvoiceLineDTO[];
  stato?: string;
  tipo?: string;
  totale?: number;
  totale_imponibile?: number;
  totale_iva?: number;
}

export interface DtoLoginRequest {
  email: string;
  password: string;
}

export interface DtoLoginResult {
  access_token?: string;
  expires_in?: number;
  token_type?: string;
  user?: DtoAuthUserResponse;
}

export interface DtoMapNamedPoint {
  lat?: number;
  lng?: number;
  nome?: string;
}

export interface DtoMapPOI {
  id?: string;
  lat?: number;
  lng?: number;
  nome?: string;
}

export interface DtoMapPoint {
  lat?: number;
  lng?: number;
}

export interface DtoMapRoute {
  autista?: DtoDriverResponse;
  carico?: DtoMapPoint;
  cliente?: DtoCustomerResponse;
  current_position?: DtoMapPoint;
  data_consegna?: string;
  data_ritiro?: string;
  distance_km?: number;
  duration_hours?: number;
  eta_hours?: number;
  garage?: DtoMapNamedPoint;
  id?: string;
  motrice?: DtoMotriceResponse;
  progress?: number;
  progressivo?: string;
  remaining_km?: number;
  road_points?: DtoMapPoint[];
  scarico?: DtoMapPoint;
  stato?: string;
  tariffa?: number;
  tipologia?: string;
  wash_station?: DtoMapNamedPoint;
}

export interface DtoMapStats {
  chiusi?: number;
  in_viaggio?: number;
  pianificabili?: number;
}

export interface DtoMapTripsResponse {
  garages?: DtoMapNamedPoint[];
  poi?: DtoMapPOI[];
  routes?: DtoMapRoute[];
  stats?: DtoMapStats;
  wash_stations?: DtoMapNamedPoint[];
}

export interface DtoMonthlyOrderTrend {
  mese?: string;
  ordini?: number;
  totale?: number;
}

export interface DtoMotriceAvailabilityResponse {
  active?: boolean;
  anno?: number;
  created_at?: string;
  disponibilita?: string;
  id?: string;
  marca?: string;
  modello?: string;
  note?: string;
  portata_kg?: number;
  targa?: string;
  updated_at?: string;
}

export interface DtoMotriceListResponse {
  data?: DtoMotriceResponse[];
  total?: number;
}

export interface DtoMotriceRequest {
  anno?: number;
  marca?: string;
  modello?: string;
  note?: string;
  portata_kg?: number;
  targa: string;
}

export interface DtoMotriceResponse {
  active?: boolean;
  anno?: number;
  created_at?: string;
  id?: string;
  marca?: string;
  modello?: string;
  note?: string;
  portata_kg?: number;
  targa?: string;
  updated_at?: string;
}

export interface DtoOKResult {
  ok?: boolean;
}

export interface DtoOrderAssignRequest {
  autista_id?: string;
  garage_id?: string;
  motrice_id?: string;
  /**
   * RouteWaypoints: la sequenza scelta dal manager tra le alternative
   * proposte (POST /orders/{id}/route-alternatives). Opzionale — se
   * assente l'ordine viene comunque assegnato, semplicemente senza un
   * OrderRoute calcolato. Il server ricalcola sempre la geometria via ORS,
   * non si fida di quella eventualmente mandata dal client.
   */
  route_waypoints?: DtoRouteWaypointDTO[];
  semirimorchio_id?: string;
  vettore_id?: string;
  wash_station_id?: string;
}

export interface DtoOrderItemRequestDTO {
  peso?: number;
  prodotto_id?: string;
  quantita?: number;
}

export interface DtoOrderItemResponseDTO {
  peso?: number;
  prodotto?: DtoProductResponse;
  quantita?: number;
}

export interface DtoOrderRequest {
  andata_ritorno?: boolean;
  categoria_trasporto?: string;
  cliente_id: string;
  costi_accessori?: Record<string, any>[];
  data_consegna?: string;
  data_ritiro?: string;
  destinazione_carico_id?: string;
  destinazione_scarico_id?: string;
  items?: DtoOrderItemRequestDTO[];
  note?: string;
  ora_consegna_a?: string;
  ora_consegna_da?: string;
  ora_ritiro_a?: string;
  ora_ritiro_da?: string;
  rif_ordine_cliente?: string;
  servizi_accessori?: string[];
  tariffa?: number;
  tipo_tariffa?: string;
  tipologia?: string;
}

export interface DtoOrderResponse {
  andata_ritorno?: boolean;
  autista?: DtoDriverResponse;
  autista_id?: string;
  categoria_trasporto?: string;
  cliente?: DtoCustomerResponse;
  cliente_id?: string;
  costi_accessori?: Record<string, any>[];
  created_at?: string;
  data_consegna?: string;
  data_ritiro?: string;
  destinazione_carico?: DtoDestinationResponse;
  destinazione_carico_id?: string;
  destinazione_scarico?: DtoDestinationResponse;
  destinazione_scarico_id?: string;
  fattura_id?: string;
  garage?: DtoGarageResponse;
  garage_id?: string;
  id?: string;
  items?: DtoOrderItemResponseDTO[];
  motrice?: DtoMotriceResponse;
  motrice_id?: string;
  note?: string;
  ora_consegna_a?: string;
  ora_consegna_da?: string;
  ora_ritiro_a?: string;
  ora_ritiro_da?: string;
  progressivo?: string;
  rif_ordine_cliente?: string;
  route?: DtoRouteResponseDTO;
  route_id?: string;
  semirimorchio?: DtoSemirimorchioResponse;
  semirimorchio_id?: string;
  servizi_accessori?: string[];
  stato?: "PIANIFICABILE" | "PIANIFICATO" | "VIAGGIO" | "CHIUSO" | "SCARTATO";
  tariffa?: number;
  tipo_tariffa?: string;
  tipologia?: string;
  updated_at?: string;
  vettore?: DtoCarrierResponse;
  vettore_id?: string;
  viaggio_id?: string;
  wash_station?: DtoWashStationResponse;
  wash_station_id?: string;
}

export interface DtoOrderReturnSuggestion {
  order?: DtoOrderResponse;
  reasons?: string[];
  score?: number;
}

export interface DtoOrderReturnSuggestionsResponse {
  candidates?: DtoOrderReturnSuggestion[];
  count?: number;
  source_order?: DtoOrderSourceSummary;
}

export interface DtoOrderRouteAlternativesRequest {
  garage_id?: string;
  wash_station_id?: string;
}

export interface DtoOrderRouteAlternativesResponse {
  alternatives?: DtoRouteAlternativeDTO[];
}

export interface DtoOrderRouteUpdateRequest {
  /** @minItems 2 */
  waypoints: DtoRouteWaypointDTO[];
}

export interface DtoOrderSourceSummary {
  cliente?: DtoCustomerResponse;
  data_consegna?: string;
  destinazione_scarico?: DtoDestinationResponse;
  id?: string;
  progressivo?: string;
}

export interface DtoPatchUserRequest {
  active?: boolean;
  name?: string;
  profile_id?: string;
}

export interface DtoPdfExtractedValueDTO {
  confidence?: number;
  method?: string;
  value?: string;
}

export interface DtoPdfExtractionResponse {
  extraction?: Record<string, DtoPdfExtractedValueDTO>;
  order?: DtoInboundOrderDraftDTO;
  template?: DtoPdfTemplateRef;
  values?: Record<string, string>;
}

export interface DtoPdfRenderBlockDTO {
  bounds_norm?: Record<string, number>;
  text?: string;
}

export interface DtoPdfRenderPageDTO {
  blocks?: DtoPdfRenderBlockDTO[];
  height?: number;
  image_b64?: string;
  page_num?: number;
  width?: number;
}

export interface DtoPdfRenderResponse {
  filename?: string;
  page_count?: number;
  pages?: DtoPdfRenderPageDTO[];
}

export interface DtoPdfTemplateFieldDTO {
  h?: number;
  id?: string;
  label?: string;
  page?: number;
  target: string;
  w?: number;
  x?: number;
  y?: number;
}

export interface DtoPdfTemplateMatchResponse {
  match?: DtoPdfTemplateResponse;
}

export interface DtoPdfTemplateRef {
  id?: string;
  name?: string;
}

export interface DtoPdfTemplateRequest {
  /** Client is the default client name for orders imported with this template. */
  client?: string;
  fields?: DtoPdfTemplateFieldDTO[];
  name: string;
  /**
   * Senders holds full addresses or "@domain" patterns used to preselect
   * the template from a mail sender.
   */
  senders?: string[];
}

export interface DtoPdfTemplateResponse {
  client?: string;
  created_at?: string;
  fields?: DtoPdfTemplateFieldDTO[];
  id?: string;
  name?: string;
  senders?: string[];
  updated_at?: string;
}

export interface DtoPriceListItemAddResult {
  item_id?: string;
  items_count?: number;
  ok?: boolean;
}

export interface DtoPriceListItemDeleteResult {
  items_count?: number;
  ok?: boolean;
}

export interface DtoPriceListItemRequestDTO {
  destinazione_carico_id?: string;
  destinazione_scarico_id?: string;
  item_id?: string;
  minimo_tassabile?: number;
  perc_adeguamento_carburante?: number;
  prodotto_id?: string;
  range_peso_max?: number;
  range_peso_min?: number;
  tariffa?: number;
  tipo_tariffa?: string;
  tipo_trasporto?: string;
  unita_peso?: string;
}

export interface DtoPriceListItemResponseDTO {
  destinazione_carico?: DtoDestinationResponse;
  destinazione_scarico?: DtoDestinationResponse;
  item_id?: string;
  minimo_tassabile?: number;
  perc_adeguamento_carburante?: number;
  prodotto?: DtoProductResponse;
  range_peso_max?: number;
  range_peso_min?: number;
  tariffa?: number;
  tipo_tariffa?: string;
  tipo_trasporto?: string;
  unita_peso?: string;
}

export interface DtoPriceListItemUpdateResult {
  item_id?: string;
  ok?: boolean;
}

export interface DtoPriceListRequest {
  cliente_id: string;
  data_fine?: string;
  data_inizio?: string;
  items?: DtoPriceListItemRequestDTO[];
  note?: string;
}

export interface DtoPriceListResponse {
  active?: boolean;
  cliente?: DtoCustomerResponse;
  cliente_id?: string;
  created_at?: string;
  data_fine?: string;
  data_inizio?: string;
  id?: string;
  in_uso?: boolean;
  items?: DtoPriceListItemResponseDTO[];
  note?: string;
}

export interface DtoPriceListUpdateResult {
  duplicated?: boolean;
  new_id?: string;
  ok?: boolean;
}

export interface DtoProductListResponse {
  data?: DtoProductResponse[];
  total?: number;
}

export interface DtoProductRequest {
  codice: string;
  descrizione: string;
  note?: string;
  unita_misura?: string;
}

export interface DtoProductResponse {
  active?: boolean;
  codice?: string;
  created_at?: string;
  descrizione?: string;
  id?: string;
  note?: string;
  unita_misura?: string;
  updated_at?: string;
}

export interface DtoRecomputeSegmentsResult {
  km_totali?: number;
  ok?: boolean;
  segmenti_count?: number;
}

export interface DtoRegisterClientResult {
  email?: string;
  message?: string;
  verified?: boolean;
}

export interface DtoRegisterRequest {
  customer_id?: string;
  email: string;
  /** @minLength 1 */
  name: string;
  /** @minLength 12 */
  password: string;
  role: "admin" | "amministrazione" | "planner" | "operatore" | "cliente";
}

export interface DtoResendVerificationRequest {
  email: string;
}

export interface DtoRouteAlternativeDTO {
  distance_km?: number;
  duration_min?: number;
  points?: number[][];
  waypoints?: DtoRouteWaypointResponseDTO[];
}

export interface DtoRouteResponseDTO {
  distance_km?: number;
  duration_min?: number;
  edited_manually?: boolean;
  id?: string;
  points?: number[][];
  waypoints?: DtoRouteWaypointResponseDTO[];
}

export interface DtoRouteWaypointDTO {
  ref_id: string;
  tipo: "garage" | "destinazione" | "wash_station";
}

export interface DtoRouteWaypointResponseDTO {
  lat?: number;
  lng?: number;
  nome?: string;
  ref_id?: string;
  tipo?: string;
}

export interface DtoSemirimorchioAvailabilityResponse {
  active?: boolean;
  created_at?: string;
  disponibilita?: string;
  id?: string;
  note?: string;
  portata_kg?: number;
  scompartature?: number;
  targa?: string;
  tipo?: string;
  updated_at?: string;
}

export interface DtoSemirimorchioListResponse {
  data?: DtoSemirimorchioResponse[];
  total?: number;
}

export interface DtoSemirimorchioRequest {
  note?: string;
  portata_kg?: number;
  scompartature?: number;
  targa: string;
  tipo?: string;
}

export interface DtoSemirimorchioResponse {
  active?: boolean;
  created_at?: string;
  id?: string;
  note?: string;
  portata_kg?: number;
  scompartature?: number;
  targa?: string;
  tipo?: string;
  updated_at?: string;
}

export interface DtoTariffLookupResult {
  found?: boolean;
  item_id?: string;
  listino_id?: string;
  minimo_tassabile?: number;
  perc_adeguamento_carburante?: number;
  score?: number;
  tariffa?: number;
  tariffa_base?: number;
  tipo_tariffa?: string;
}

export interface DtoTransportCategoryRequest {
  descrizione?: string;
  nome: string;
}

export interface DtoTransportCategoryResponse {
  active?: boolean;
  descrizione?: string;
  id?: string;
  nome?: string;
}

export interface DtoTripDetailResponse {
  autista?: DtoDriverResponse;
  autista_id?: string;
  costo_stimato?: number;
  created_at?: string;
  data_arrivo?: string;
  data_partenza?: string;
  garage?: DtoGarageResponse;
  garage_id?: string;
  id?: string;
  km_totali?: number;
  motrice?: DtoMotriceResponse;
  motrice_id?: string;
  note?: string;
  ordini?: DtoOrderResponse[];
  ordini_ids?: string[];
  segmenti?: DtoTripSegmentDTO[];
  semirimorchio?: DtoSemirimorchioResponse;
  semirimorchio_id?: string;
  stato?: string;
  vettore?: DtoCarrierResponse;
  vettore_id?: string;
}

export interface DtoTripRequest {
  autista_id?: string;
  data_arrivo?: string;
  data_partenza?: string;
  garage_id?: string;
  motrice_id?: string;
  note?: string;
  ordini_ids?: string[];
  semirimorchio_id?: string;
  vettore_id?: string;
}

export interface DtoTripResponse {
  autista?: DtoDriverResponse;
  autista_id?: string;
  costo_stimato?: number;
  created_at?: string;
  data_arrivo?: string;
  data_partenza?: string;
  garage?: DtoGarageResponse;
  garage_id?: string;
  id?: string;
  km_totali?: number;
  motrice?: DtoMotriceResponse;
  motrice_id?: string;
  note?: string;
  ordini_ids?: string[];
  segmenti?: DtoTripSegmentDTO[];
  semirimorchio?: DtoSemirimorchioResponse;
  semirimorchio_id?: string;
  stato?: string;
  vettore?: DtoCarrierResponse;
  vettore_id?: string;
}

export interface DtoTripSegmentDTO {
  destinazione_lat?: number;
  destinazione_lng?: number;
  destinazione_nome?: string;
  km?: number;
  ordine?: number;
  ordine_id?: string;
  origine_lat?: number;
  origine_lng?: number;
  origine_nome?: string;
  tempo_stimato_min?: number;
  tipo?: string;
}

export interface DtoUpdateUserRequest {
  /**
   * @minLength 3
   * @maxLength 150
   */
  login: string;
  name?: string;
  password?: string;
  /**
   * "cliente" is accepted here only so re-saving a cliente account's
   * name/active status (role unchanged) doesn't fail validation — the
   * admin UI never offers switching a role TO cliente via this edit path
   * (that always goes through /auth/register instead, which also asks for
   * the Customer to link).
   */
  role: "admin" | "amministrazione" | "planner" | "operatore" | "cliente";
}

export interface DtoVehicleTypeRequest {
  descrizione?: string;
  nome: string;
}

export interface DtoVehicleTypeResponse {
  active?: boolean;
  descrizione?: string;
  id?: string;
  nome?: string;
}

export interface DtoVerifyEmailRequest {
  token: string;
}

export interface DtoWashStationListResponse {
  data?: DtoWashStationResponse[];
  total?: number;
}

export interface DtoWashStationRequest {
  active?: boolean;
  citta?: string;
  indirizzo?: string;
  lat: number;
  lng: number;
  nome: string;
  note?: string;
  orario_a?: string;
  orario_da?: string;
  tipo?: string;
}

export interface DtoWashStationResponse {
  active?: boolean;
  citta?: string;
  created_at?: string;
  id?: string;
  indirizzo?: string;
  lat?: number;
  lng?: number;
  nome?: string;
  note?: string;
  orario_a?: string;
  orario_da?: string;
  tipo?: string;
  updated_at?: string;
}

export interface ModelsUser {
  active?: boolean;
  created_at?: string;
  /**
   * CustomerID links a RoleCliente account to its Customer/anagrafica —
   * always nil for staff roles. Set once, at registration (see
   * AuthService.RegisterClient), never reassigned.
   */
  customer_id?: string;
  /**
   * Email verification (self-service client registration only — see
   * AuthService.RegisterClient). VerificationToken is non-nil exactly
   * while a confirmation is outstanding; Login only refuses access when a
   * token was actually issued and never confirmed, so accounts created
   * before this feature existed (VerificationToken always nil) are never
   * retroactively locked out.
   */
  email_verified_at?: string;
  id?: number;
  last_login_at?: string;
  /**
   * @minLength 3
   * @maxLength 150
   */
  login: string;
  /** @maxLength 255 */
  name?: string;
  password_hash: string;
  role: "admin" | "amministrazione" | "planner" | "operatore" | "cliente";
  updated_at?: string;
}
