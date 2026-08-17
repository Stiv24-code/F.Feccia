package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/internal/services/mocks"
)

// newInboundOrderMineTestApp mirrors order_mine_handler_test.go's
// newOrderMineTestApp: a stub middleware injects "customer_id" into Locals,
// the same claim middleware.JWTAuthMiddleware sets from the JWT in production.
func newInboundOrderMineTestApp(svc services.InboundOrder, customers services.Customer, destinations services.Destination, customerID string) *fiber.App {
	app := fiber.New()
	h := NewInboundOrderHandler(svc, nil, nil, customers, destinations)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("customer_id", customerID)
		return c.Next()
	})
	app.Get("/me/inbound-orders", h.ListMyInboundOrders)
	app.Post("/me/inbound-orders", h.CreateMyInboundOrder)

	return app
}

func doInboundRequest(t *testing.T, app *fiber.App, method, path string, body interface{}) *http.Response {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("failed to encode body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	return resp
}

func TestInboundOrderHandler_CreateMyInboundOrder_ForcesOwnClienteIDAndPortalSource(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	myCustomerID := uuid.New()
	var captured dto.InboundOrderRequest

	mockInbound := mocks.NewMockInboundOrder(ctrl)
	mockCustomers := mocks.NewMockCustomer(ctrl)
	mockDestinations := mocks.NewMockDestination(ctrl)

	mockCustomers.EXPECT().GetByID(gomock.Any(), myCustomerID).Return(&dto.CustomerResponse{ID: myCustomerID, RagioneSociale: "ACME S.p.A.", Email: "acme@example.com"}, nil)
	mockInbound.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, req dto.InboundOrderRequest) (*dto.InboundOrderResponse, error) {
		captured = req
		return &dto.InboundOrderResponse{ClienteID: req.ClienteID, Source: req.Source, Client: req.Client}, nil
	})

	app := newInboundOrderMineTestApp(mockInbound, mockCustomers, mockDestinations, myCustomerID.String())
	// No destinazione ids in the body — the handler must still fall back
	// gracefully (no Destinations.GetByID call expected) instead of erroring.
	resp := doInboundRequest(t, app, http.MethodPost, "/me/inbound-orders", dto.ClientInboundOrderRequest{
		OrderRequest: dto.OrderRequest{Tariffa: 1200, Note: "Test"},
		Product:      "Pasta alimentare",
		Kg:           9462,
	})

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
	if captured.ClienteID == nil || *captured.ClienteID != myCustomerID {
		t.Fatalf("expected ClienteID forced to %q, got %v", myCustomerID, captured.ClienteID)
	}
	if captured.Source != models.InboundOrderSourcePortal {
		t.Fatalf("expected Source %q, got %q", models.InboundOrderSourcePortal, captured.Source)
	}
	if captured.Product != "Pasta alimentare" || captured.Kg != 9462 {
		t.Fatalf("expected client-supplied Product/Kg to pass through, got %q/%d", captured.Product, captured.Kg)
	}
	if captured.Client != "ACME S.p.A." {
		t.Fatalf("expected Client resolved from the authenticated customer, got %q", captured.Client)
	}
}

func TestInboundOrderHandler_CreateMyInboundOrder_MissingCustomerIDClaim(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	app := newInboundOrderMineTestApp(mocks.NewMockInboundOrder(ctrl), mocks.NewMockCustomer(ctrl), mocks.NewMockDestination(ctrl), "")
	resp := doInboundRequest(t, app, http.MethodPost, "/me/inbound-orders", dto.OrderRequest{})

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestInboundOrderHandler_ListMyInboundOrders_ScopesByOwnCustomerID(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	customerID := uuid.New()
	mockInbound := mocks.NewMockInboundOrder(ctrl)
	mockInbound.EXPECT().ListForClient(gomock.Any(), customerID).Return([]dto.InboundOrderResponse{{ClienteID: &customerID}}, nil)

	app := newInboundOrderMineTestApp(mockInbound, mocks.NewMockCustomer(ctrl), mocks.NewMockDestination(ctrl), customerID.String())
	resp := doInboundRequest(t, app, http.MethodGet, "/me/inbound-orders", nil)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
