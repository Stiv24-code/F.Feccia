// Layer RTK Query: caching/invalidation sopra il client tipizzato generato
// da swagger (src/lib/apiClient.ts). Ogni endpoint usa `queryFn` invece di
// `query` per riusare i metodi già tipizzati di `apiClient` (path, body e
// response provengono da src/api/data-contracts.ts, generato dal backend) —
// niente duplicazione di path/tipi qui, RTK Query fa solo caching/tag.
import { createApi, fakeBaseQuery } from '@reduxjs/toolkit/query/react';
import { apiClient } from '@/lib/apiClient';
import { toQueryResult, toQueryError, toPagedQueryResult } from './rtkQueryHelpers';

// Argomenti comuni degli endpoint di elenco Anagrafiche paginati: ricerca +
// pagina/dimensione pagina (default lato backend 1/20 — vedi pkg/utils.PageParams).
export interface PagedListArgs {
  search?: string;
  page?: number;
  limit?: number;
}
// Risultato normalizzato (vedi toPagedQueryResult): sempre {items, total},
// mai undefined, indipendentemente da come swagger ha tipizzato l'envelope.
export interface PagedResult<T> {
  items: T[];
  total: number;
}
import type {
  DtoAccountingEntryRequest,
  DtoAccountingEntryResponse,
  DtoAuthUserResponse,
  DtoBankRequest,
  DtoBankResponse,
  DtoCarrierRequest,
  DtoCarrierResponse,
  DtoCountryRequest,
  DtoCountryResponse,
  DtoCustomerRequest,
  DtoCustomerResponse,
  DtoCustomerDashboardResponse,
  DtoDashboardStatsResponse,
  DtoDestinationRequest,
  DtoDestinationResponse,
  DtoDriverRequest,
  DtoDriverResponse,
  DtoDriverUnavailabilityRequest,
  DtoDriverUnavailabilityResponse,
  DtoGarageRequest,
  DtoGarageResponse,
  DtoInboundOrderResponse,
  DtoMotriceRequest,
  DtoMotriceResponse,
  DtoOrderRequest,
  DtoOrderResponse,
  DtoProductRequest,
  DtoProductResponse,
  DtoRegisterRequest,
  DtoSemirimorchioRequest,
  DtoSemirimorchioResponse,
  DtoTripResponse,
  DtoUpdateUserRequest,
  DtoWashStationRequest,
  DtoWashStationResponse,
} from '@/api/data-contracts';

