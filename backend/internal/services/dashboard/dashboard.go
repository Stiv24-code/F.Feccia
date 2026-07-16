package dashboard

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/internal/services/orders"
	"fratelli-feccia/pkg/utils"
)

type DashboardService struct {
	db *gorm.DB
}

func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{db: db}
}

// Stats mirrors GET /dashboard/stats. The monthly trend intentionally
// replicates Python's own pipeline verbatim, quirks included: it sorts
// months ASCENDING then takes the first 12 (earliest months, not most
// recent) and doesn't exclude an empty data_ritiro group — not "fixed"
// here since that would silently change reported figures.
func (s *DashboardService) Stats(ctx context.Context) (*dto.DashboardStatsResponse, error) {
	db := s.db.WithContext(ctx)

	var totalOrders, pianificabili, inViaggio, chiusi, fatturati int64
	db.Model(&models.Order{}).Count(&totalOrders)
	db.Model(&models.Order{}).Where("stato = ?", "PIANIFICABILE").Count(&pianificabili)
	db.Model(&models.Order{}).Where("stato = ?", "VIAGGIO").Count(&inViaggio)
	db.Model(&models.Order{}).Where("stato = ?", "CHIUSO").Count(&chiusi)
	db.Model(&models.Order{}).Where("stato = ?", "FATTURATO").Count(&fatturati)

	var totalCustomers, totalVehicles, totalDrivers int64
	db.Model(&models.Customer{}).Where("active = ?", true).Count(&totalCustomers)
	db.Model(&models.Vehicle{}).Where("active = ?", true).Count(&totalVehicles)
	db.Model(&models.Driver{}).Where("active = ?", true).Count(&totalDrivers)

	var totalRevenue float64
	if err := db.Model(&models.Invoice{}).Where("stato = ?", "DEFINITIVA").
		Select("COALESCE(SUM(totale), 0)").Scan(&totalRevenue).Error; err != nil {
		return nil, err
	}

	var monthly []dto.MonthlyOrderTrend
	if err := db.Model(&models.Order{}).
		Select("SUBSTR(data_ritiro, 1, 7) AS mese, COUNT(*) AS ordini, COALESCE(SUM(tariffa), 0) AS totale").
		Group("SUBSTR(data_ritiro, 1, 7)").
		Order("mese ASC").Limit(12).Scan(&monthly).Error; err != nil {
		return nil, err
	}
	if monthly == nil {
		monthly = []dto.MonthlyOrderTrend{}
	}

	return &dto.DashboardStatsResponse{
		TotalOrders: totalOrders, Pianificabili: pianificabili, InViaggio: inViaggio,
		Chiusi: chiusi, Fatturati: fatturati, TotalCustomers: totalCustomers,
		TotalVehicles: totalVehicles, TotalDrivers: totalDrivers, TotalRevenue: totalRevenue,
		MonthlyTrend: monthly,
	}, nil
}

