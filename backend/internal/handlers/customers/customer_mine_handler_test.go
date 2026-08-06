package handlers

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/internal/services/mocks"
)

// newCustomerMineTestApp wires the /me/anagrafica routes behind a stub
// middleware that injects "customer_id" into Locals — the same Local that
// middleware.JWTAuthMiddleware sets in production from the JWT claim (see
// pkg/utils/context.go's RequestCustomerID).
func newCustomerMineTestApp(svc services.Customer, customerID string) *fiber.App {
	app := fiber.New()
	h := NewCustomerHandler(svc)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("customer_id", customerID)
		return c.Next()
	})
	app.Get("/me/anagrafica", h.GetMyAnagrafica)
	app.Put("/me/anagrafica", h.UpdateMyAnagrafica)

	return app
}

func TestCustomerHandler_GetMyAnagrafica_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	customerID := uuid.New()
	mockSvc := mocks.NewMockCustomer(ctrl)
	mockSvc.EXPECT().GetByID(gomock.Any(), customerID).Return(&dto.CustomerResponse{ID: customerID, RagioneSociale: "Acme S.r.l."}, nil)

	app := newCustomerMineTestApp(mockSvc, customerID.String())
	resp := doCustomerRequest(t, app, http.MethodGet, "/me/anagrafica", nil)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestCustomerHandler_GetMyAnagrafica_MissingCustomerIDClaim(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// No customer_id claim (e.g. a malformed token) — must fail closed, never
	// fall through to a service call.
	mockSvc := mocks.NewMockCustomer(ctrl)
	app := newCustomerMineTestApp(mockSvc, "")

	resp := doCustomerRequest(t, app, http.MethodGet, "/me/anagrafica", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestCustomerHandler_UpdateMyAnagrafica_UsesOwnCustomerIDNotAnyPathOrBodyValue(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	customerID := uuid.New()
	req := dto.CustomerRequest{RagioneSociale: "Nuova Ragione Sociale"}
	mockSvc := mocks.NewMockCustomer(ctrl)
	// The route has no :id path param at all — this also asserts the
	// service is called with the claim's id, not anything client-supplied.
	mockSvc.EXPECT().Update(gomock.Any(), customerID, req).Return(&dto.CustomerResponse{ID: customerID, RagioneSociale: req.RagioneSociale}, nil)

	app := newCustomerMineTestApp(mockSvc, customerID.String())
	resp := doCustomerRequest(t, app, http.MethodPut, "/me/anagrafica", req)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
