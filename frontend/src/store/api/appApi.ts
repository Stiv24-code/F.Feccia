// Layer RTK Query: caching/invalidation sopra il client tipizzato generato
// da swagger (src/lib/apiClient.ts). Ogni endpoint usa `queryFn` invece di
// `query` per riusare i metodi già tipizzati di `apiClient` (path, body e
// response provengono da src/api/data-contracts.ts, generato dal backend) —
// niente duplicazione di path/tipi qui, RTK Query fa solo caching/tag.
import { createApi, fakeBaseQuery } from '@reduxjs/toolkit/query/react';
import { apiClient } from '@/lib/apiClient';
import { toQueryResult, toQueryError } from './rtkQueryHelpers';
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
    'MyAnagrafica', 'MyOrder',
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

    getCustomers: builder.query<DtoCustomerResponse[], string | void>({
      queryFn: (search) => toQueryResult(apiClient.v1CustomersList({ search: search || undefined })),
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

    getDestinations: builder.query<DtoDestinationResponse[], { search?: string; includeInactive?: boolean } | void>({
      queryFn: (args: { search?: string; includeInactive?: boolean } = {}) => toQueryResult(apiClient.v1DestinationsList({ search: args.search || undefined, include_inactive: args.includeInactive || undefined })),
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

    getMotrici: builder.query<DtoMotriceResponse[], string | void>({
      queryFn: (search) => toQueryResult(apiClient.v1MotriciList({ search: search || undefined })),
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

    getSemirimorchi: builder.query<DtoSemirimorchioResponse[], string | void>({
      queryFn: (search) => toQueryResult(apiClient.v1SemirimorchiList({ search: search || undefined })),
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

    getDrivers: builder.query<DtoDriverResponse[], string | void>({
      queryFn: (search) => toQueryResult(apiClient.v1DriversList({ search: search || undefined })),
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

    getCarriers: builder.query<DtoCarrierResponse[], string | void>({
      queryFn: (search) => toQueryResult(apiClient.v1CarriersList({ search: search || undefined })),
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

    getProducts: builder.query<DtoProductResponse[], string | void>({
      queryFn: (search) => toQueryResult(apiClient.v1ProductsList({ search: search || undefined })),
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
    getGarages: builder.query<DtoGarageResponse[], boolean | void>({
      queryFn: (includeInactive) => toQueryResult(apiClient.v1GaragesList({ include_inactive: includeInactive || undefined })),
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

    getWashStations: builder.query<DtoWashStationResponse[], boolean | void>({
      queryFn: (includeInactive) => toQueryResult(apiClient.v1WashStationsList({ include_inactive: includeInactive || undefined })),
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

    getCountries: builder.query<DtoCountryResponse[], string | void>({
      queryFn: (search) => toQueryResult(apiClient.v1CountriesList({ search: search || undefined })),
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

    getBanks: builder.query<DtoBankResponse[], string | void>({
      queryFn: (search) => toQueryResult(apiClient.v1BanksList({ search: search || undefined })),
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

    getAccountingEntries: builder.query<DtoAccountingEntryResponse[], string | void>({
      queryFn: (search) => toQueryResult(apiClient.v1AccountingEntriesList({ search: search || undefined })),
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
    createMyOrder: builder.mutation<DtoOrderResponse, DtoOrderRequest>({
      queryFn: (body) => toQueryResult(apiClient.v1MeOrdersCreate(body)),
      invalidatesTags: ['MyOrder'],
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
  useCreateMyOrderMutation,
  useDeleteMyOrderMutation,
  useCreateMyDestinationMutation,
} = appApi;
