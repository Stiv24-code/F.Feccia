package router

import (
	app_handlers "fratelli-feccia/internal/handlers"
	"fratelli-feccia/pkg/middleware"
	"fratelli-feccia/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// registerPdfTemplateRoutes exposes the OrderMesh PDF import flow: template
// CRUD/matching plus the stateless render/test/import endpoints. Staff-only
// (registered under authAll), unlike OrderMesh where the standalone app had
// no auth at all. In OrderMesh these lived at /api/templates and /api/pdf/*.
func registerPdfTemplateRoutes(authAll fiber.Router, handlers *app_handlers.Handler) {
	authAll.Get("/pdf-templates", handlers.PdfTemplates.ListPdfTemplates)
	authAll.Get("/pdf-templates/match", handlers.PdfTemplates.MatchPdfTemplate)

	writeGroup := authAll.Group("/pdf-templates", middleware.RequireRole(utils.RoleAdmin, utils.RoleOperatore))
	writeGroup.Post("", handlers.PdfTemplates.CreatePdfTemplate)
	writeGroup.Put("/:id", handlers.PdfTemplates.UpdatePdfTemplate)
	writeGroup.Delete("/:id", handlers.PdfTemplates.DeletePdfTemplate)

	// Render/test/import never persist anything — any staff role may use them.
	authAll.Post("/pdf/render", handlers.PdfImport.RenderPdf)
	authAll.Post("/pdf/test", handlers.PdfImport.TestPdfTemplate)
	authAll.Post("/pdf/import", handlers.PdfImport.ImportPdf)
}