export const appApi = createApi({
  reducerPath: 'appApi',
  baseQuery: fakeBaseQuery(),
  tagTypes: [
    'Customer', 'Dashboard', 'Destination', 'Motrice', 'Semirimorchio', 'Driver', 'DriverUnavailability',
    'Carrier', 'Product', 'Garage', 'WashStation', 'Country', 'Bank', 'AccountingEntry', 'AdminUser',
    'MyAnagrafica', 'MyOrder', 'MyInboundOrder',
  ],
  endpoints: (builder) => ({
    getDashboardStats: builder.query<DtoDashboardStatsResponse, void>({
      queryFn: () => toQueryResult(apiClient.v1DashboardStatsList()),
      providesTags: ['Dashboard'],
    }),
    getRecentOrders: builder.query<DtoOrderResponse[], void>({
      queryFn: () => toQueryResult(apiClient.v1DashboardRecentOrdersList()),
      providesTags: ['Dashboard'],
    }),
    getCustomerDashboard: builder.query<DtoCustomerDashboardResponse, string>({
      queryFn: (customerId) => toQueryResult(apiClient.v1DashboardCustomerDetail(customerId)),
      providesTags: ['Dashboard'],
    }),

    getCustomers: builder.query<PagedResult<DtoCustomerResponse>, PagedListArgs | void>({
      queryFn: (args: PagedListArgs = {}) => toPagedQueryResult(apiClient.v1CustomersList({ search: args.search || undefined, page: args.page, limit: args.limit })),
      providesTags: ['Customer'],
    }),
    createCustomer: builder.mutation<DtoCustomerResponse, DtoCustomerRequest>({
      queryFn: (body) => toQueryResult(apiClient.v1CustomersCreate(body)),
      invalidatesTags: ['Customer'],
    }),
    updateCustomer: builder.mutation<DtoCustomerResponse, { id: string; body: DtoCustomerRequest }>({
      queryFn: ({ id, body }) => toQueryResult(apiClient.v1CustomersUpdate(id, body)),
      invalidatesTags: ['Customer'],
    }),
    deleteCustomer: builder.mutation<void, string>({
      queryFn: (id) => toQueryResult(apiClient.v1CustomersDelete(id)),
      invalidatesTags: ['Customer'],
    }),

    getDestinations: builder.query<PagedResult<DtoDestinationResponse>, (PagedListArgs & { includeInactive?: boolean }) | void>({
      queryFn: (args: PagedListArgs & { includeInactive?: boolean } = {}) => toPagedQueryResult(apiClient.v1DestinationsList({ search: args.search || undefined, include_inactive: args.includeInactive || undefined, page: args.page, limit: args.limit })),
      providesTags: ['Destination'],
    }),
    createDestination: builder.mutation<DtoDestinationResponse, DtoDestinationRequest>({
      queryFn: (body) => toQueryResult(apiClient.v1DestinationsCreate(body)),
      invalidatesTags: ['Destination'],
    }),
    updateDestination: builder.mutation<DtoDestinationResponse, { id: string; body: DtoDestinationRequest }>({
      queryFn: ({ id, body }) => toQueryResult(apiClient.v1DestinationsUpdate(id, body)),
      invalidatesTags: ['Destination'],
    }),
    deleteDestination: builder.mutation<void, string>({
      queryFn: (id) => toQueryResult(apiClient.v1DestinationsDelete(id)),
      invalidatesTags: ['Destination'],
    }),

    getMotrici: builder.query<PagedResult<DtoMotriceResponse>, PagedListArgs | void>({
      queryFn: (args: PagedListArgs = {}) => toPagedQueryResult(apiClient.v1MotriciList({ search: args.search || undefined, page: args.page, limit: args.limit })),
      providesTags: ['Motrice'],
    }),
    createMotrice: builder.mutation<DtoMotriceResponse, DtoMotriceRequest>({
      queryFn: (body) => toQueryResult(apiClient.v1MotriciCreate(body)),
      invalidatesTags: ['Motrice'],
    }),
    updateMotrice: builder.mutation<DtoMotriceResponse, { id: string; body: DtoMotriceRequest }>({
      queryFn: ({ id, body }) => toQueryResult(apiClient.v1MotriciUpdate(id, body)),
      invalidatesTags: ['Motrice'],
    }),
    deleteMotrice: builder.mutation<void, string>({
      queryFn: (id) => toQueryResult(apiClient.v1MotriciDelete(id)),
      invalidatesTags: ['Motrice'],
    }),

    getSemirimorchi: builder.query<PagedResult<DtoSemirimorchioResponse>, PagedListArgs | void>({
      queryFn: (args: PagedListArgs = {}) => toPagedQueryResult(apiClient.v1SemirimorchiList({ search: args.search || undefined, page: args.page, limit: args.limit })),
      providesTags: ['Semirimorchio'],
    }),
    createSemirimorchio: builder.mutation<DtoSemirimorchioResponse, DtoSemirimorchioRequest>({
      queryFn: (body) => toQueryResult(apiClient.v1SemirimorchiCreate(body)),
      invalidatesTags: ['Semirimorchio'],
    }),
    updateSemirimorchio: builder.mutation<DtoSemirimorchioResponse, { id: string; body: DtoSemirimorchioRequest }>({
      queryFn: ({ id, body }) => toQueryResult(apiClient.v1SemirimorchiUpdate(id, body)),
      invalidatesTags: ['Semirimorchio'],
    }),
    deleteSemirimorchio: builder.mutation<void, string>({
      queryFn: (id) => toQueryResult(apiClient.v1SemirimorchiDelete(id)),
      invalidatesTags: ['Semirimorchio'],
    }),

    getDrivers: builder.query<PagedResult<DtoDriverResponse>, PagedListArgs | void>({
      queryFn: (args: PagedListArgs = {}) => toPagedQueryResult(apiClient.v1DriversList({ search: args.search || undefined, page: args.page, limit: args.limit })),
      providesTags: ['Driver'],
    }),
    createDriver: builder.mutation<DtoDriverResponse, DtoDriverRequest>({
      queryFn: (body) => toQueryResult(apiClient.v1DriversCreate(body)),
      invalidatesTags: ['Driver'],
    }),
    getDriverTrips: builder.query<DtoTripResponse[], string>({
      queryFn: (autistaId) => toQueryResult(apiClient.v1TripsList({ autista_id: autistaId })),
    }),
    getDriverUnavailability: builder.query<DtoDriverUnavailabilityResponse[], string>({
      queryFn: (autistaId) => toQueryResult(apiClient.v1DriverUnavailabilityList({ autista_id: autistaId })),
      providesTags: ['DriverUnavailability'],
    }),
    createDriverUnavailability: builder.mutation<DtoDriverUnavailabilityResponse, DtoDriverUnavailabilityRequest>({
      queryFn: (body) => toQueryResult(apiClient.v1DriverUnavailabilityCreate(body)),
      invalidatesTags: ['DriverUnavailability', 'Driver'],
    }),
    deleteDriverUnavailability: builder.mutation<Record<string, boolean>, string>({
      queryFn: (id) => toQueryResult(apiClient.v1DriverUnavailabilityDelete(id)),
      invalidatesTags: ['DriverUnavailability', 'Driver'],
    }),
    updateDriver: builder.mutation<DtoDriverResponse, { id: string; body: DtoDriverRequest }>({
      queryFn: ({ id, body }) => toQueryResult(apiClient.v1DriversUpdate(id, body)),
      invalidatesTags: ['Driver'],
    }),
    deleteDriver: builder.mutation<void, string>({
      queryFn: (id) => toQueryResult(apiClient.v1DriversDelete(id)),
      invalidatesTags: ['Driver'],
    }),

    getCarriers: builder.query<PagedResult<DtoCarrierResponse>, PagedListArgs | void>({
      queryFn: (args: PagedListArgs = {}) => toPagedQueryResult(apiClient.v1CarriersList({ search: args.search || undefined, page: args.page, limit: args.limit })),
      providesTags: ['Carrier'],
    }),
    createCarrier: builder.mutation<DtoCarrierResponse, DtoCarrierRequest>({
      queryFn: (body) => toQueryResult(apiClient.v1CarriersCreate(body)),
      invalidatesTags: ['Carrier'],
    }),
    updateCarrier: builder.mutation<DtoCarrierResponse, { id: string; body: DtoCarrierRequest }>({
      queryFn: ({ id, body }) => toQueryResult(apiClient.v1CarriersUpdate(id, body)),
      invalidatesTags: ['Carrier'],
    }),
    deleteCarrier: builder.mutation<void, string>({
      queryFn: (id) => toQueryResult(apiClient.v1CarriersDelete(id)),
      invalidatesTags: ['Carrier'],
    }),

    getProducts: builder.query<PagedResult<DtoProductResponse>, PagedListArgs | void>({
      queryFn: (args: PagedListArgs = {}) => toPagedQueryResult(apiClient.v1ProductsList({ search: args.search || undefined, page: args.page, limit: args.limit })),
      providesTags: ['Product'],
    }),
    createProduct: builder.mutation<DtoProductResponse, DtoProductRequest>({
      queryFn: (body) => toQueryResult(apiClient.v1ProductsCreate(body)),
      invalidatesTags: ['Product'],
    }),
    updateProduct: builder.mutation<DtoProductResponse, { id: string; body: DtoProductRequest }>({
      queryFn: ({ id, body }) => toQueryResult(apiClient.v1ProductsUpdate(id, body)),
      invalidatesTags: ['Product'],
    }),
    deleteProduct: builder.mutation<void, string>({
      queryFn: (id) => toQueryResult(apiClient.v1ProductsDelete(id)),
      invalidatesTags: ['Product'],
    }),

    // Garages: nessun filtro `search` lato backend (a differenza degli altri).
    getGarages: builder.query<PagedResult<DtoGarageResponse>, ({ includeInactive?: boolean } & PagedListArgs) | void>({
      queryFn: (args: { includeInactive?: boolean } & PagedListArgs = {}) => toPagedQueryResult(apiClient.v1GaragesList({ include_inactive: args.includeInactive || undefined, page: args.page, limit: args.limit })),
      providesTags: ['Garage'],
    }),
    createGarage: builder.mutation<DtoGarageResponse, DtoGarageRequest>({
      queryFn: (body) => toQueryResult(apiClient.v1GaragesCreate(body)),
      invalidatesTags: ['Garage'],
    }),
    updateGarage: builder.mutation<DtoGarageResponse, { id: string; body: DtoGarageRequest }>({
      queryFn: ({ id, body }) => toQueryResult(apiClient.v1GaragesUpdate(id, body)),
      invalidatesTags: ['Garage'],
    }),
    deleteGarage: builder.mutation<void, string>({
      queryFn: (id) => toQueryResult(apiClient.v1GaragesDelete(id)),
      invalidatesTags: ['Garage'],
    }),

    getWashStations: builder.query<PagedResult<DtoWashStationResponse>, ({ includeInactive?: boolean } & PagedListArgs) | void>({
      queryFn: (args: { includeInactive?: boolean } & PagedListArgs = {}) => toPagedQueryResult(apiClient.v1WashStationsList({ include_inactive: args.includeInactive || undefined, page: args.page, limit: args.limit })),
      providesTags: ['WashStation'],
    }),
    // Elenco completo senza paginazione — per i picker (select "punto di
    // lavaggio") che devono ordinare/filtrare sull'intero set, non su una
    // pagina troncata a un limite arbitrario.
    getAllWashStations: builder.query<DtoWashStationResponse[], void>({
      queryFn: () => toQueryResult(apiClient.v1WashStationsAllList()),
      providesTags: ['WashStation'],
    }),
    createWashStation: builder.mutation<DtoWashStationResponse, DtoWashStationRequest>({
      queryFn: (body) => toQueryResult(apiClient.v1WashStationsCreate(body)),
      invalidatesTags: ['WashStation'],
    }),
    updateWashStation: builder.mutation<DtoWashStationResponse, { id: string; body: DtoWashStationRequest }>({
      queryFn: ({ id, body }) => toQueryResult(apiClient.v1WashStationsUpdate(id, body)),
      invalidatesTags: ['WashStation'],
    }),
    deleteWashStation: builder.mutation<void, string>({
      queryFn: (id) => toQueryResult(apiClient.v1WashStationsDelete(id)),
      invalidatesTags: ['WashStation'],
    }),

    getCountries: builder.query<PagedResult<DtoCountryResponse>, PagedListArgs | void>({
      queryFn: (args: PagedListArgs = {}) => toPagedQueryResult(apiClient.v1CountriesList({ search: args.search || undefined, page: args.page, limit: args.limit })),
      providesTags: ['Country'],
    }),
    createCountry: builder.mutation<DtoCountryResponse, DtoCountryRequest>({
      queryFn: (body) => toQueryResult(apiClient.v1CountriesCreate(body)),
      invalidatesTags: ['Country'],
    }),
    updateCountry: builder.mutation<DtoCountryResponse, { id: string; body: DtoCountryRequest }>({
      queryFn: ({ id, body }) => toQueryResult(apiClient.v1CountriesUpdate(id, body)),
      invalidatesTags: ['Country'],
    }),
    deleteCountry: builder.mutation<void, string>({
      queryFn: (id) => toQueryResult(apiClient.v1CountriesDelete(id)),
      invalidatesTags: ['Country'],
    }),

    getBanks: builder.query<PagedResult<DtoBankResponse>, PagedListArgs | void>({
      queryFn: (args: PagedListArgs = {}) => toPagedQueryResult(apiClient.v1BanksList({ search: args.search || undefined, page: args.page, limit: args.limit })),
      providesTags: ['Bank'],
    }),
    createBank: builder.mutation<DtoBankResponse, DtoBankRequest>({
      queryFn: (body) => toQueryResult(apiClient.v1BanksCreate(body)),
      invalidatesTags: ['Bank'],
    }),
    updateBank: builder.mutation<DtoBankResponse, { id: string; body: DtoBankRequest }>({
      queryFn: ({ id, body }) => toQueryResult(apiClient.v1BanksUpdate(id, body)),
      invalidatesTags: ['Bank'],
    }),
    deleteBank: builder.mutation<void, string>({
      queryFn: (id) => toQueryResult(apiClient.v1BanksDelete(id)),
      invalidatesTags: ['Bank'],
    }),

    getAccountingEntries: builder.query<PagedResult<DtoAccountingEntryResponse>, PagedListArgs | void>({
      queryFn: (args: PagedListArgs = {}) => toPagedQueryResult(apiClient.v1AccountingEntriesList({ search: args.search || undefined, page: args.page, limit: args.limit })),
      providesTags: ['AccountingEntry'],
    }),
    createAccountingEntry: builder.mutation<DtoAccountingEntryResponse, DtoAccountingEntryRequest>({
      queryFn: (body) => toQueryResult(apiClient.v1AccountingEntriesCreate(body)),
      invalidatesTags: ['AccountingEntry'],
    }),
    updateAccountingEntry: builder.mutation<DtoAccountingEntryResponse, { id: string; body: DtoAccountingEntryRequest }>({
      queryFn: ({ id, body }) => toQueryResult(apiClient.v1AccountingEntriesUpdate(id, body)),
      invalidatesTags: ['AccountingEntry'],
    }),
    deleteAccountingEntry: builder.mutation<void, string>({
      queryFn: (id) => toQueryResult(apiClient.v1AccountingEntriesDelete(id)),
      invalidatesTags: ['AccountingEntry'],
    }),

    getAdminUsers: builder.query<DtoAuthUserResponse[], void>({
      queryFn: () => toQueryResult(apiClient.v1AdminUsersList()),
      providesTags: ['AdminUser'],
    }),
    createAdminUser: builder.mutation<DtoAuthUserResponse, DtoRegisterRequest>({
      // Creazione utente admin-facing: passa da /auth/register (non da
      // POST /admin/users, che opera su uno shape Login/Role diverso e non
      // usato dal frontend — vedi backend/internal/services/admin_panel).
      queryFn: (body) => toQueryResult(apiClient.v1AuthRegisterCreate(body)),
      invalidatesTags: ['AdminUser'],
    }),
    // Nome/ruolo e stato attivo vivono su due endpoint distinti sul backend
    // (PUT /admin/users/{id} per login/name/role, PATCH per active — vedi
    // backend/internal/services/admin_panel/admin_user.go): un solo
    // mutation qui nasconde la doppia chiamata, così i componenti non
    // devono conoscere questo dettaglio implementativo.
    updateAdminUser: builder.mutation<
      DtoAuthUserResponse,
      { id: number; login: string; name: string; role: DtoUpdateUserRequest['role']; active: boolean }
    >({
      queryFn: async ({ id, login, name, role, active }) => {
        try {
          await apiClient.v1AdminUsersUpdate(id, { login, name, role });
          const res = await apiClient.v1AdminUsersPartialUpdate(id, { active });
          return { data: res.data };
        } catch (err) {
          return { error: toQueryError(err) };
        }
      },
      invalidatesTags: ['AdminUser'],
    }),
    deleteAdminUser: builder.mutation<void, number>({
      queryFn: (id) => toQueryResult(apiClient.v1AdminUsersDelete(id)),
      invalidatesTags: ['AdminUser'],
    }),

    // Portale cliente (ruolo "cliente"): il backend forza sempre il proprio
    // customer_id preso dal JWT, ignorando qualunque id passato qui — questi
    // endpoint non accettano/servono mai un id esplicito.
    getMyAnagrafica: builder.query<DtoCustomerResponse, void>({
      queryFn: () => toQueryResult(apiClient.v1MeAnagraficaList()),
      providesTags: ['MyAnagrafica'],
    }),
    updateMyAnagrafica: builder.mutation<DtoCustomerResponse, DtoCustomerRequest>({
      queryFn: (body) => toQueryResult(apiClient.v1MeAnagraficaUpdate(body)),
      invalidatesTags: ['MyAnagrafica'],
    }),
    getMyOrders: builder.query<DtoOrderResponse[], { stato?: string } | void>({
      queryFn: (args?: { stato?: string }) => toQueryResult(apiClient.v1MeOrdersList({ stato: args?.stato || undefined })),
      providesTags: ['MyOrder'],
    }),
    getMyOrder: builder.query<DtoOrderResponse, string>({
      queryFn: (id) => toQueryResult(apiClient.v1MeOrdersDetail(id)),
      providesTags: ['MyOrder'],
    }),
    // Una richiesta del cliente non crea più direttamente un Order: entra
    // come InboundOrder "da confermare" (stessa coda di revisione di
    // mail/PDF) — GET per farla vedere al cliente stesso finché è in attesa,
    // POST per crearla. Una volta accettata dall'operatore compare tra i
    // 'MyOrder' normali, non più qui.
    getMyInboundOrders: builder.query<DtoInboundOrderResponse[], void>({
      queryFn: () => toQueryResult(apiClient.v1MeInboundOrdersList()),
      providesTags: ['MyInboundOrder'],
    }),
    createMyInboundOrder: builder.mutation<DtoInboundOrderResponse, DtoOrderRequest>({
      queryFn: (body) => toQueryResult(apiClient.v1MeInboundOrdersCreate(body)),
      invalidatesTags: ['MyInboundOrder'],
    }),
    deleteMyOrder: builder.mutation<void, string>({
      queryFn: (id) => toQueryResult(apiClient.v1MeOrdersDelete(id)),
      invalidatesTags: ['MyOrder'],
    }),
    // Pool condiviso con lo staff (stesso tag 'Destination' di getDestinations)
    // — un cliente può solo aggiungere, non modificare/eliminare quelle esistenti.
    createMyDestination: builder.mutation<DtoDestinationResponse, DtoDestinationRequest>({
      queryFn: (body) => toQueryResult(apiClient.v1MeDestinationsCreate(body)),
      invalidatesTags: ['Destination'],
    }),
  }),
});

