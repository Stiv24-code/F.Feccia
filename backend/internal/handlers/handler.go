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
	geocode_handlers "fratelli-feccia/internal/handlers/geocode"
	inboundorders_handlers "fratelli-feccia/internal/handlers/inboundorders"
	invoices_handlers "fratelli-feccia/internal/handlers/invoices"
	mapview_handlers "fratelli-feccia/internal/handlers/mapview"
	masterdata_handlers "fratelli-feccia/internal/handlers/masterdata"
	motrici_handlers "fratelli-feccia/internal/handlers/motrici"
	orders_handlers "fratelli-feccia/internal/handlers/orders"
	pdfimport_handlers "fratelli-feccia/internal/handlers/pdfimport"
	pdftemplates_handlers "fratelli-feccia/internal/handlers/pdftemplates"
	pricelists_handlers "fratelli-feccia/internal/handlers/pricelists"
	products_handlers "fratelli-feccia/internal/handlers/products"
	semirimorchi_handlers "fratelli-feccia/internal/handlers/semirimorchi"
	trips_handlers "fratelli-feccia/internal/handlers/trips"
	users_handlers "fratelli-feccia/internal/handlers/users"
	washstations_handlers "fratelli-feccia/internal/handlers/washstations"
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
	WashStations         *washstations_handlers.WashStationHandler
	Drivers              *drivers_handlers.DriverHandler
	Products             *products_handlers.ProductHandler
	Masterdata           *masterdata_handlers.MasterdataHandler
	Anagrafiche          *anagrafiche_handlers.AnagraficheHandler
	DriverUnavailability *driverunavailability_handlers.DriverUnavailabilityHandler
	Orders               *orders_handlers.OrderHandler
	Motrici              *motrici_handlers.MotriceHandler
	Semirimorchi         *semirimorchi_handlers.SemirimorchioHandler
	Trips                *trips_handlers.TripHandler
	PriceLists           *pricelists_handlers.PriceListHandler
	Invoices             *invoices_handlers.InvoiceHandler
	Dashboard            *dashboard_handlers.DashboardHandler
	Map                  *mapview_handlers.MapHandler
	Availability         *availability_handlers.AvailabilityHandler
	Export               *export_handlers.ExportHandler
	Geocode              *geocode_handlers.GeocodeHandler
	PdfTemplates         *pdftemplates_handlers.PdfTemplateHandler
	PdfImport            *pdfimport_handlers.PdfImportHandler
	InboundOrders        *inboundorders_handlers.InboundOrderHandler
}

func NewHandler(services *services.Service, auditLogger *audit.Logger, jwtCfg utils.JWTConfig) *Handler {
	return &Handler{
		Auth:                 auth_handlers.NewAuthHandler(services.Authentication.Auth, auditLogger, jwtCfg),
		Admin:                users_handlers.NewUserHandler(services.Admin.Admin),
		Customers:            customers_handlers.NewCustomerHandler(services.Customers.Customer),
		Destinations:         destinations_handlers.NewDestinationHandler(services.Destinations.Destination),
		Carriers:             carriers_handlers.NewCarrierHandler(services.Carriers.Carrier),
		Garages:              garages_handlers.NewGarageHandler(services.Garages.Garage),
		WashStations:         washstations_handlers.NewWashStationHandler(services.WashStations.WashStation),
		Drivers:              drivers_handlers.NewDriverHandler(services.Drivers.Driver),
		Products:             products_handlers.NewProductHandler(services.Products.Product),
		Masterdata:           masterdata_handlers.NewMasterdataHandler(services.MasterdataGroup.Masterdata),
		Anagrafiche:          anagrafiche_handlers.NewAnagraficheHandler(services.AnagraficheGroup.Anagrafiche),
		DriverUnavailability: driverunavailability_handlers.NewDriverUnavailabilityHandler(services.DriverUnavailabilityGroup.DriverUnavailability),
		Orders:               orders_handlers.NewOrderHandler(services.Orders.Order),
		Motrici:              motrici_handlers.NewMotriceHandler(services.Motrici.Motrice),
		Semirimorchi:         semirimorchi_handlers.NewSemirimorchioHandler(services.Semirimorchi.Semirimorchio),
		Trips:                trips_handlers.NewTripHandler(services.Trips.Trip),
		PriceLists:           pricelists_handlers.NewPriceListHandler(services.PriceLists.PriceList),
		Invoices:             invoices_handlers.NewInvoiceHandler(services.Invoices.Invoice),
		Dashboard:            dashboard_handlers.NewDashboardHandler(services.DashboardGroup.Dashboard),
		Map:                  mapview_handlers.NewMapHandler(services.MapGroup.Map),
		Availability:         availability_handlers.NewAvailabilityHandler(services.AvailabilityGroup.Availability),
		Export:               export_handlers.NewExportHandler(services.ExportGroup.Export),
		Geocode:              geocode_handlers.NewGeocodeHandler(services.GeocodeGroup.Geocode),
		PdfTemplates:         pdftemplates_handlers.NewPdfTemplateHandler(services.PdfTemplates.PdfTemplate),
		PdfImport:            pdfimport_handlers.NewPdfImportHandler(services.PdfTemplates.PdfTemplate, services.PdfEngineGroup.PdfEngine),
		InboundOrders:        inboundorders_handlers.NewInboundOrderHandler(services.InboundOrders.InboundOrder, services.MailScraperGroup.MailScraper, services.PdfEngineGroup.PdfEngine),
	}
}
