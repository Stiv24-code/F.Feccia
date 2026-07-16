package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerInvoiceRoutes mirrors backend/routers/invoices.py's CRUD +
// finalize + PDF export (write is admin+amministrazione, reads open to any
// authenticated role).
func registerInvoiceRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	requireAmministrazione := middleware.RequireRole(utils.RoleAdmin, utils.RoleAmministrazione)

	authAll.Get("/invoices", handlers.Invoices.ListInvoices)
	authAll.Get("/invoices/:id", handlers.Invoices.GetInvoiceByID)
	authAll.Get("/invoices/:id/pdf", handlers.Invoices.GetInvoicePDF)
	authAll.Get("/invoices/:id/pdf-url", handlers.Invoices.GetInvoicePDFURL)

	authAll.Post("/invoices", requireAmministrazione, handlers.Invoices.CreateInvoice)
	authAll.Patch("/invoices/:id/finalize", requireAmministrazione, handlers.Invoices.FinalizeInvoice)
	authAll.Delete("/invoices/:id", requireAmministrazione, handlers.Invoices.DeleteInvoice)
}