export const {
  useGetDashboardStatsQuery,
  useGetRecentOrdersQuery,
  useGetCustomerDashboardQuery,
  useGetCustomersQuery,
  useCreateCustomerMutation,
  useUpdateCustomerMutation,
  useDeleteCustomerMutation,
  useGetDestinationsQuery,
  useCreateDestinationMutation,
  useUpdateDestinationMutation,
  useDeleteDestinationMutation,
  useGetMotriciQuery,
  useCreateMotriceMutation,
  useUpdateMotriceMutation,
  useDeleteMotriceMutation,
  useGetSemirimorchiQuery,
  useCreateSemirimorchioMutation,
  useUpdateSemirimorchioMutation,
  useDeleteSemirimorchioMutation,
  useGetDriversQuery,
  useCreateDriverMutation,
  useUpdateDriverMutation,
  useDeleteDriverMutation,
  useGetDriverTripsQuery,
  useGetDriverUnavailabilityQuery,
  useCreateDriverUnavailabilityMutation,
  useDeleteDriverUnavailabilityMutation,
  useGetCarriersQuery,
  useCreateCarrierMutation,
  useUpdateCarrierMutation,
  useDeleteCarrierMutation,
  useGetProductsQuery,
  useCreateProductMutation,
  useUpdateProductMutation,
  useDeleteProductMutation,
  useGetGaragesQuery,
  useCreateGarageMutation,
  useUpdateGarageMutation,
  useDeleteGarageMutation,
  useGetWashStationsQuery,
  useGetAllWashStationsQuery,
  useCreateWashStationMutation,
  useUpdateWashStationMutation,
  useDeleteWashStationMutation,
  useGetCountriesQuery,
  useCreateCountryMutation,
  useUpdateCountryMutation,
  useDeleteCountryMutation,
  useGetBanksQuery,
  useCreateBankMutation,
  useUpdateBankMutation,
  useDeleteBankMutation,
  useGetAccountingEntriesQuery,
  useCreateAccountingEntryMutation,
  useUpdateAccountingEntryMutation,
  useDeleteAccountingEntryMutation,
  useGetAdminUsersQuery,
  useCreateAdminUserMutation,
  useUpdateAdminUserMutation,
  useDeleteAdminUserMutation,
  useGetMyAnagraficaQuery,
  useUpdateMyAnagraficaMutation,
  useGetMyOrdersQuery,
  useGetMyOrderQuery,
  useDeleteMyOrderMutation,
  useGetMyInboundOrdersQuery,
  useCreateMyInboundOrderMutation,
  useCreateMyDestinationMutation,
} = appApi;
