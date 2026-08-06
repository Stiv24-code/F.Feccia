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
	"fratelli-feccia/internal/services"
	"fratelli-feccia/internal/services/mocks"
	"fratelli-feccia/internal/services/orders"
)

// newOrderMineTestApp wires the /me/orders routes behind a stub middleware
// that injects "customer_id" into Locals, mirroring what
// middleware.JWTAuthMiddleware sets in production from the JWT claim.
func newOrderMineTestApp(svc services.Order, customerID string) *fiber.App {
	app := fiber.New()
	h := NewOrderHandler(svc)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("customer_id", customerID)
		return c.Next()
	})
	app.Get("/me/orders", h.ListMyOrders)
	app.Get("/me/orders/:id", h.GetMyOrderByID)
	app.Post("/me/orders", h.CreateMyOrder)
	app.Delete("/me/orders/:id", h.DeleteMyOrder)

	return app
}

func doOrderRequest(t *testing.T, app *fiber.App, method, path string, body interface{}) *http.Response {
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

func TestOrderHandler_ListMyOrders_ScopesByOwnCustomerID(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	customerID := uuid.New()
	mockSvc := mocks.NewMockOrder(ctrl)
	mockSvc.EXPECT().List(gomock.Any(), orders.ListFilters{ClienteID: customerID.String()}).Return([]dto.OrderResponse{{ClienteID: customerID.String()}}, nil)

	app := newOrderMineTestApp(mockSvc, customerID.String())
	resp := doOrderRequest(t, app, http.MethodGet, "/me/orders", nil)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestOrderHandler_ListMyOrders_MissingCustomerIDClaim(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockOrder(ctrl)
	app := newOrderMineTestApp(mockSvc, "")

	resp := doOrderRequest(t, app, http.MethodGet, "/me/orders", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestOrderHandler_GetMyOrderByID_ForeignOrderReturns404(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	myCustomerID := uuid.New()
	otherCustomerID := uuid.New()
	orderID := uuid.New()
	mockSvc := mocks.NewMockOrder(ctrl)
	mockSvc.EXPECT().GetByID(gomock.Any(), orderID).Return(&dto.OrderResponse{ID: orderID, ClienteID: otherCustomerID.String()}, nil)

	app := newOrderMineTestApp(mockSvc, myCustomerID.String())
	resp := doOrderRequest(t, app, http.MethodGet, "/me/orders/"+orderID.String(), nil)

	// 404, not 403 — must not confirm to the caller that an order belonging
	// to a different customer exists.
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestOrderHandler_GetMyOrderByID_OwnOrderReturns200(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	customerID := uuid.New()
	orderID := uuid.New()
	mockSvc := mocks.NewMockOrder(ctrl)
	mockSvc.EXPECT().GetByID(gomock.Any(), orderID).Return(&dto.OrderResponse{ID: orderID, ClienteID: customerID.String()}, nil)

	app := newOrderMineTestApp(mockSvc, customerID.String())
	resp := doOrderRequest(t, app, http.MethodGet, "/me/orders/"+orderID.String(), nil)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestOrderHandler_CreateMyOrder_ForcesOwnClienteIDIgnoringBody(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	myCustomerID := uuid.New()
	someoneElsesID := uuid.New()
	var capturedClienteID string

	mockSvc := mocks.NewMockOrder(ctrl)
	mockSvc.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, req dto.OrderRequest) (*dto.OrderResponse, error) {
		capturedClienteID = req.ClienteID
		return &dto.OrderResponse{ClienteID: req.ClienteID}, nil
	})

	app := newOrderMineTestApp(mockSvc, myCustomerID.String())
	// Body tries to create the order under a different customer — must be
	// silently overridden, never trusted.
	resp := doOrderRequest(t, app, http.MethodPost, "/me/orders", dto.OrderRequest{ClienteID: someoneElsesID.String(), Tariffa: 100})

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
	if capturedClienteID != myCustomerID.String() {
		t.Fatalf("expected ClienteID forced to %q, got %q", myCustomerID.String(), capturedClienteID)
	}
}

func TestOrderHandler_DeleteMyOrder_ForeignOrderReturns404AndNeverCallsDelete(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	myCustomerID := uuid.New()
	otherCustomerID := uuid.New()
	orderID := uuid.New()
	mockSvc := mocks.NewMockOrder(ctrl)
	mockSvc.EXPECT().GetByID(gomock.Any(), orderID).Return(&dto.OrderResponse{ID: orderID, ClienteID: otherCustomerID.String()}, nil)
	// No .EXPECT().Delete(...) at all — gomock fails the test if Delete is
	// called unexpectedly, proving ownership is checked before deleting.

	app := newOrderMineTestApp(mockSvc, myCustomerID.String())
	resp := doOrderRequest(t, app, http.MethodDelete, "/me/orders/"+orderID.String(), nil)

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestOrderHandler_DeleteMyOrder_OwnOrderDeletes(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	customerID := uuid.New()
	orderID := uuid.New()
	mockSvc := mocks.NewMockOrder(ctrl)
	mockSvc.EXPECT().GetByID(gomock.Any(), orderID).Return(&dto.OrderResponse{ID: orderID, ClienteID: customerID.String()}, nil)
	mockSvc.EXPECT().Delete(gomock.Any(), orderID).Return(nil)

	app := newOrderMineTestApp(mockSvc, customerID.String())
	resp := doOrderRequest(t, app, http.MethodDelete, "/me/orders/"+orderID.String(), nil)

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
}
