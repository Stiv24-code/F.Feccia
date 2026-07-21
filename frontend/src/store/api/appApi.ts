// Layer RTK Query: caching/invalidation sopra il client tipizzato generato
// da swagger (src/lib/apiClient.ts). Ogni endpoint usa `queryFn` invece di
// `query` per riusare i metodi già tipizzati di `apiClient` (path, body e
// response provengono da src/api/data-contracts.ts, generato dal backend) —
// niente duplicazione di path/tipi qui, RTK Query fa solo caching/tag.
import { createApi, fakeBaseQuery } from '@reduxjs/toolkit/query/react';
import { apiClient } from '@/lib/apiClient';
import { toQueryResult } from './rtkQueryHelpers';
import type {
  DtoAccountingEntryRequest,
  DtoAccountingEntryResponse,
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
  DtoGarageRequest,
  DtoGarageResponse,
  DtoOrderResponse,
  DtoProductRequest,
  DtoProductResponse,
  DtoVehicleRequest,
  DtoVehicleResponse,
  DtoWashStationRequest,
  DtoWashStationResponse,
} from '@/api/data-contracts';

export const appApi = createApi({
  reducerPath: 'appApi',
  baseQuery: fakeBaseQuery(),
  tagTypes: [
    'Customer', 'Dashboard', 'Destination', 'Vehicle', 'Driver',
    'Carrier', 'Product', 'Garage', 'WashStation', 'Country', 'Bank', 'AccountingEntry',
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
      queryFn: (args) => toQueryResult(apiClient.v1DestinationsList({ search: args?.search || undefined, include_inactive: args?.includeInactive || undefined })),
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

    getVehicles: builder.query<DtoVehicleResponse[], string | void>({
      queryFn: (search) => toQueryResult(apiClient.v1VehiclesList({ search: search || undefined })),
      providesTags: ['Vehicle'],
    }),
    createVehicle: builder.mutation<DtoVehicleResponse, DtoVehicleRequest>({
      queryFn: (body) => toQueryResult(apiClient.v1VehiclesCreate(body)),
      invalidatesTags: ['Vehicle'],
    }),
    updateVehicle: builder.mutation<DtoVehicleResponse, { id: string; body: DtoVehicleRequest }>({
      queryFn: ({ id, body }) => toQueryResult(apiClient.v1VehiclesUpdate(id, body)),
      invalidatesTags: ['Vehicle'],
    }),
    deleteVehicle: builder.mutation<void, string>({
      queryFn: (id) => toQueryResult(apiClient.v1VehiclesDelete(id)),
      invalidatesTags: ['Vehicle'],
    }),

    getDrivers: builder.query<DtoDriverResponse[], string | void>({
      queryFn: (search) => toQueryResult(apiClient.v1DriversList({ search: search || undefined })),
      providesTags: ['Driver'],
    }),
    createDriver: builder.mutation<DtoDriverResponse, DtoDriverRequest>({
      queryFn: (body) => toQueryResult(apiClient.v1DriversCreate(body)),
      invalidatesTags: ['Driver'],
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
  useGetVehiclesQuery,
  useCreateVehicleMutation,
  useUpdateVehicleMutation,
  useDeleteVehicleMutation,
  useGetDriversQuery,
  useCreateDriverMutation,
  useUpdateDriverMutation,
  useDeleteDriverMutation,
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
} = appApi;
