package handlers

import (
	anagrafiche_handlers "fratelli-feccia/internal/handlers/anagrafiche"
	auth_handlers "fratelli-feccia/internal/handlers/auth"
	availability_handlers "fratelli-feccia/internal/handlers/availability"
	carriers_handlers "fratelli-feccia/internal/handlers/carriers"
	customers_handlers "fratelli-feccia/internal/handlers/customers"
	dashboard_handlers "fratelli-feccia/internal/handlers/dashboard"
	destinations_handlers "fratelli-feccia/internal/handlers/destinations"
	drivers_handlers "fratelli-feccia/internal/handlers/drivers"
	driverunavailability_handlers "fratelli-feccia/internal/handlers/driverunavailability"
	export_handlers "fratelli-feccia/internal/handlers/export"
	garages_handlers "fratelli-feccia/internal/handlers/garages"
	invoices_handlers "fratelli-feccia/internal/handlers/invoices"
	mapview_handlers "fratelli-feccia/internal/handlers/mapview"
	masterdata_handlers "fratelli-feccia/internal/handlers/masterdata"
	orders_handlers "fratelli-feccia/internal/handlers/orders"
	pricelists_handlers "fratelli-feccia/internal/handlers/pricelists"
	products_handlers "fratelli-feccia/internal/handlers/products"
	trips_handlers "fratelli-feccia/internal/handlers/trips"
	users_handlers "fratelli-feccia/internal/handlers/users"
	vehicles_handlers "fratelli-feccia/internal/handlers/vehicles"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/audit"
	"fratelli-feccia/pkg/utils"
)

// Handler aggregates all HTTP handlers.
type Handler struct {
	Auth                 *auth_handlers.AuthHandler
	Admin                *users_handlers.UserHandler
	Customers            *customers_handlers.CustomerHandler
	Destinations         *destinations_handlers.DestinationHandler
	Carriers             *carriers_handlers.CarrierHandler
	Garages              *garages_handlers.GarageHandler
	Drivers              *drivers_handlers.DriverHandler
	Products             *products_handlers.ProductHandler
	Masterdata           *masterdata_handlers.MasterdataHandler
	Anagrafiche          *anagrafiche_handlers.AnagraficheHandler
	DriverUnavailability *driverunavailability_handlers.DriverUnavailabilityHandler
	Orders               *orders_handlers.OrderHandler
	Vehicles             *vehicles_handlers.VehicleHandler
	Trips                *trips_handlers.TripHandler
	PriceLists           *pricelists_handlers.PriceListHandler
	Invoices             *invoices_handlers.InvoiceHandler
	Dashboard            *dashboard_handlers.DashboardHandler
	Map                  *mapview_handlers.MapHandler
	Availability         *availability_handlers.AvailabilityHandler
	Export               *export_handlers.ExportHandler
}

func NewHandler(services *services.Service, auditLogger *audit.Logger, jwtCfg utils.JWTConfig) *Handler {
	return &Handler{
		Auth:                 auth_handlers.NewAuthHandler(services.Authentication.Auth, auditLogger, jwtCfg),
		Admin:                users_handlers.NewUserHandler(services.Admin.Admin),
		Customers:            customers_handlers.NewCustomerHandler(services.Customers.Customer),
		Destinations:         destinations_handlers.NewDestinationHandler(services.Destinations.Destination),
		Carriers:             carriers_handlers.NewCarrierHandler(services.Carriers.Carrier),
		Garages:              garages_handlers.NewGarageHandler(services.Garages.Garage),
		Drivers:              drivers_handlers.NewDriverHandler(services.Drivers.Driver),
		Products:             products_handlers.NewProductHandler(services.Products.Product),
		Masterdata:           masterdata_handlers.NewMasterdataHandler(services.MasterdataGroup.Masterdata),
		Anagrafiche:          anagrafiche_handlers.NewAnagraficheHandler(services.AnagraficheGroup.Anagrafiche),
		DriverUnavailability: driverunavailability_handlers.NewDriverUnavailabilityHandler(services.DriverUnavailabilityGroup.DriverUnavailability),
		Orders:               orders_handlers.NewOrderHandler(services.Orders.Order),
		Vehicles:             vehicles_handlers.NewVehicleHandler(services.Vehicles.Vehicle),
		Trips:                trips_handlers.NewTripHandler(services.Trips.Trip),
		PriceLists:           pricelists_handlers.NewPriceListHandler(services.PriceLists.PriceList),
		Invoices:             invoices_handlers.NewInvoiceHandler(services.Invoices.Invoice),
		Dashboard:            dashboard_handlers.NewDashboardHandler(services.DashboardGroup.Dashboard),
		Map:                  mapview_handlers.NewMapHandler(services.MapGroup.Map),
		Availability:         availability_handlers.NewAvailabilityHandler(services.AvailabilityGroup.Availability),
		Export:               export_handlers.NewExportHandler(services.ExportGroup.Export),
	}
}
