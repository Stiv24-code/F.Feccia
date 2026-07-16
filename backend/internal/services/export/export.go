// Package export ports backend/routers/export.py: orders -> .xlsx, using
// excelize as the Go equivalent of openpyxl.
package export

import (
	"bytes"
	"context"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	"fratelli-feccia/internal/models"
)

const maxExportRows = 5000

type ExportService struct {
	db *gorm.DB
}

func NewExportService(db *gorm.DB) *ExportService {
	return &ExportService{db: db}
}

// OrdersFilter mirrors the query params of GET /export/orders.
type OrdersFilter struct {
	Stato  string
	DataDa string
	DataA  string
}

// OrdersExcel mirrors export_orders_excel: same columns, same order,
// sorted ascending by data_ritiro.
func (s *ExportService) OrdersExcel(ctx context.Context, filter OrdersFilter) ([]byte, error) {
	query := s.db.WithContext(ctx).Model(&models.Order{})
	if filter.Stato != "" {
		query = query.Where("stato = ?", filter.Stato)
	}
	if filter.DataDa != "" {
		query = query.Where("data_ritiro >= ?", filter.DataDa)
	}
	if filter.DataA != "" {
		query = query.Where("data_ritiro <= ?", filter.DataA)
	}

	var orders []models.Order
	if err := query.Order("data_ritiro ASC").Limit(maxExportRows).Find(&orders).Error; err != nil {
		return nil, err
	}

	xf := excelize.NewFile()
	defer xf.Close()
	const sheet = "Ordini"
	xf.SetSheetName("Sheet1", sheet)

	headers := []string{
		"Progressivo", "Cliente", "Carico", "Scarico",
		"Data Ritiro", "Data Consegna", "Tariffa", "Tipologia", "Stato",
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xf.SetCellValue(sheet, cell, h)
	}

	for i, o := range orders {
		row := i + 2
		values := []interface{}{
			o.Progressivo, o.ClienteNome, o.DestinazioneCaricoNome, o.DestinazioneScaricoNome,
			o.DataRitiro, o.DataConsegna, o.Tariffa, o.Tipologia, o.Stato,
		}
		for col, v := range values {
			cell, _ := excelize.CoordinatesToCellName(col+1, row)
			xf.SetCellValue(sheet, cell, v)
		}
	}

	var buf bytes.Buffer
	if err := xf.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