// CustomerDashboard mirrors GET /dashboard/customer/{id}: KPIs, monthly
// trend (most recent 12 months, ascending), top-5 destinations and
// tipologia/categoria distribution — all scoped to one customer's orders.
func (s *DashboardService) CustomerDashboard(ctx context.Context, customerID uuid.UUID) (*dto.CustomerDashboardResponse, error) {
	db := s.db.WithContext(ctx)

	var customer models.Customer
	if err := db.First(&customer, "id = ?", customerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAPIError(404, "Cliente non trovato")
		}
		return nil, err
	}
	clienteID := customerID.String()

	var kpi dto.CustomerDashboardKPI
	if err := db.Model(&models.Order{}).Where("cliente_id = ?", clienteID).
		Select(`COUNT(*) AS ordini_totali,
			COALESCE(SUM(CASE WHEN stato = 'FATTURATO' THEN 1 ELSE 0 END), 0) AS ordini_fatturati,
			COALESCE(SUM(CASE WHEN stato = 'CHIUSO' THEN 1 ELSE 0 END), 0) AS ordini_chiusi,
			COALESCE(SUM(CASE WHEN stato = 'VIAGGIO' THEN 1 ELSE 0 END), 0) AS ordini_in_viaggio,
			COALESCE(SUM(CASE WHEN stato = 'PIANIFICABILE' THEN 1 ELSE 0 END), 0) AS ordini_pianificabili,
			COALESCE(SUM(CASE WHEN stato = 'FATTURATO' THEN tariffa ELSE 0 END), 0) AS fatturato_netto,
			COALESCE(AVG(tariffa), 0) AS tariffa_media`).
		Scan(&kpi).Error; err != nil {
		return nil, err
	}
	kpi.FatturatoNetto = roundTo2(kpi.FatturatoNetto)
	kpi.TariffaMedia = roundTo2(kpi.TariffaMedia)

	var monthly []dto.CustomerDashboardMonthly
	if err := db.Model(&models.Order{}).Where("cliente_id = ? AND SUBSTR(data_ritiro, 1, 7) <> ''", clienteID).
		Select("SUBSTR(data_ritiro, 1, 7) AS mese, COUNT(*) AS ordini, COALESCE(SUM(tariffa), 0) AS fatturato").
		Group("SUBSTR(data_ritiro, 1, 7)").
		Order("mese DESC").Limit(12).Scan(&monthly).Error; err != nil {
		return nil, err
	}
	for i := range monthly {
		monthly[i].Fatturato = roundTo2(monthly[i].Fatturato)
	}
	reverseMonthly(monthly)
	if monthly == nil {
		monthly = []dto.CustomerDashboardMonthly{}
	}

	var topDest []dto.CustomerDashboardDestination
	if err := db.Model(&models.Order{}).Where("cliente_id = ? AND destinazione_scarico_nome <> ''", clienteID).
		Select("destinazione_scarico_nome AS nome, COUNT(*) AS ordini, COALESCE(SUM(tariffa), 0) AS fatturato").
		Group("destinazione_scarico_nome").
		Order("ordini DESC").Limit(5).Scan(&topDest).Error; err != nil {
		return nil, err
	}
	for i := range topDest {
		topDest[i].Fatturato = roundTo2(topDest[i].Fatturato)
	}
	if topDest == nil {
		topDest = []dto.CustomerDashboardDestination{}
	}

	var byTipologia []dto.CustomerDashboardTipologia
	if err := db.Model(&models.Order{}).Where("cliente_id = ?", clienteID).
		Select(`COALESCE(NULLIF(tipologia, ''), '—') AS tipologia, COUNT(*) AS ordini`).
		Group("tipologia").
		Order("ordini DESC").Limit(20).Scan(&byTipologia).Error; err != nil {
		return nil, err
	}
	if byTipologia == nil {
		byTipologia = []dto.CustomerDashboardTipologia{}
	}

	var byCategoria []dto.CustomerDashboardCategoria
	if err := db.Model(&models.Order{}).Where("cliente_id = ? AND categoria_trasporto <> ''", clienteID).
		Select("categoria_trasporto AS categoria, COUNT(*) AS ordini").
		Group("categoria_trasporto").
		Order("ordini DESC").Limit(10).Scan(&byCategoria).Error; err != nil {
		return nil, err
	}
	if byCategoria == nil {
		byCategoria = []dto.CustomerDashboardCategoria{}
	}

	return &dto.CustomerDashboardResponse{
		Customer: dto.CustomerDashboardSummary{
			ID: customer.ID, RagioneSociale: customer.RagioneSociale,
			Citta: customer.Citta, PartitaIva: customer.PartitaIva,
		},
		KPI: kpi, MonthlyTrend: monthly, TopDestinazioni: topDest,
		PerTipologia: byTipologia, PerCategoria: byCategoria,
	}, nil
}

// RecentOrders mirrors GET /dashboard/recent-orders (last 10, newest first).
func (s *DashboardService) RecentOrders(ctx context.Context) ([]dto.OrderResponse, error) {
	var recent []models.Order
	if err := s.db.WithContext(ctx).Preload("Items").Order("created_at DESC").Limit(10).Find(&recent).Error; err != nil {
		return nil, err
	}
	result := make([]dto.OrderResponse, len(recent))
	for i, o := range recent {
		result[i] = orders.ToResponse(o)
	}
	return result, nil
}

func reverseMonthly(s []dto.CustomerDashboardMonthly) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func roundTo2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
