package handlers

import (
	"github.com/gofiber/fiber/v2"

	"fratelli-feccia/internal/services"
	"fratelli-feccia/internal/services/export"
	"fratelli-feccia/pkg/utils"
)

type ExportHandler struct {
	Service services.Export
}

func NewExportHandler(service services.Export) *ExportHandler {
	return &ExportHandler{Service: service}
}

// @Summary Export orders to xlsx
// @Tags Export
// @Security BearerAuth
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param stato query string false "Filter by stato"
// @Param data_da query string false "data_ritiro >= (YYYY-MM-DD)"
// @Param data_a query string false "data_ritiro <= (YYYY-MM-DD)"
// @Success 200 {file} binary
// @Router /api/v1/export/orders [get]
func (h *ExportHandler) Orders(c *fiber.Ctx) error {
	filter := export.OrdersFilter{Stato: c.Query("stato"), DataDa: c.Query("data_da"), DataA: c.Query("data_a")}
	data, err := h.Service.OrdersExcel(utils.RequestContext(c), filter)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	c.Set(fiber.HeaderContentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set(fiber.HeaderContentDisposition, "attachment; filename=ordini.xlsx")
	return c.Send(data)
}
