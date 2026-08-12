package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/internal/services/mocks"
)

func newCustomerTestApp(svc services.Customer) *fiber.App {
	app := fiber.New()
	h := NewCustomerHandler(svc)

	app.Get("/customers", h.ListCustomers)
	app.Get("/customers/:id", h.GetCustomerByID)
	app.Post("/customers", h.CreateCustomer)
	app.Put("/customers/:id", h.UpdateCustomer)
	app.Delete("/customers/:id", h.DeleteCustomer)

	return app
}

func doCustomerRequest(t *testing.T, app *fiber.App, method, path string, body interface{}) *http.Response {
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

func TestCustomerHandler_List_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockCustomer(ctrl)
	mockSvc.EXPECT().List(gomock.Any(), "acme", gomock.Any()).Return([]dto.CustomerResponse{{RagioneSociale: "Acme S.r.l."}}, int64(1), nil)

	app := newCustomerTestApp(mockSvc)
	resp := doCustomerRequest(t, app, http.MethodGet, "/customers?search=acme", nil)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestCustomerHandler_GetByID_InvalidID(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockCustomer(ctrl)
	app := newCustomerTestApp(mockSvc)

	resp := doCustomerRequest(t, app, http.MethodGet, "/customers/not-a-uuid", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestCustomerHandler_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	id := uuid.New()
	mockSvc := mocks.NewMockCustomer(ctrl)
	mockSvc.EXPECT().GetByID(gomock.Any(), id).Return(nil, nil)

	app := newCustomerTestApp(mockSvc)
	resp := doCustomerRequest(t, app, http.MethodGet, "/customers/"+id.String(), nil)

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestCustomerHandler_Create_ValidationError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockCustomer(ctrl)
	app := newCustomerTestApp(mockSvc)

	resp := doCustomerRequest(t, app, http.MethodPost, "/customers", map[string]interface{}{})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestCustomerHandler_Create_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req := dto.CustomerRequest{RagioneSociale: "Acme S.r.l."}
	mockSvc := mocks.NewMockCustomer(ctrl)
	mockSvc.EXPECT().Create(gomock.Any(), req).Return(&dto.CustomerResponse{RagioneSociale: "Acme S.r.l."}, nil)

	app := newCustomerTestApp(mockSvc)
	resp := doCustomerRequest(t, app, http.MethodPost, "/customers", req)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
}

func TestCustomerHandler_Update_NotFound(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	id := uuid.New()
	req := dto.CustomerRequest{RagioneSociale: "Acme S.r.l."}
	mockSvc := mocks.NewMockCustomer(ctrl)
	mockSvc.EXPECT().Update(gomock.Any(), id, req).Return(nil, gorm.ErrRecordNotFound)

	app := newCustomerTestApp(mockSvc)
	resp := doCustomerRequest(t, app, http.MethodPut, "/customers/"+id.String(), req)

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestCustomerHandler_Delete_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	id := uuid.New()
	mockSvc := mocks.NewMockCustomer(ctrl)
	mockSvc.EXPECT().Delete(gomock.Any(), id).Return(nil)

	app := newCustomerTestApp(mockSvc)
	resp := doCustomerRequest(t, app, http.MethodDelete, "/customers/"+id.String(), nil)

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
}
