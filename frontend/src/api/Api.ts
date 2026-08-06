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

import {
  DtoAccessoryCostRequest,
  DtoAccessoryCostResponse,
  DtoAccountingEntryRequest,
  DtoAccountingEntryResponse,
  DtoAuthUserResponse,
  DtoBankRequest,
  DtoBankResponse,
  DtoCarrierRequest,
  DtoCarrierResponse,
  DtoClientRegisterRequest,
  DtoCountryRequest,
  DtoCountryResponse,
  DtoCreateUserRequest,
  DtoCustomerDashboardResponse,
  DtoCustomerRequest,
  DtoCustomerResponse,
  DtoDashboardStatsResponse,
  DtoDestinationRequest,
  DtoDestinationResponse,
  DtoDriverAvailabilityResponse,
  DtoDriverRequest,
  DtoDriverResponse,
  DtoDriverUnavailabilityRequest,
  DtoDriverUnavailabilityResponse,
  DtoGPSHistoryResponse,
  DtoGPSLiveVehicle,
  DtoGPSUpdateResult,
  DtoGPSWebhookPayload,
  DtoGarageRequest,
  DtoGarageResponse,
  DtoGeocodeResultDTO,
  DtoInvoiceFinalizeResult,
  DtoInvoicePDFURLResult,
  DtoInvoiceRequest,
  DtoInvoiceResponse,
  DtoLoginRequest,
  DtoLoginResult,
  DtoMapTripsResponse,
  DtoOKResult,
  DtoOrderAssignRequest,
  DtoOrderRequest,
  DtoOrderResponse,
  DtoOrderReturnSuggestionsResponse,
  DtoOrderRouteAlternativesRequest,
  DtoOrderRouteAlternativesResponse,
  DtoOrderRouteUpdateRequest,
  DtoPatchUserRequest,
  DtoPriceListItemAddResult,
  DtoPriceListItemDeleteResult,
  DtoPriceListItemRequestDTO,
  DtoPriceListItemUpdateResult,
  DtoPriceListRequest,
  DtoPriceListResponse,
  DtoPriceListUpdateResult,
  DtoProductRequest,
  DtoProductResponse,
  DtoRecomputeSegmentsResult,
  DtoRegisterRequest,
  DtoTariffLookupResult,
  DtoTemperatureReadingResponse,
  DtoTemperatureThresholdsRequest,
  DtoTemperatureThresholdsResult,
  DtoTemperatureWebhookRequest,
  DtoTemperatureWebhookResult,
  DtoTransportCategoryRequest,
  DtoTransportCategoryResponse,
  DtoTripDetailResponse,
  DtoTripRequest,
  DtoTripResponse,
  DtoUpdateUserRequest,
  DtoVehicleAvailabilityResponse,
  DtoVehicleGPSUpdateRequest,
  DtoVehicleRequest,
  DtoVehicleResponse,
  DtoVehicleTypeRequest,
  DtoVehicleTypeResponse,
  DtoWashStationRequest,
  DtoWashStationResponse,
  ModelsUser,
} from "./data-contracts";
import { ContentType, HttpClient, RequestParams } from "./http-client";

export class Api<SecurityDataType = unknown> {
  http: HttpClient<SecurityDataType>;

  constructor(http: HttpClient<SecurityDataType>) {
    this.http = http;
  }

