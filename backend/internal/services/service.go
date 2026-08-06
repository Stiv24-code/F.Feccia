//go:generate mockgen -source=service.go -destination=mocks/services_mock.go -package=mocks

package services

import (
	"context"
	"fratelli-feccia/internal/dto"
	admin "fratelli-feccia/internal/services/admin_panel"
	"fratelli-feccia/internal/services/anagrafiche"
	"fratelli-feccia/internal/services/auth"
	"fratelli-feccia/internal/services/availability"
	"fratelli-feccia/internal/services/carriers"
	"fratelli-feccia/internal/services/customers"
	"fratelli-feccia/internal/services/dashboard"
	"fratelli-feccia/internal/services/destinations"
	"fratelli-feccia/internal/services/drivers"
	"fratelli-feccia/internal/services/driverunavailability"
	"fratelli-feccia/internal/services/export"
	"fratelli-feccia/internal/services/garages"
	"fratelli-feccia/internal/services/geocode"
	"fratelli-feccia/internal/services/invoices"
	"fratelli-feccia/internal/services/mapview"
	"fratelli-feccia/internal/services/masterdata"
	"fratelli-feccia/internal/services/orders"
	"fratelli-feccia/internal/services/pricelists"
	"fratelli-feccia/internal/services/products"
	"fratelli-feccia/internal/services/trips"
	"fratelli-feccia/internal/services/vehicles"
	"fratelli-feccia/internal/services/washstations"
	"fratelli-feccia/pkg/s3invoices"
	"fratelli-feccia/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminService interface {
	ListUsers(ctx context.Context, page, limit int) ([]dto.UserResponse, int64, error)
	GetUserByID(ctx context.Context, id int64) (*dto.UserResponse, error)
	CreateUser(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error)
	UpdateUser(ctx context.Context, id int64, item dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(ctx context.Context, id int64) error
	CountAdmins(ctx context.Context) (int64, error)
	ListAllUsers(ctx context.Context) ([]dto.AuthUserResponse, error)
	PatchUser(ctx context.Context, id int64, req dto.PatchUserRequest) (*dto.AuthUserResponse, error)
}

type Auth interface {
	Login(login, password string) (*dto.LoginResult, error)
	Refresh(refreshToken string) (*dto.LoginResult, error)
	Me(userID int64) (*dto.AuthUserResponse, error)
	Register(req dto.RegisterRequest) (*dto.AuthUserResponse, error)
	RegisterClient(req dto.ClientRegisterRequest) (*dto.LoginResult, error)
}

type Customer interface {
	List(ctx context.Context, search string) ([]dto.CustomerResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.CustomerResponse, error)
	Create(ctx context.Context, req dto.CustomerRequest) (*dto.CustomerResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.CustomerRequest) (*dto.CustomerResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Destination interface {
	List(ctx context.Context, search string, includeInactive bool) ([]dto.DestinationResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.DestinationResponse, error)
	Create(ctx context.Context, req dto.DestinationRequest) (*dto.DestinationResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.DestinationRequest) (*dto.DestinationResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Carrier interface {
	List(ctx context.Context, search string) ([]dto.CarrierResponse, error)
	Create(ctx context.Context, req dto.CarrierRequest) (*dto.CarrierResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.CarrierRequest) (*dto.CarrierResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Garage interface {
	List(ctx context.Context, includeInactive bool) ([]dto.GarageResponse, error)
	Create(ctx context.Context, req dto.GarageRequest) (*dto.GarageResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.GarageRequest) (*dto.GarageResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type WashStation interface {
	List(ctx context.Context, includeInactive bool) ([]dto.WashStationResponse, error)
	Create(ctx context.Context, req dto.WashStationRequest) (*dto.WashStationResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.WashStationRequest) (*dto.WashStationResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Driver interface {
	List(ctx context.Context, search string) ([]dto.DriverResponse, error)
	Create(ctx context.Context, req dto.DriverRequest) (*dto.DriverResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.DriverRequest) (*dto.DriverResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Product interface {
	List(ctx context.Context, search string) ([]dto.ProductResponse, error)
	Create(ctx context.Context, req dto.ProductRequest) (*dto.ProductResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.ProductRequest) (*dto.ProductResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Masterdata interface {
	ListVehicleTypes(ctx context.Context) ([]dto.VehicleTypeResponse, error)
	CreateVehicleType(ctx context.Context, req dto.VehicleTypeRequest) (*dto.VehicleTypeResponse, error)
	ListAccessoryCosts(ctx context.Context) ([]dto.AccessoryCostResponse, error)
	CreateAccessoryCost(ctx context.Context, req dto.AccessoryCostRequest) (*dto.AccessoryCostResponse, error)
	ListTransportCategories(ctx context.Context) ([]dto.TransportCategoryResponse, error)
	CreateTransportCategory(ctx context.Context, req dto.TransportCategoryRequest) (*dto.TransportCategoryResponse, error)
}

type Anagrafiche interface {
	ListCountries(ctx context.Context, search string) ([]dto.CountryResponse, error)
	CreateCountry(ctx context.Context, req dto.CountryRequest) (*dto.CountryResponse, error)
	UpdateCountry(ctx context.Context, id uuid.UUID, req dto.CountryRequest) (*dto.CountryResponse, error)
	DeleteCountry(ctx context.Context, id uuid.UUID) error
	ListBanks(ctx context.Context, search string) ([]dto.BankResponse, error)
	CreateBank(ctx context.Context, req dto.BankRequest) (*dto.BankResponse, error)
	UpdateBank(ctx context.Context, id uuid.UUID, req dto.BankRequest) (*dto.BankResponse, error)
	DeleteBank(ctx context.Context, id uuid.UUID) error
	ListAccountingEntries(ctx context.Context, search, tipo string) ([]dto.AccountingEntryResponse, error)
	CreateAccountingEntry(ctx context.Context, req dto.AccountingEntryRequest) (*dto.AccountingEntryResponse, error)
	UpdateAccountingEntry(ctx context.Context, id uuid.UUID, req dto.AccountingEntryRequest) (*dto.AccountingEntryResponse, error)
	DeleteAccountingEntry(ctx context.Context, id uuid.UUID) error
}

type DriverUnavailability interface {
	List(ctx context.Context, autistaID uuid.UUID, dataDa, dataA string) ([]dto.DriverUnavailabilityResponse, error)
	Create(ctx context.Context, req dto.DriverUnavailabilityRequest) (*dto.DriverUnavailabilityResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Order interface {
	List(ctx context.Context, f orders.ListFilters) ([]dto.OrderResponse, error)
	Create(ctx context.Context, req dto.OrderRequest) (*dto.OrderResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.OrderRequest) (*dto.OrderResponse, error)
	Assign(ctx context.Context, id uuid.UUID, req dto.OrderAssignRequest) (*dto.OrderResponse, error)
	Unassign(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error)
	Start(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error)
	Close(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error)
	Discard(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ReturnSuggestions(ctx context.Context, id uuid.UUID, maxDaysGap, limit int) (*dto.OrderReturnSuggestionsResponse, error)
	GetCMRPDF(ctx context.Context, id uuid.UUID) ([]byte, string, error)
	RouteAlternatives(ctx context.Context, id uuid.UUID, garageID, washStationID string) ([]dto.RouteAlternativeDTO, error)
	UpdateRoute(ctx context.Context, id uuid.UUID, waypoints []dto.RouteWaypointDTO) (*dto.OrderResponse, error)
}

type Vehicle interface {
	List(ctx context.Context, search string) ([]dto.VehicleResponse, error)
	Create(ctx context.Context, req dto.VehicleRequest) (*dto.VehicleResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.VehicleRequest) (*dto.VehicleResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateGPSByID(ctx context.Context, idOrTarga string, req dto.VehicleGPSUpdateRequest) (*dto.GPSUpdateResult, error)
	UpdateGPSByPlate(ctx context.Context, targa string, req dto.VehicleGPSUpdateRequest) (*dto.GPSUpdateResult, error)
	GetGPSHistory(ctx context.Context, vehicleIDOrTarga string, limit int) ([]dto.GPSHistoryResponse, error)
	GetAllGPSLive(ctx context.Context) ([]dto.GPSLiveVehicle, error)
	IngestGPSWebhook(ctx context.Context, vendor string, payload dto.GPSWebhookPayload) (*dto.GPSUpdateResult, error)
	IngestTemperatureWebhook(ctx context.Context, vendor string, payload dto.TemperatureWebhookRequest) (*dto.TemperatureWebhookResult, error)
	GetTemperatureHistory(ctx context.Context, vehicleID uuid.UUID, limit int, onlyAlerts bool) ([]dto.TemperatureReadingResponse, error)
	SetTemperatureThresholds(ctx context.Context, vehicleID uuid.UUID, req dto.TemperatureThresholdsRequest) (*dto.TemperatureThresholdsResult, error)
}

type Trip interface {
	List(ctx context.Context, stato string, limit int) ([]dto.TripResponse, error)
	Create(ctx context.Context, req dto.TripRequest) (*dto.TripResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.TripDetailResponse, error)
	RecomputeSegments(ctx context.Context, id uuid.UUID) (*dto.RecomputeSegmentsResult, error)
	Start(ctx context.Context, id uuid.UUID) (*dto.OKResult, error)
	Complete(ctx context.Context, id uuid.UUID) (*dto.OKResult, error)
	AddOrder(ctx context.Context, tripID, orderID uuid.UUID) (*dto.OKResult, error)
	GetInstructionsPDF(ctx context.Context, id uuid.UUID) ([]byte, string, error)
}

type PriceList interface {
	List(ctx context.Context, clienteID string) ([]dto.PriceListResponse, error)
	Create(ctx context.Context, req dto.PriceListRequest) (*dto.PriceListResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.PriceListResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.PriceListRequest) (*dto.PriceListUpdateResult, error)
	Delete(ctx context.Context, id uuid.UUID) error
	AddItem(ctx context.Context, plID uuid.UUID, item dto.PriceListItemRequestDTO) (*dto.PriceListItemAddResult, error)
	UpdateItem(ctx context.Context, plID, itemID uuid.UUID, item dto.PriceListItemRequestDTO) (*dto.PriceListItemUpdateResult, error)
	DeleteItem(ctx context.Context, plID, itemID uuid.UUID) (*dto.PriceListItemDeleteResult, error)
	LookupTariff(ctx context.Context, clienteID, caricoID, scaricoID, prodottoID string, peso float64) (*dto.TariffLookupResult, error)
}

type Invoice interface {
	List(ctx context.Context, stato, clienteID string) ([]dto.InvoiceResponse, error)
	Create(ctx context.Context, req dto.InvoiceRequest) (*dto.InvoiceResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.InvoiceResponse, error)
	Finalize(ctx context.Context, id uuid.UUID) (*dto.InvoiceFinalizeResult, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetPDF(ctx context.Context, id uuid.UUID) ([]byte, string, error)
	GetPDFPresignedURL(ctx context.Context, id uuid.UUID) (*dto.InvoicePDFURLResult, error)
}

type Dashboard interface {
	Stats(ctx context.Context) (*dto.DashboardStatsResponse, error)
	CustomerDashboard(ctx context.Context, customerID uuid.UUID) (*dto.CustomerDashboardResponse, error)
	RecentOrders(ctx context.Context) ([]dto.OrderResponse, error)
}

type Map interface {
	Trips(ctx context.Context) (*dto.MapTripsResponse, error)
}

type Geocode interface {
	Search(ctx context.Context, query string) ([]dto.GeocodeResultDTO, error)
}

type Availability interface {
	VehicleAvailability(ctx context.Context, dataDa, dataA string) ([]dto.VehicleAvailabilityResponse, error)
	DriverAvailability(ctx context.Context, dataDa, dataA string) ([]dto.DriverAvailabilityResponse, error)
}

type Export interface {
	OrdersExcel(ctx context.Context, filter export.OrdersFilter) ([]byte, error)
}

type Admin struct {
	Admin AdminService
}

type Authentication struct {
	Auth Auth
}

type Customers struct {
	Customer Customer
}

type Destinations struct {
	Destination Destination
}

type Carriers struct {
	Carrier Carrier
}

type Garages struct {
	Garage Garage
}

type WashStations struct {
	WashStation WashStation
}

type Drivers struct {
	Driver Driver
}

type Products struct {
	Product Product
}

type MasterdataGroup struct {
	Masterdata Masterdata
}

type AnagraficheGroup struct {
	Anagrafiche Anagrafiche
}

type DriverUnavailabilityGroup struct {
	DriverUnavailability DriverUnavailability
}

type Orders struct {
	Order Order
}

type Vehicles struct {
	Vehicle Vehicle
}

type Trips struct {
	Trip Trip
}

type PriceLists struct {
	PriceList PriceList
}

type Invoices struct {
	Invoice Invoice
}

type DashboardGroup struct {
	Dashboard Dashboard
}

type MapGroup struct {
	Map Map
}

type AvailabilityGroup struct {
	Availability Availability
}

type ExportGroup struct {
	Export Export
}

type GeocodeGroup struct {
	Geocode Geocode
}

type Service struct {
	Admin
	Authentication
	Customers
	Destinations
	Carriers
	Garages
	WashStations
	Drivers
	Products
	MasterdataGroup
	AnagraficheGroup
	DriverUnavailabilityGroup
	Orders
	Vehicles
	Trips
	PriceLists
	Invoices
	DashboardGroup
	MapGroup
	AvailabilityGroup
	ExportGroup
	GeocodeGroup
}

func NewService(db *gorm.DB, jwtConf utils.JWTConfig, s3Client *s3invoices.Client, orsApiKey, orsBaseURL string) *Service {
	return &Service{
		Admin: Admin{
			Admin: admin.NewAdminService(db, jwtConf),
		},
		Authentication: Authentication{
			Auth: auth.NewAuthService(db, jwtConf),
		},
		Customers: Customers{
			Customer: customers.NewCustomerService(db),
		},
		Destinations: Destinations{
			Destination: destinations.NewDestinationService(db),
		},
		Carriers: Carriers{
			Carrier: carriers.NewCarrierService(db),
		},
		Garages: Garages{
			Garage: garages.NewGarageService(db),
		},
		WashStations: WashStations{
			WashStation: washstations.NewWashStationService(db),
		},
		Drivers: Drivers{
			Driver: drivers.NewDriverService(db),
		},
		Products: Products{
			Product: products.NewProductService(db),
		},
		MasterdataGroup: MasterdataGroup{
			Masterdata: masterdata.NewMasterdataService(db),
		},
		AnagraficheGroup: AnagraficheGroup{
			Anagrafiche: anagrafiche.NewAnagraficheService(db),
		},
		DriverUnavailabilityGroup: DriverUnavailabilityGroup{
			DriverUnavailability: driverunavailability.NewDriverUnavailabilityService(db),
		},
		Orders: Orders{
			Order: orders.NewOrderService(db, orsApiKey, orsBaseURL),
		},
		Vehicles: Vehicles{
			Vehicle: vehicles.NewVehicleService(db),
		},
		Trips: Trips{
			Trip: trips.NewTripService(db, orsApiKey, orsBaseURL),
		},
		PriceLists: PriceLists{
			PriceList: pricelists.NewPriceListService(db),
		},
		Invoices: Invoices{
			Invoice: invoices.NewInvoiceService(db, s3Client),
		},
		DashboardGroup: DashboardGroup{
			Dashboard: dashboard.NewDashboardService(db),
		},
		MapGroup: MapGroup{
			Map: mapview.NewMapService(db, orsApiKey, orsBaseURL),
		},
		AvailabilityGroup: AvailabilityGroup{
			Availability: availability.NewAvailabilityService(db),
		},
		ExportGroup: ExportGroup{
			Export: export.NewExportService(db),
		},
		GeocodeGroup: GeocodeGroup{
			Geocode: geocode.NewGeocodeService(orsApiKey, orsBaseURL),
		},
	}
}
