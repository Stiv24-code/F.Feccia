// Tipi domain TMS — mirror dei Pydantic backend (issue #27, fase iniziale).
//
// Convertito a mano per ora; quando #43 (OpenAPI generator) sarà fatto,
// sostituiremo questo file con tipi auto-generati da `openapi-typescript`.
//
// Le pagine consumano tipi parziali (es. Order, Customer) tramite
// `import type { Order } from '@/types/domain'`. La conversione progressiva
// delle pagine a TS sostituirà gli `any` impliciti con questi tipi.
//
// OrderStato fa già eccezione: il backend espone `stato` come vero enum
// swagger (vedi dto.OrderResponse in backend/internal/dto/dto.go), quindi lo
// deriviamo dal client generato invece di duplicare i 5 valori a mano — un
// `swag init` + `yarn generate:api` che aggiunge/toglie uno stato aggiorna
// anche questo tipo automaticamente.

import type { DtoOrderResponse } from '@/api/data-contracts';

export type UserRole = 'admin' | 'amministrazione' | 'planner' | 'operatore';

export interface User {
  id: string;
  email: string;
  name: string;
  role: UserRole | string;
  active: boolean;
}

export interface Customer {
  id: string;
  ragione_sociale: string;
  indirizzo?: string;
  citta?: string;
  cap?: string;
  provincia?: string;
  nazione?: string;
  partita_iva?: string;
  codice_fiscale?: string;
  telefono?: string;
  email?: string;
  pec?: string;
  condizioni_pagamento?: string;
  note?: string;
  richiede_rif_ordine?: boolean;
  active: boolean;
  created_at?: string;
}

export interface Destination {
  id: string;
  nome: string;
  indirizzo?: string;
  citta?: string;
  cap?: string;
  provincia?: string;
  nazione?: string;
  vincoli_scarico?: string;
  note?: string;
  active: boolean;
  created_at?: string;
}

export interface Garage {
  id: string;
  nome: string;
  indirizzo?: string;
  citta?: string;
  lat?: number | null;
  lng?: number | null;
  note?: string;
  active: boolean;
  created_at?: string;
}

export interface WashStation {
  id: string;
  nome: string;
  tipo?: string;
  indirizzo?: string;
  citta?: string;
  lat?: number | null;
  lng?: number | null;
  note?: string;
  active: boolean;
  created_at?: string;
}

export interface Vehicle {
  id: string;
  targa: string;
  tipo_veicolo?: string;
  marca?: string;
  modello?: string;
  anno?: number;
  scompartature?: number;
  portata_kg?: number;
  note?: string;
  gps_tracker_url?: string;
  gps_tracker_tipo?: string;
  last_lat?: number;
  last_lng?: number;
  last_speed_kmh?: number;
  last_heading?: number;
  last_gps_update?: string;
  gps_active?: boolean;
  active: boolean;
}

export interface Driver {
  id: string;
  nome: string;
  cognome: string;
  codice_fiscale?: string;
  patente?: string;
  scadenza_patente?: string | null;
  telefono?: string;
  email?: string;
  note?: string;
  active: boolean;
}

export interface Carrier {
  id: string;
  ragione_sociale: string;
  partita_iva?: string;
  indirizzo?: string;
  citta?: string;
  telefono?: string;
  email?: string;
  active: boolean;
}

export interface Product {
  id: string;
  codice: string;
  descrizione: string;
  unita_misura?: string;
  active: boolean;
}

export interface OrderItem {
  prodotto?: Product;
  quantita?: number;
  peso?: number;
}

export type OrderStato = NonNullable<DtoOrderResponse['stato']>;

// Riferimenti anagrafica (cliente, destinazioni, garage, autista, vettore,
// wash station) sono relazioni FK reali lato backend — Preloaded e annidate
// per intero nella response, non più snapshot piatti "*_nome". Gli `*_id`
// piatti restano (utili per URL/binding Select), l'oggetto annidato è nil
// finché il riferimento non è impostato.
export interface Order {
  id: string;
  progressivo?: string;
  cliente_id: string;
  cliente?: Customer;
  destinazione_carico_id?: string;
  destinazione_carico?: Destination;
  destinazione_scarico_id?: string;
  destinazione_scarico?: Destination;
  data_ritiro?: string;
  data_consegna?: string;
  tariffa?: number;
  tipo_tariffa?: string;
  tipologia?: string;
  categoria_trasporto?: string;
  rif_ordine_cliente?: string;
  andata_ritorno?: boolean;
  note?: string;
  items?: OrderItem[];
  servizi_accessori?: string[];
  costi_accessori?: Array<Record<string, unknown>>;
  stato: OrderStato;
  garage_id?: string;
  garage?: Garage;
  targa_motrice?: string;
  targa_rimorchio?: string;
  autista_id?: string;
  autista?: Driver;
  vettore_id?: string;
  vettore?: Carrier;
  wash_station_id?: string;
  wash_station?: WashStation;
  route_id?: string;
  route?: OrderRoute;
  viaggio_id?: string;
  fattura_id?: string;
  created_at?: string;
  updated_at?: string;
}

export interface RouteWaypoint {
  tipo: 'garage' | 'destinazione' | 'wash_station';
  ref_id: string;
  nome?: string;
  lat?: number;
  lng?: number;
}

export interface OrderRoute {
  id: string;
  waypoints?: RouteWaypoint[];
  points?: [number, number][];
  distance_km?: number;
  duration_min?: number;
  edited_manually?: boolean;
}

export interface RouteAlternative {
  waypoints?: RouteWaypoint[];
  points?: [number, number][];
  distance_km?: number;
  duration_min?: number;
}

export type TripStato = 'IN_CORSO' | 'COMPLETATO';

export interface Trip {
  id: string;
  ordini_ids?: string[];
  targa_motrice?: string;
  targa_rimorchio?: string;
  autista_id?: string;
  autista?: Driver;
  garage_id?: string;
  garage?: Garage;
  km_totali?: number;
  costo_stimato?: number;
  stato: TripStato | string;
  data_partenza?: string;
  data_arrivo?: string;
  created_at?: string;
}

export type InvoiceStato = 'PROFORMA' | 'DEFINITIVA';

export interface InvoiceLine {
  ordine_id?: string;
  descrizione?: string;
  prodotto?: string;
  peso?: number;
  quantita?: number;
  tariffa?: number;
  totale?: number;
  iva_codice?: string;
}

export interface Invoice {
  id: string;
  numero?: string;
  cliente_id: string;
  cliente?: Customer;
  data_fattura?: string;
  data_scadenza?: string;
  righe?: InvoiceLine[];
  totale_imponibile?: number;
  totale_iva?: number;
  totale?: number;
  stato: InvoiceStato | string;
  tipo?: string;
  created_at?: string;
}

export interface PriceListItem {
  item_id?: string;
  prodotto?: Product;
  destinazione_carico?: Destination;
  destinazione_scarico?: Destination;
  tariffa?: number;
  tipo_tariffa?: string;
  range_peso_min?: number;
  range_peso_max?: number;
  minimo_tassabile?: number;
  perc_adeguamento_carburante?: number;
  unita_peso?: string;
  tipo_trasporto?: string;
}

export interface PriceList {
  id: string;
  cliente_id: string;
  cliente?: Customer;
  data_inizio: string;
  data_fine: string;
  items?: PriceListItem[];
  note?: string;
  in_uso?: boolean;
  active: boolean;
  created_at?: string;
}

export interface AuthLoginResponse {
  access_token: string;
  token_type: 'Bearer';
  expires_in: number;
  user: User;
}