  /**
   * No description
   *
   * @tags Masterdata
   * @name V1AccessoryCostsList
   * @summary List accessory costs
   * @request GET:/api/v1/accessory-costs
   * @secure
   */
  v1AccessoryCostsList = (params: RequestParams = {}) =>
    this.http.request<DtoAccessoryCostResponse[], any>({
      path: `/api/v1/accessory-costs`,
      method: "GET",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Masterdata
   * @name V1AccessoryCostsCreate
   * @summary Create accessory cost
   * @request POST:/api/v1/accessory-costs
   * @secure
   */
  v1AccessoryCostsCreate = (
    item: DtoAccessoryCostRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoAccessoryCostResponse, Record<string, string>>({
      path: `/api/v1/accessory-costs`,
      method: "POST",
      body: item,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Anagrafiche
   * @name V1AccountingEntriesList
   * @summary List accounting entries
   * @request GET:/api/v1/accounting-entries
   * @secure
   */
  v1AccountingEntriesList = (
    query?: {
      /** Filter by codice/descrizione */
      search?: string;
      /** Filter by tipo (ricavo|costo) */
      tipo?: string;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoAccountingEntryResponse[], any>({
      path: `/api/v1/accounting-entries`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Anagrafiche
   * @name V1AccountingEntriesCreate
   * @summary Create accounting entry
   * @request POST:/api/v1/accounting-entries
   * @secure
   */
  v1AccountingEntriesCreate = (
    entry: DtoAccountingEntryRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoAccountingEntryResponse, Record<string, string>>({
      path: `/api/v1/accounting-entries`,
      method: "POST",
      body: entry,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Anagrafiche
   * @name V1AccountingEntriesUpdate
   * @summary Update accounting entry (full replace)
   * @request PUT:/api/v1/accounting-entries/{id}
   * @secure
   */
  v1AccountingEntriesUpdate = (
    id: string,
    entry: DtoAccountingEntryRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoAccountingEntryResponse, Record<string, string>>({
      path: `/api/v1/accounting-entries/${id}`,
      method: "PUT",
      body: entry,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Anagrafiche
   * @name V1AccountingEntriesDelete
   * @summary Delete accounting entry (logical, sets active=false)
   * @request DELETE:/api/v1/accounting-entries/{id}
   * @secure
   */
  v1AccountingEntriesDelete = (id: string, params: RequestParams = {}) =>
    this.http.request<void, Record<string, string>>({
      path: `/api/v1/accounting-entries/${id}`,
      method: "DELETE",
      secure: true,
      ...params,
    });
  /**
   * @description Full unpaginated roster with email/profile_id/active, mirrors admin.py's GET /admin/users
   *
   * @tags Admin Users
   * @name V1AdminUsersList
   * @summary List all users (frontend admin panel)
   * @request GET:/api/v1/admin/users
   * @secure
   */
  v1AdminUsersList = (params: RequestParams = {}) =>
    this.http.request<DtoAuthUserResponse[], any>({
      path: `/api/v1/admin/users`,
      method: "GET",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * @description Create a new user
   *
   * @tags Admin Users
   * @name V1AdminUsersCreate
   * @summary Create user
   * @request POST:/api/v1/admin/users
   * @secure
   */
  v1AdminUsersCreate = (
    user: DtoCreateUserRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<ModelsUser, Record<string, string>>({
      path: `/api/v1/admin/users`,
      method: "POST",
      body: user,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * @description Returns paginated list of users
   *
   * @tags Admin Users
   * @name V1AdminUsersListList
   * @summary List users
   * @request GET:/api/v1/admin/users-list
   * @secure
   */
  v1AdminUsersListList = (
    query?: {
      /** Page number (default 1) */
      page?: number;
      /** Items per page (default 20, max 100) */
      limit?: number;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<Record<string, any>, Record<string, string>>({
      path: `/api/v1/admin/users-list`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * @description Retrieve a single user by ID
   *
   * @tags Admin Users
   * @name V1AdminUsersDetail
   * @summary Get user by ID
   * @request GET:/api/v1/admin/users/{id}
   * @secure
   */
  v1AdminUsersDetail = (id: number, params: RequestParams = {}) =>
    this.http.request<ModelsUser, Record<string, string>>({
      path: `/api/v1/admin/users/${id}`,
      method: "GET",
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * @description Update existing user
   *
   * @tags Admin Users
   * @name V1AdminUsersUpdate
   * @summary Update user
   * @request PUT:/api/v1/admin/users/{id}
   * @secure
   */
  v1AdminUsersUpdate = (
    id: number,
    user: DtoUpdateUserRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<ModelsUser, Record<string, string>>({
      path: `/api/v1/admin/users/${id}`,
      method: "PUT",
      body: user,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * @description Delete user by ID
   *
   * @tags Admin Users
   * @name V1AdminUsersDelete
   * @summary Delete user
   * @request DELETE:/api/v1/admin/users/{id}
   * @secure
   */
  v1AdminUsersDelete = (id: number, params: RequestParams = {}) =>
    this.http.request<void, Record<string, string>>({
      path: `/api/v1/admin/users/${id}`,
      method: "DELETE",
      secure: true,
      type: ContentType.Json,
      ...params,
    });
  /**
   * @description Updates name/active only, mirrors admin.py's PATCH /admin/users/{id}
   *
   * @tags Admin Users
   * @name V1AdminUsersPartialUpdate
   * @summary Partially update a user (frontend admin panel)
   * @request PATCH:/api/v1/admin/users/{id}
   * @secure
   */
  v1AdminUsersPartialUpdate = (
    id: number,
    user: DtoPatchUserRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoAuthUserResponse, Record<string, string>>({
      path: `/api/v1/admin/users/${id}`,
      method: "PATCH",
      body: user,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * @description Authenticate user; access token in body, refresh token as httpOnly cookie
   *
   * @tags Auth
   * @name V1AuthLoginCreate
   * @summary User login
   * @request POST:/api/v1/auth/login
   */
  v1AuthLoginCreate = (
    credentials: DtoLoginRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoLoginResult, Record<string, string>>({
      path: `/api/v1/auth/login`,
      method: "POST",
      body: credentials,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * @description Clears the refresh cookie; best-effort, works without a valid token
   *
   * @tags Auth
   * @name V1AuthLogoutCreate
   * @summary Logout
   * @request POST:/api/v1/auth/logout
   */
  v1AuthLogoutCreate = (params: RequestParams = {}) =>
    this.http.request<Record<string, boolean>, any>({
      path: `/api/v1/auth/logout`,
      method: "POST",
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Auth
   * @name V1AuthMeList
   * @summary Current user
   * @request GET:/api/v1/auth/me
   * @secure
   */
  v1AuthMeList = (params: RequestParams = {}) =>
    this.http.request<DtoAuthUserResponse, Record<string, string>>({
      path: `/api/v1/auth/me`,
      method: "GET",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * @description Exchange the httpOnly refresh cookie for a new access token (and rotate the cookie)
   *
   * @tags Auth
   * @name V1AuthRefreshCreate
   * @summary Refresh tokens
   * @request POST:/api/v1/auth/refresh
   */
  v1AuthRefreshCreate = (params: RequestParams = {}) =>
    this.http.request<DtoLoginResult, Record<string, string>>({
      path: `/api/v1/auth/refresh`,
      method: "POST",
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Auth
   * @name V1AuthRegisterCreate
   * @summary Register a new user (admin-only)
   * @request POST:/api/v1/auth/register
   * @secure
   */
  v1AuthRegisterCreate = (
    user: DtoRegisterRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoAuthUserResponse, Record<string, string>>({
      path: `/api/v1/auth/register`,
      method: "POST",
      body: user,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * @description Creates a Customer (anagrafica) + a "cliente" account atomically, then logs it in immediately (no approval step) — access token in body, refresh token as httpOnly cookie, same as Login.
   *
   * @tags Auth
   * @name V1AuthRegisterClienteCreate
   * @summary Self-service client registration (public)
   * @request POST:/api/v1/auth/register-cliente
   */
  v1AuthRegisterClienteCreate = (
    registration: DtoClientRegisterRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoLoginResult, Record<string, string>>({
      path: `/api/v1/auth/register-cliente`,
      method: "POST",
      body: registration,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Availability
   * @name V1AvailabilityDriversList
   * @summary Driver availability for a date range
   * @request GET:/api/v1/availability/drivers
   * @secure
   */
  v1AvailabilityDriversList = (
    query: {
      /** Range start (YYYY-MM-DD) */
      data_da: string;
      /** Range end (YYYY-MM-DD) */
      data_a: string;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoDriverAvailabilityResponse[], any>({
      path: `/api/v1/availability/drivers`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Availability
   * @name V1AvailabilityVehiclesList
   * @summary Vehicle availability for a date range
   * @request GET:/api/v1/availability/vehicles
   * @secure
   */
  v1AvailabilityVehiclesList = (
    query: {
      /** Range start (YYYY-MM-DD) */
      data_da: string;
      /** Range end (YYYY-MM-DD) */
      data_a: string;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoVehicleAvailabilityResponse[], any>({
      path: `/api/v1/availability/vehicles`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Anagrafiche
   * @name V1BanksList
   * @summary List banks
   * @request GET:/api/v1/banks
   * @secure
   */
  v1BanksList = (
    query?: {
      /** Filter by nome/bic_swift */
      search?: string;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoBankResponse[], any>({
      path: `/api/v1/banks`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Anagrafiche
   * @name V1BanksCreate
   * @summary Create bank
   * @request POST:/api/v1/banks
   * @secure
   */
  v1BanksCreate = (bank: DtoBankRequest, params: RequestParams = {}) =>
    this.http.request<DtoBankResponse, Record<string, string>>({
      path: `/api/v1/banks`,
      method: "POST",
      body: bank,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Anagrafiche
   * @name V1BanksUpdate
   * @summary Update bank (full replace)
   * @request PUT:/api/v1/banks/{id}
   * @secure
   */
  v1BanksUpdate = (
    id: string,
    bank: DtoBankRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoBankResponse, Record<string, string>>({
      path: `/api/v1/banks/${id}`,
      method: "PUT",
      body: bank,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Anagrafiche
   * @name V1BanksDelete
   * @summary Delete bank (logical, sets active=false)
   * @request DELETE:/api/v1/banks/{id}
   * @secure
   */
  v1BanksDelete = (id: string, params: RequestParams = {}) =>
    this.http.request<void, Record<string, string>>({
      path: `/api/v1/banks/${id}`,
      method: "DELETE",
      secure: true,
      ...params,
    });
  /**
   * No description
   *
   * @tags Carriers
   * @name V1CarriersList
   * @summary List carriers
   * @request GET:/api/v1/carriers
   * @secure
   */
  v1CarriersList = (
    query?: {
      /** Filter by ragione sociale */
      search?: string;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoCarrierResponse[], any>({
      path: `/api/v1/carriers`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Carriers
   * @name V1CarriersCreate
   * @summary Create carrier
   * @request POST:/api/v1/carriers
   * @secure
   */
  v1CarriersCreate = (carrier: DtoCarrierRequest, params: RequestParams = {}) =>
    this.http.request<DtoCarrierResponse, Record<string, string>>({
      path: `/api/v1/carriers`,
      method: "POST",
      body: carrier,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Carriers
   * @name V1CarriersUpdate
   * @summary Update carrier (full replace)
   * @request PUT:/api/v1/carriers/{id}
   * @secure
   */
  v1CarriersUpdate = (
    id: string,
    carrier: DtoCarrierRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoCarrierResponse, Record<string, string>>({
      path: `/api/v1/carriers/${id}`,
      method: "PUT",
      body: carrier,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Carriers
   * @name V1CarriersDelete
   * @summary Delete carrier (logical, sets active=false)
   * @request DELETE:/api/v1/carriers/{id}
   * @secure
   */
  v1CarriersDelete = (id: string, params: RequestParams = {}) =>
    this.http.request<void, Record<string, string>>({
      path: `/api/v1/carriers/${id}`,
      method: "DELETE",
      secure: true,
      ...params,
    });
  /**
   * No description
   *
   * @tags Anagrafiche
   * @name V1CountriesList
   * @summary List countries
   * @request GET:/api/v1/countries
   * @secure
   */
  v1CountriesList = (
    query?: {
      /** Filter by nome/iso2 */
      search?: string;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoCountryResponse[], any>({
      path: `/api/v1/countries`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Anagrafiche
   * @name V1CountriesCreate
   * @summary Create country
   * @request POST:/api/v1/countries
   * @secure
   */
  v1CountriesCreate = (
    country: DtoCountryRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoCountryResponse, Record<string, string>>({
      path: `/api/v1/countries`,
      method: "POST",
      body: country,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Anagrafiche
   * @name V1CountriesUpdate
   * @summary Update country (full replace)
   * @request PUT:/api/v1/countries/{id}
   * @secure
   */
  v1CountriesUpdate = (
    id: string,
    country: DtoCountryRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoCountryResponse, Record<string, string>>({
      path: `/api/v1/countries/${id}`,
      method: "PUT",
      body: country,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Anagrafiche
   * @name V1CountriesDelete
   * @summary Delete country (logical, sets active=false)
   * @request DELETE:/api/v1/countries/{id}
   * @secure
   */
  v1CountriesDelete = (id: string, params: RequestParams = {}) =>
    this.http.request<void, Record<string, string>>({
      path: `/api/v1/countries/${id}`,
      method: "DELETE",
      secure: true,
      ...params,
    });
  /**
   * @description Returns active customers, optionally filtered by ragione sociale
   *
   * @tags Customers
   * @name V1CustomersList
   * @summary List customers
   * @request GET:/api/v1/customers
   * @secure
   */
  v1CustomersList = (
    query?: {
      /** Filter by ragione sociale (substring, case-insensitive) */
      search?: string;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoCustomerResponse[], any>({
      path: `/api/v1/customers`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Customers
   * @name V1CustomersCreate
   * @summary Create customer
   * @request POST:/api/v1/customers
   * @secure
   */
  v1CustomersCreate = (
    customer: DtoCustomerRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoCustomerResponse, Record<string, string>>({
      path: `/api/v1/customers`,
      method: "POST",
      body: customer,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Customers
   * @name V1CustomersDetail
   * @summary Get customer by ID
   * @request GET:/api/v1/customers/{id}
   * @secure
   */
  v1CustomersDetail = (id: string, params: RequestParams = {}) =>
    this.http.request<DtoCustomerResponse, Record<string, string>>({
      path: `/api/v1/customers/${id}`,
      method: "GET",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Customers
   * @name V1CustomersUpdate
   * @summary Update customer (full replace)
   * @request PUT:/api/v1/customers/{id}
   * @secure
   */
  v1CustomersUpdate = (
    id: string,
    customer: DtoCustomerRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoCustomerResponse, Record<string, string>>({
      path: `/api/v1/customers/${id}`,
      method: "PUT",
      body: customer,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Customers
   * @name V1CustomersDelete
   * @summary Delete customer (logical, sets active=false)
   * @request DELETE:/api/v1/customers/{id}
   * @secure
   */
  v1CustomersDelete = (id: string, params: RequestParams = {}) =>
    this.http.request<void, Record<string, string>>({
      path: `/api/v1/customers/${id}`,
      method: "DELETE",
      secure: true,
      ...params,
    });
  /**
   * No description
   *
   * @tags Dashboard
   * @name V1DashboardCustomerDetail
   * @summary Per-customer commercial dashboard
   * @request GET:/api/v1/dashboard/customer/{customer_id}
   * @secure
   */
  v1DashboardCustomerDetail = (
    customerId: string,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoCustomerDashboardResponse, Record<string, string>>({
      path: `/api/v1/dashboard/customer/${customerId}`,
      method: "GET",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Dashboard
   * @name V1DashboardRecentOrdersList
   * @summary Recent orders (last 10)
   * @request GET:/api/v1/dashboard/recent-orders
   * @secure
   */
  v1DashboardRecentOrdersList = (params: RequestParams = {}) =>
    this.http.request<DtoOrderResponse[], any>({
      path: `/api/v1/dashboard/recent-orders`,
      method: "GET",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Dashboard
   * @name V1DashboardStatsList
   * @summary Global dashboard KPIs
   * @request GET:/api/v1/dashboard/stats
   * @secure
   */
  v1DashboardStatsList = (params: RequestParams = {}) =>
    this.http.request<DtoDashboardStatsResponse, any>({
      path: `/api/v1/dashboard/stats`,
      method: "GET",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Destinations
   * @name V1DestinationsList
   * @summary List destinations
   * @request GET:/api/v1/destinations
   * @secure
   */
  v1DestinationsList = (
    query?: {
      /** Filter by nome (substring, case-insensitive) */
      search?: string;
      /** Include logically deleted (active=false) destinations */
      include_inactive?: boolean;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoDestinationResponse[], any>({
      path: `/api/v1/destinations`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Destinations
   * @name V1DestinationsCreate
   * @summary Create destination
   * @request POST:/api/v1/destinations
   * @secure
   */
  v1DestinationsCreate = (
    destination: DtoDestinationRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoDestinationResponse, Record<string, string>>({
      path: `/api/v1/destinations`,
      method: "POST",
      body: destination,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Destinations
   * @name V1DestinationsDetail
   * @summary Get destination by ID
   * @request GET:/api/v1/destinations/{id}
   * @secure
   */
  v1DestinationsDetail = (id: string, params: RequestParams = {}) =>
    this.http.request<DtoDestinationResponse, Record<string, string>>({
      path: `/api/v1/destinations/${id}`,
      method: "GET",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Destinations
   * @name V1DestinationsUpdate
   * @summary Update destination (full replace)
   * @request PUT:/api/v1/destinations/{id}
   * @secure
   */
  v1DestinationsUpdate = (
    id: string,
    destination: DtoDestinationRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoDestinationResponse, Record<string, string>>({
      path: `/api/v1/destinations/${id}`,
      method: "PUT",
      body: destination,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Destinations
   * @name V1DestinationsDelete
   * @summary Delete destination (logical, sets active=false)
   * @request DELETE:/api/v1/destinations/{id}
   * @secure
   */
  v1DestinationsDelete = (id: string, params: RequestParams = {}) =>
    this.http.request<void, Record<string, string>>({
      path: `/api/v1/destinations/${id}`,
      method: "DELETE",
      secure: true,
      ...params,
    });
  /**
   * No description
   *
   * @tags DriverUnavailability
   * @name V1DriverUnavailabilityList
   * @summary List driver unavailability periods
   * @request GET:/api/v1/driver-unavailability
   * @secure
   */
  v1DriverUnavailabilityList = (
    query?: {
      /** Filter by driver ID (UUID) */
      autista_id?: string;
      /** Range start (overlap filter, requires data_a too) */
      data_da?: string;
      /** Range end (overlap filter, requires data_da too) */
      data_a?: string;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<
      DtoDriverUnavailabilityResponse[],
      Record<string, string>
    >({
      path: `/api/v1/driver-unavailability`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags DriverUnavailability
   * @name V1DriverUnavailabilityCreate
   * @summary Create driver unavailability period
   * @request POST:/api/v1/driver-unavailability
   * @secure
   */
  v1DriverUnavailabilityCreate = (
    item: DtoDriverUnavailabilityRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoDriverUnavailabilityResponse, Record<string, string>>({
      path: `/api/v1/driver-unavailability`,
      method: "POST",
      body: item,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags DriverUnavailability
   * @name V1DriverUnavailabilityDelete
   * @summary Delete driver unavailability period (hard delete)
   * @request DELETE:/api/v1/driver-unavailability/{id}
   * @secure
   */
  v1DriverUnavailabilityDelete = (id: string, params: RequestParams = {}) =>
    this.http.request<Record<string, boolean>, Record<string, string>>({
      path: `/api/v1/driver-unavailability/${id}`,
      method: "DELETE",
      secure: true,
      ...params,
    });
  /**
   * No description
   *
   * @tags Drivers
   * @name V1DriversList
   * @summary List drivers
   * @request GET:/api/v1/drivers
   * @secure
   */
  v1DriversList = (
    query?: {
      /** Filter by nome/cognome */
      search?: string;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoDriverResponse[], any>({
      path: `/api/v1/drivers`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Drivers
   * @name V1DriversCreate
   * @summary Create driver
   * @request POST:/api/v1/drivers
   * @secure
   */
  v1DriversCreate = (driver: DtoDriverRequest, params: RequestParams = {}) =>
    this.http.request<DtoDriverResponse, Record<string, string>>({
      path: `/api/v1/drivers`,
      method: "POST",
      body: driver,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Drivers
   * @name V1DriversUpdate
   * @summary Update driver (full replace)
   * @request PUT:/api/v1/drivers/{id}
   * @secure
   */
  v1DriversUpdate = (
    id: string,
    driver: DtoDriverRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoDriverResponse, Record<string, string>>({
      path: `/api/v1/drivers/${id}`,
      method: "PUT",
      body: driver,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Drivers
   * @name V1DriversDelete
   * @summary Delete driver (logical, sets active=false)
   * @request DELETE:/api/v1/drivers/{id}
   * @secure
   */
  v1DriversDelete = (id: string, params: RequestParams = {}) =>
    this.http.request<void, Record<string, string>>({
      path: `/api/v1/drivers/${id}`,
      method: "DELETE",
      secure: true,
      ...params,
    });
  /**
   * No description
   *
   * @tags Export
   * @name V1ExportOrdersList
   * @summary Export orders to xlsx
   * @request GET:/api/v1/export/orders
   * @secure
   */
  v1ExportOrdersList = (
    query?: {
      /** Filter by stato */
      stato?: string;
      /** data_ritiro >= (YYYY-MM-DD) */
      data_da?: string;
      /** data_ritiro <= (YYYY-MM-DD) */
      data_a?: string;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<Blob, any>({
      path: `/api/v1/export/orders`,
      method: "GET",
      query: query,
      secure: true,
      ...params,
    });
  /**
   * No description
   *
   * @tags Garages
   * @name V1GaragesList
   * @summary List garages
   * @request GET:/api/v1/garages
   * @secure
   */
  v1GaragesList = (
    query?: {
      /** Include logically deleted (active=false) garages */
      include_inactive?: boolean;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoGarageResponse[], any>({
      path: `/api/v1/garages`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Garages
   * @name V1GaragesCreate
   * @summary Create garage
   * @request POST:/api/v1/garages
   * @secure
   */
  v1GaragesCreate = (garage: DtoGarageRequest, params: RequestParams = {}) =>
    this.http.request<DtoGarageResponse, Record<string, string>>({
      path: `/api/v1/garages`,
      method: "POST",
      body: garage,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Garages
   * @name V1GaragesUpdate
   * @summary Update garage (full replace)
   * @request PUT:/api/v1/garages/{id}
   * @secure
   */
  v1GaragesUpdate = (
    id: string,
    garage: DtoGarageRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoGarageResponse, Record<string, string>>({
      path: `/api/v1/garages/${id}`,
      method: "PUT",
      body: garage,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Garages
   * @name V1GaragesDelete
   * @summary Delete garage (logical, sets active=false)
   * @request DELETE:/api/v1/garages/{id}
   * @secure
   */
  v1GaragesDelete = (id: string, params: RequestParams = {}) =>
    this.http.request<void, Record<string, string>>({
      path: `/api/v1/garages/${id}`,
      method: "DELETE",
      secure: true,
      ...params,
    });
  /**
   * No description
   *
   * @tags Geocode
   * @name V1GeocodeSearchList
   * @summary Forward-geocode an address (Destination/Garage/WashStation forms)
   * @request GET:/api/v1/geocode/search
   * @secure
   */
  v1GeocodeSearchList = (
    query: {
      /** Free-text address/place to search */
      q: string;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoGeocodeResultDTO[], any>({
      path: `/api/v1/geocode/search`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Invoices
   * @name V1InvoicesList
   * @summary List invoices
   * @request GET:/api/v1/invoices
   * @secure
   */
  v1InvoicesList = (
    query?: {
      /** Filter by stato */
      stato?: string;
      /** Filter by cliente_id */
      cliente_id?: string;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoInvoiceResponse[], any>({
      path: `/api/v1/invoices`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Invoices
   * @name V1InvoicesCreate
   * @summary Create invoice (PROFORMA)
   * @request POST:/api/v1/invoices
   * @secure
   */
  v1InvoicesCreate = (invoice: DtoInvoiceRequest, params: RequestParams = {}) =>
    this.http.request<DtoInvoiceResponse, any>({
      path: `/api/v1/invoices`,
      method: "POST",
      body: invoice,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Invoices
   * @name V1InvoicesDetail
   * @summary Get invoice by ID
   * @request GET:/api/v1/invoices/{id}
   * @secure
   */
  v1InvoicesDetail = (id: string, params: RequestParams = {}) =>
    this.http.request<DtoInvoiceResponse, Record<string, string>>({
      path: `/api/v1/invoices/${id}`,
      method: "GET",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Invoices
   * @name V1InvoicesDelete
   * @summary Delete invoice (only PROFORMA, hard delete)
   * @request DELETE:/api/v1/invoices/{id}
   * @secure
   */
  v1InvoicesDelete = (id: string, params: RequestParams = {}) =>
    this.http.request<void, Record<string, string>>({
      path: `/api/v1/invoices/${id}`,
      method: "DELETE",
      secure: true,
      ...params,
    });
  /**
   * No description
   *
   * @tags Invoices
   * @name V1InvoicesFinalizePartialUpdate
   * @summary Finalize invoice (PROFORMA -> DEFINITIVA)
   * @request PATCH:/api/v1/invoices/{id}/finalize
   * @secure
   */
  v1InvoicesFinalizePartialUpdate = (id: string, params: RequestParams = {}) =>
    this.http.request<DtoInvoiceFinalizeResult, Record<string, string>>({
      path: `/api/v1/invoices/${id}/finalize`,
      method: "PATCH",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Invoices
   * @name V1InvoicesPdfList
   * @summary Invoice PDF (S3 archived copy if available, else generated on the fly)
   * @request GET:/api/v1/invoices/{id}/pdf
   * @secure
   */
  v1InvoicesPdfList = (id: string, params: RequestParams = {}) =>
    this.http.request<Blob, Record<string, string>>({
      path: `/api/v1/invoices/${id}/pdf`,
      method: "GET",
      secure: true,
      ...params,
    });
  /**
   * No description
   *
   * @tags Invoices
   * @name V1InvoicesPdfUrlList
   * @summary Presigned S3 URL for the invoice PDF (DEFINITIVA + archived only)
   * @request GET:/api/v1/invoices/{id}/pdf-url
   * @secure
   */
  v1InvoicesPdfUrlList = (id: string, params: RequestParams = {}) =>
    this.http.request<DtoInvoicePDFURLResult, Record<string, string>>({
      path: `/api/v1/invoices/${id}/pdf-url`,
      method: "GET",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Map
   * @name V1MapTripsList
   * @summary Live map: active trips, POI, garages, stats
   * @request GET:/api/v1/map/trips
   * @secure
   */
  v1MapTripsList = (params: RequestParams = {}) =>
    this.http.request<DtoMapTripsResponse, any>({
      path: `/api/v1/map/trips`,
      method: "GET",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Auth
   * @name V1MeAnagraficaList
   * @summary Get the logged-in client's own anagrafica
   * @request GET:/api/v1/me/anagrafica
   * @secure
   */
  v1MeAnagraficaList = (params: RequestParams = {}) =>
    this.http.request<DtoCustomerResponse, Record<string, string>>({
      path: `/api/v1/me/anagrafica`,
      method: "GET",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Auth
   * @name V1MeAnagraficaUpdate
   * @summary Update the logged-in client's own anagrafica (full replace)
   * @request PUT:/api/v1/me/anagrafica
   * @secure
   */
  v1MeAnagraficaUpdate = (
    customer: DtoCustomerRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoCustomerResponse, Record<string, string>>({
      path: `/api/v1/me/anagrafica`,
      method: "PUT",
      body: customer,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * @description Same shared destinations pool as staff — a client can add a new pickup/delivery address to pick from when creating its own orders, but (unlike staff) cannot update or delete existing ones.
   *
   * @tags Auth
   * @name V1MeDestinationsCreate
   * @summary Create destination as the logged-in client
   * @request POST:/api/v1/me/destinations
   * @secure
   */
  v1MeDestinationsCreate = (
    destination: DtoDestinationRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoDestinationResponse, Record<string, string>>({
      path: `/api/v1/me/destinations`,
      method: "POST",
      body: destination,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Auth
   * @name V1MeOrdersList
   * @summary List the logged-in client's own orders
   * @request GET:/api/v1/me/orders
   * @secure
   */
  v1MeOrdersList = (
    query?: {
      /** Filter by stato */
      stato?: string;
      /** data_ritiro >= data_da */
      data_da?: string;
      /** data_ritiro <= data_a */
      data_a?: string;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoOrderResponse[], Record<string, string>>({
      path: `/api/v1/me/orders`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * @description cliente_id in the body is ignored — the order is always created under the caller's own anagrafica.
   *
   * @tags Auth
   * @name V1MeOrdersCreate
   * @summary Create an order as the logged-in client
   * @request POST:/api/v1/me/orders
   * @secure
   */
  v1MeOrdersCreate = (order: DtoOrderRequest, params: RequestParams = {}) =>
    this.http.request<DtoOrderResponse, Record<string, string>>({
      path: `/api/v1/me/orders`,
      method: "POST",
      body: order,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Auth
   * @name V1MeOrdersDetail
   * @summary Get one of the logged-in client's own orders by ID
   * @request GET:/api/v1/me/orders/{id}
   * @secure
   */
  v1MeOrdersDetail = (id: string, params: RequestParams = {}) =>
    this.http.request<DtoOrderResponse, Record<string, string>>({
      path: `/api/v1/me/orders/${id}`,
      method: "GET",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Auth
   * @name V1MeOrdersDelete
   * @summary Delete one of the logged-in client's own orders (only PIANIFICABILE, hard delete)
   * @request DELETE:/api/v1/me/orders/{id}
   * @secure
   */
  v1MeOrdersDelete = (id: string, params: RequestParams = {}) =>
    this.http.request<void, Record<string, string>>({
      path: `/api/v1/me/orders/${id}`,
      method: "DELETE",
      secure: true,
      ...params,
    });
  /**
   * No description
   *
   * @tags Orders
   * @name V1OrdersList
   * @summary List orders
   * @request GET:/api/v1/orders
   * @secure
   */
  v1OrdersList = (
    query?: {
      /** Filter by stato */
      stato?: string;
      /** Filter by cliente_id */
      cliente_id?: string;
      /** data_ritiro >= data_da */
      data_da?: string;
      /** data_ritiro <= data_a */
      data_a?: string;
      /** Filter by cliente_nome/progressivo/rif_ordine_cliente/destinazioni */
      search?: string;
      /** Filter by tipologia */
      tipologia?: string;
      /** Max results (default 500) */
      limit?: number;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoOrderResponse[], any>({
      path: `/api/v1/orders`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Orders
   * @name V1OrdersCreate
   * @summary Create order
   * @request POST:/api/v1/orders
   * @secure
   */
  v1OrdersCreate = (order: DtoOrderRequest, params: RequestParams = {}) =>
    this.http.request<DtoOrderResponse, Record<string, string>>({
      path: `/api/v1/orders`,
      method: "POST",
      body: order,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Orders
   * @name V1OrdersDetail
   * @summary Get order by ID
   * @request GET:/api/v1/orders/{id}
   * @secure
   */
  v1OrdersDetail = (id: string, params: RequestParams = {}) =>
    this.http.request<DtoOrderResponse, Record<string, string>>({
      path: `/api/v1/orders/${id}`,
      method: "GET",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Orders
   * @name V1OrdersUpdate
   * @summary Update order (full replace of the create-able fields)
   * @request PUT:/api/v1/orders/{id}
   * @secure
   */
  v1OrdersUpdate = (
    id: string,
    order: DtoOrderRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoOrderResponse, Record<string, string>>({
      path: `/api/v1/orders/${id}`,
      method: "PUT",
      body: order,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Orders
   * @name V1OrdersDelete
   * @summary Delete order (only PIANIFICABILE, hard delete)
   * @request DELETE:/api/v1/orders/{id}
   * @secure
   */
  v1OrdersDelete = (id: string, params: RequestParams = {}) =>
    this.http.request<void, Record<string, string>>({
      path: `/api/v1/orders/${id}`,
      method: "DELETE",
      secure: true,
      ...params,
    });
  /**
   * No description
   *
   * @tags Orders
   * @name V1OrdersAssignPartialUpdate
   * @summary Assign order to a vehicle/driver/carrier (PIANIFICABILE -> PIANIFICATO)
   * @request PATCH:/api/v1/orders/{id}/assign
   * @secure
   */
  v1OrdersAssignPartialUpdate = (
    id: string,
    assign: DtoOrderAssignRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoOrderResponse, Record<string, string>>({
      path: `/api/v1/orders/${id}/assign`,
      method: "PATCH",
      body: assign,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Orders
   * @name V1OrdersClosePartialUpdate
   * @summary Close order (VIAGGIO -> CHIUSO)
   * @request PATCH:/api/v1/orders/{id}/close
   * @secure
   */
  v1OrdersClosePartialUpdate = (id: string, params: RequestParams = {}) =>
    this.http.request<DtoOrderResponse, Record<string, string>>({
      path: `/api/v1/orders/${id}/close`,
      method: "PATCH",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Orders
   * @name V1OrdersCmrPdfList
   * @summary CMR waybill PDF for an order
   * @request GET:/api/v1/orders/{id}/cmr/pdf
   * @secure
   */
  v1OrdersCmrPdfList = (id: string, params: RequestParams = {}) =>
    this.http.request<Blob, Record<string, string>>({
      path: `/api/v1/orders/${id}/cmr/pdf`,
      method: "GET",
      secure: true,
      ...params,
    });
  /**
   * No description
   *
   * @tags Orders
   * @name V1OrdersDiscardPartialUpdate
   * @summary Discard order (PIANIFICABILE|PIANIFICATO -> SCARTATO)
   * @request PATCH:/api/v1/orders/{id}/discard
   * @secure
   */
  v1OrdersDiscardPartialUpdate = (id: string, params: RequestParams = {}) =>
    this.http.request<DtoOrderResponse, Record<string, string>>({
      path: `/api/v1/orders/${id}/discard`,
      method: "PATCH",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Orders
   * @name V1OrdersReturnSuggestionsList
   * @summary Return-trip order suggestions
   * @request GET:/api/v1/orders/{id}/return-suggestions
   * @secure
   */
  v1OrdersReturnSuggestionsList = (
    id: string,
    query?: {
      /** Days after data_consegna to search (0-14, default 2) */
      max_days_gap?: number;
      /** Max candidates (1-100, default 20) */
      limit?: number;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<
      DtoOrderReturnSuggestionsResponse,
      Record<string, string>
    >({
      path: `/api/v1/orders/${id}/return-suggestions`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Orders
   * @name V1OrdersRoutePartialUpdate
   * @summary Recompute and persist an order's route for an edited waypoint sequence
   * @request PATCH:/api/v1/orders/{id}/route
   * @secure
   */
  v1OrdersRoutePartialUpdate = (
    id: string,
    body: DtoOrderRouteUpdateRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoOrderResponse, Record<string, string>>({
      path: `/api/v1/orders/${id}/route`,
      method: "PATCH",
      body: body,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * @description Ephemeral — nothing is persisted, the manager picks one and it travels in the Assign/UpdateRoute call.
   *
   * @tags Orders
   * @name V1OrdersRouteAlternativesCreate
   * @summary Compute up to 3 truck-aware route alternatives for an order
   * @request POST:/api/v1/orders/{id}/route-alternatives
   * @secure
   */
  v1OrdersRouteAlternativesCreate = (
    id: string,
    body: DtoOrderRouteAlternativesRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<
      DtoOrderRouteAlternativesResponse,
      Record<string, string>
    >({
      path: `/api/v1/orders/${id}/route-alternatives`,
      method: "POST",
      body: body,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Orders
   * @name V1OrdersStartPartialUpdate
   * @summary Start order (PIANIFICATO -> VIAGGIO)
   * @request PATCH:/api/v1/orders/{id}/start
   * @secure
   */
  v1OrdersStartPartialUpdate = (id: string, params: RequestParams = {}) =>
    this.http.request<DtoOrderResponse, Record<string, string>>({
      path: `/api/v1/orders/${id}/start`,
      method: "PATCH",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * @description Reverse of Assign: clears garage/mezzo/autista/vettore/wash_station and the computed route.
   *
   * @tags Orders
   * @name V1OrdersUnassignPartialUpdate
   * @summary Unassign order (PIANIFICATO -> PIANIFICABILE)
   * @request PATCH:/api/v1/orders/{id}/unassign
   * @secure
   */
  v1OrdersUnassignPartialUpdate = (id: string, params: RequestParams = {}) =>
    this.http.request<DtoOrderResponse, Record<string, string>>({
      path: `/api/v1/orders/${id}/unassign`,
      method: "PATCH",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags PriceLists
   * @name V1PricelistsList
   * @summary List pricelists
   * @request GET:/api/v1/pricelists
   * @secure
   */
  v1PricelistsList = (
    query?: {
      /** Filter by cliente_id */
      cliente_id?: string;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoPriceListResponse[], any>({
      path: `/api/v1/pricelists`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags PriceLists
   * @name V1PricelistsCreate
   * @summary Create pricelist
   * @request POST:/api/v1/pricelists
   * @secure
   */
  v1PricelistsCreate = (
    pricelist: DtoPriceListRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoPriceListResponse, any>({
      path: `/api/v1/pricelists`,
      method: "POST",
      body: pricelist,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags PriceLists
   * @name V1PricelistsLookupTariffList
   * @summary Lookup the best matching tariff for an order context
   * @request GET:/api/v1/pricelists/lookup-tariff
   * @secure
   */
  v1PricelistsLookupTariffList = (
    query: {
      /** Cliente ID */
      cliente_id: string;
      /** Destinazione carico ID */
      carico_id?: string;
      /** Destinazione scarico ID */
      scarico_id?: string;
      /** Prodotto ID */
      prodotto_id?: string;
      /** Peso (Kg) */
      peso?: number;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoTariffLookupResult, any>({
      path: `/api/v1/pricelists/lookup-tariff`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags PriceLists
   * @name V1PricelistsDetail
   * @summary Get pricelist by ID
   * @request GET:/api/v1/pricelists/{id}
   * @secure
   */
  v1PricelistsDetail = (id: string, params: RequestParams = {}) =>
    this.http.request<DtoPriceListResponse, Record<string, string>>({
      path: `/api/v1/pricelists/${id}`,
      method: "GET",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags PriceLists
   * @name V1PricelistsUpdate
   * @summary Update pricelist (duplicates if in_uso, else in-place)
   * @request PUT:/api/v1/pricelists/{id}
   * @secure
   */
  v1PricelistsUpdate = (
    id: string,
    pricelist: DtoPriceListRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoPriceListUpdateResult, Record<string, string>>({
      path: `/api/v1/pricelists/${id}`,
      method: "PUT",
      body: pricelist,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags PriceLists
   * @name V1PricelistsDelete
   * @summary Delete pricelist (logical, sets active=false)
   * @request DELETE:/api/v1/pricelists/{id}
   * @secure
   */
  v1PricelistsDelete = (id: string, params: RequestParams = {}) =>
    this.http.request<void, any>({
      path: `/api/v1/pricelists/${id}`,
      method: "DELETE",
      secure: true,
      ...params,
    });
  /**
   * No description
   *
   * @tags PriceLists
   * @name V1PricelistsItemsCreate
   * @summary Add a tariff rule to a pricelist
   * @request POST:/api/v1/pricelists/{id}/items
   * @secure
   */
  v1PricelistsItemsCreate = (
    id: string,
    item: DtoPriceListItemRequestDTO,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoPriceListItemAddResult, Record<string, string>>({
      path: `/api/v1/pricelists/${id}/items`,
      method: "POST",
      body: item,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags PriceLists
   * @name V1PricelistsItemsUpdate
   * @summary Update a tariff rule
   * @request PUT:/api/v1/pricelists/{id}/items/{item_id}
   * @secure
   */
  v1PricelistsItemsUpdate = (
    id: string,
    itemId: string,
    item: DtoPriceListItemRequestDTO,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoPriceListItemUpdateResult, Record<string, string>>({
      path: `/api/v1/pricelists/${id}/items/${itemId}`,
      method: "PUT",
      body: item,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags PriceLists
   * @name V1PricelistsItemsDelete
   * @summary Delete a tariff rule
   * @request DELETE:/api/v1/pricelists/{id}/items/{item_id}
   * @secure
   */
  v1PricelistsItemsDelete = (
    id: string,
    itemId: string,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoPriceListItemDeleteResult, Record<string, string>>({
      path: `/api/v1/pricelists/${id}/items/${itemId}`,
      method: "DELETE",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Products
   * @name V1ProductsList
   * @summary List products
   * @request GET:/api/v1/products
   * @secure
   */
  v1ProductsList = (
    query?: {
      /** Filter by codice/descrizione */
      search?: string;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoProductResponse[], any>({
      path: `/api/v1/products`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Products
   * @name V1ProductsCreate
   * @summary Create product
   * @request POST:/api/v1/products
   * @secure
   */
  v1ProductsCreate = (product: DtoProductRequest, params: RequestParams = {}) =>
    this.http.request<DtoProductResponse, Record<string, string>>({
      path: `/api/v1/products`,
      method: "POST",
      body: product,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Products
   * @name V1ProductsUpdate
   * @summary Update product (full replace)
   * @request PUT:/api/v1/products/{id}
   * @secure
   */
  v1ProductsUpdate = (
    id: string,
    product: DtoProductRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoProductResponse, Record<string, string>>({
      path: `/api/v1/products/${id}`,
      method: "PUT",
      body: product,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Products
   * @name V1ProductsDelete
   * @summary Delete product (logical, sets active=false)
   * @request DELETE:/api/v1/products/{id}
   * @secure
   */
  v1ProductsDelete = (id: string, params: RequestParams = {}) =>
    this.http.request<void, Record<string, string>>({
      path: `/api/v1/products/${id}`,
      method: "DELETE",
      secure: true,
      ...params,
    });
  /**
   * No description
   *
   * @tags Masterdata
   * @name V1TransportCategoriesList
   * @summary List transport categories
   * @request GET:/api/v1/transport-categories
   * @secure
   */
  v1TransportCategoriesList = (params: RequestParams = {}) =>
    this.http.request<DtoTransportCategoryResponse[], any>({
      path: `/api/v1/transport-categories`,
      method: "GET",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Masterdata
   * @name V1TransportCategoriesCreate
   * @summary Create transport category
   * @request POST:/api/v1/transport-categories
   * @secure
   */
  v1TransportCategoriesCreate = (
    item: DtoTransportCategoryRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoTransportCategoryResponse, Record<string, string>>({
      path: `/api/v1/transport-categories`,
      method: "POST",
      body: item,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Trips
   * @name V1TripsList
   * @summary List trips
   * @request GET:/api/v1/trips
   * @secure
   */
  v1TripsList = (
    query?: {
      /** Filter by stato */
      stato?: string;
      /** Max results (default 200) */
      limit?: number;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoTripResponse[], any>({
      path: `/api/v1/trips`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * @description Syncs any PIANIFICABILE orders in ordini_ids to PIANIFICATO and computes route segments via OSRM
   *
   * @tags Trips
   * @name V1TripsCreate
   * @summary Create trip
   * @request POST:/api/v1/trips
   * @secure
   */
  v1TripsCreate = (trip: DtoTripRequest, params: RequestParams = {}) =>
    this.http.request<DtoTripResponse, any>({
      path: `/api/v1/trips`,
      method: "POST",
      body: trip,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Trips
   * @name V1TripsDetail
   * @summary Get trip by ID (includes joined orders)
   * @request GET:/api/v1/trips/{id}
   * @secure
   */
  v1TripsDetail = (id: string, params: RequestParams = {}) =>
    this.http.request<DtoTripDetailResponse, Record<string, string>>({
      path: `/api/v1/trips/${id}`,
      method: "GET",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Trips
   * @name V1TripsAddOrderPartialUpdate
   * @summary Add an order to a trip
   * @request PATCH:/api/v1/trips/{id}/add-order
   * @secure
   */
  v1TripsAddOrderPartialUpdate = (
    id: string,
    query: {
      /** Order ID (UUID) */
      order_id: string;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoOKResult, Record<string, string>>({
      path: `/api/v1/trips/${id}/add-order`,
      method: "PATCH",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Trips
   * @name V1TripsCompletePartialUpdate
   * @summary Complete a trip (closes its VIAGGIO orders)
   * @request PATCH:/api/v1/trips/{id}/complete
   * @secure
   */
  v1TripsCompletePartialUpdate = (id: string, params: RequestParams = {}) =>
    this.http.request<DtoOKResult, Record<string, string>>({
      path: `/api/v1/trips/${id}/complete`,
      method: "PATCH",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Trips
   * @name V1TripsInstructionsPdfList
   * @summary Operational instructions PDF for a trip
   * @request GET:/api/v1/trips/{id}/instructions/pdf
   * @secure
   */
  v1TripsInstructionsPdfList = (id: string, params: RequestParams = {}) =>
    this.http.request<Blob, Record<string, string>>({
      path: `/api/v1/trips/${id}/instructions/pdf`,
      method: "GET",
      secure: true,
      ...params,
    });
  /**
   * No description
   *
   * @tags Trips
   * @name V1TripsRecomputeSegmentsCreate
   * @summary Recompute a trip's route segments
   * @request POST:/api/v1/trips/{id}/recompute-segments
   * @secure
   */
  v1TripsRecomputeSegmentsCreate = (id: string, params: RequestParams = {}) =>
    this.http.request<DtoRecomputeSegmentsResult, Record<string, string>>({
      path: `/api/v1/trips/${id}/recompute-segments`,
      method: "POST",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Trips
   * @name V1TripsStartPartialUpdate
   * @summary Start a trip (PIANIFICATO -> IN_CORSO, starts its PIANIFICATO orders to VIAGGIO)
   * @request PATCH:/api/v1/trips/{id}/start
   * @secure
   */
  v1TripsStartPartialUpdate = (id: string, params: RequestParams = {}) =>
    this.http.request<DtoOKResult, Record<string, string>>({
      path: `/api/v1/trips/${id}/start`,
      method: "PATCH",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Masterdata
   * @name V1VehicleTypesList
   * @summary List vehicle types
   * @request GET:/api/v1/vehicle-types
   * @secure
   */
  v1VehicleTypesList = (params: RequestParams = {}) =>
    this.http.request<DtoVehicleTypeResponse[], any>({
      path: `/api/v1/vehicle-types`,
      method: "GET",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Masterdata
   * @name V1VehicleTypesCreate
   * @summary Create vehicle type
   * @request POST:/api/v1/vehicle-types
   * @secure
   */
  v1VehicleTypesCreate = (
    item: DtoVehicleTypeRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoVehicleTypeResponse, Record<string, string>>({
      path: `/api/v1/vehicle-types`,
      method: "POST",
      body: item,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Vehicles
   * @name V1VehiclesList
   * @summary List vehicles
   * @request GET:/api/v1/vehicles
   * @secure
   */
  v1VehiclesList = (
    query?: {
      /** Filter by targa */
      search?: string;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoVehicleResponse[], any>({
      path: `/api/v1/vehicles`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Vehicles
   * @name V1VehiclesCreate
   * @summary Create vehicle
   * @request POST:/api/v1/vehicles
   * @secure
   */
  v1VehiclesCreate = (vehicle: DtoVehicleRequest, params: RequestParams = {}) =>
    this.http.request<DtoVehicleResponse, Record<string, string>>({
      path: `/api/v1/vehicles`,
      method: "POST",
      body: vehicle,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Vehicles
   * @name V1VehiclesGpsLiveList
   * @summary Live GPS positions for all active vehicles
   * @request GET:/api/v1/vehicles/gps-live
   * @secure
   */
  v1VehiclesGpsLiveList = (params: RequestParams = {}) =>
    this.http.request<DtoGPSLiveVehicle[], any>({
      path: `/api/v1/vehicles/gps-live`,
      method: "GET",
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Vehicles
   * @name V1VehiclesGpsPositionByPlateCreate
   * @summary Update vehicle GPS position by targa
   * @request POST:/api/v1/vehicles/gps-position-by-plate/{targa}
   * @secure
   */
  v1VehiclesGpsPositionByPlateCreate = (
    targa: string,
    position: DtoVehicleGPSUpdateRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoGPSUpdateResult, Record<string, string>>({
      path: `/api/v1/vehicles/gps-position-by-plate/${targa}`,
      method: "POST",
      body: position,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Vehicles
   * @name V1VehiclesUpdate
   * @summary Update vehicle (full replace of the create-able fields)
   * @request PUT:/api/v1/vehicles/{id}
   * @secure
   */
  v1VehiclesUpdate = (
    id: string,
    vehicle: DtoVehicleRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoVehicleResponse, Record<string, string>>({
      path: `/api/v1/vehicles/${id}`,
      method: "PUT",
      body: vehicle,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Vehicles
   * @name V1VehiclesDelete
   * @summary Delete vehicle (logical, sets active=false)
   * @request DELETE:/api/v1/vehicles/{id}
   * @secure
   */
  v1VehiclesDelete = (id: string, params: RequestParams = {}) =>
    this.http.request<void, any>({
      path: `/api/v1/vehicles/${id}`,
      method: "DELETE",
      secure: true,
      ...params,
    });
  /**
   * No description
   *
   * @tags Vehicles
   * @name V1VehiclesGpsHistoryList
   * @summary GPS history for a vehicle
   * @request GET:/api/v1/vehicles/{id}/gps-history
   * @secure
   */
  v1VehiclesGpsHistoryList = (
    id: string,
    query?: {
      /** Max results (default 100) */
      limit?: number;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoGPSHistoryResponse[], any>({
      path: `/api/v1/vehicles/${id}/gps-history`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Vehicles
   * @name V1VehiclesGpsPositionCreate
   * @summary Update vehicle GPS position (by id or targa)
   * @request POST:/api/v1/vehicles/{id}/gps-position
   * @secure
   */
  v1VehiclesGpsPositionCreate = (
    id: string,
    position: DtoVehicleGPSUpdateRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoGPSUpdateResult, Record<string, string>>({
      path: `/api/v1/vehicles/${id}/gps-position`,
      method: "POST",
      body: position,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Vehicles
   * @name V1VehiclesTemperatureList
   * @summary Temperature history for a vehicle
   * @request GET:/api/v1/vehicles/{id}/temperature
   * @secure
   */
  v1VehiclesTemperatureList = (
    id: string,
    query?: {
      /** Max results (default 200) */
      limit?: number;
      /** Only out-of-range readings */
      only_alerts?: boolean;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoTemperatureReadingResponse[], any>({
      path: `/api/v1/vehicles/${id}/temperature`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Vehicles
   * @name V1VehiclesTemperatureThresholdsPartialUpdate
   * @summary Set temperature thresholds for a vehicle
   * @request PATCH:/api/v1/vehicles/{id}/temperature-thresholds
   * @secure
   */
  v1VehiclesTemperatureThresholdsPartialUpdate = (
    id: string,
    thresholds: DtoTemperatureThresholdsRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoTemperatureThresholdsResult, Record<string, string>>({
      path: `/api/v1/vehicles/${id}/temperature-thresholds`,
      method: "PATCH",
      body: thresholds,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags WashStations
   * @name V1WashStationsList
   * @summary List wash stations
   * @request GET:/api/v1/wash-stations
   * @secure
   */
  v1WashStationsList = (
    query?: {
      /** Include logically deleted (active=false) wash stations */
      include_inactive?: boolean;
    },
    params: RequestParams = {},
  ) =>
    this.http.request<DtoWashStationResponse[], any>({
      path: `/api/v1/wash-stations`,
      method: "GET",
      query: query,
      secure: true,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags WashStations
   * @name V1WashStationsCreate
   * @summary Create wash station
   * @request POST:/api/v1/wash-stations
   * @secure
   */
  v1WashStationsCreate = (
    washStation: DtoWashStationRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoWashStationResponse, Record<string, string>>({
      path: `/api/v1/wash-stations`,
      method: "POST",
      body: washStation,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags WashStations
   * @name V1WashStationsUpdate
   * @summary Update wash station (full replace)
   * @request PUT:/api/v1/wash-stations/{id}
   * @secure
   */
  v1WashStationsUpdate = (
    id: string,
    washStation: DtoWashStationRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoWashStationResponse, Record<string, string>>({
      path: `/api/v1/wash-stations/${id}`,
      method: "PUT",
      body: washStation,
      secure: true,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags WashStations
   * @name V1WashStationsDelete
   * @summary Delete wash station (logical, sets active=false)
   * @request DELETE:/api/v1/wash-stations/{id}
   * @secure
   */
  v1WashStationsDelete = (id: string, params: RequestParams = {}) =>
    this.http.request<void, Record<string, string>>({
      path: `/api/v1/wash-stations/${id}`,
      method: "DELETE",
      secure: true,
      ...params,
    });
  /**
   * No description
   *
   * @tags Vehicles
   * @name V1WebhooksGpsCreate
   * @summary GPS provider webhook ingestion
   * @request POST:/api/v1/webhooks/gps/{vendor}
   */
  v1WebhooksGpsCreate = (
    vendor: string,
    payload: DtoGPSWebhookPayload,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoGPSUpdateResult, Record<string, string>>({
      path: `/api/v1/webhooks/gps/${vendor}`,
      method: "POST",
      body: payload,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
  /**
   * No description
   *
   * @tags Vehicles
   * @name V1WebhooksTemperatureCreate
   * @summary Temperature sensor webhook ingestion
   * @request POST:/api/v1/webhooks/temperature/{vendor}
   */
  v1WebhooksTemperatureCreate = (
    vendor: string,
    payload: DtoTemperatureWebhookRequest,
    params: RequestParams = {},
  ) =>
    this.http.request<DtoTemperatureWebhookResult, Record<string, string>>({
      path: `/api/v1/webhooks/temperature/${vendor}`,
      method: "POST",
      body: payload,
      type: ContentType.Json,
      format: "json",
      ...params,
    });
}
