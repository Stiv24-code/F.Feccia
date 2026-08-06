package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/internal/services/mocks"
)

func newDestinationMineTestApp(svc services.Destination) *fiber.App {
	app := fiber.New()
	h := NewDestinationHandler(svc)
	app.Post("/me/destinations", h.CreateMyDestination)
	return app
}

func doDestinationRequest(t *testing.T, app *fiber.App, method, path string, body interface{}) *http.Response {
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

func TestDestinationHandler_CreateMyDestination_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lat, lng := 45.4642, 9.1900
	req := dto.DestinationRequest{Nome: "Magazzino Cliente", Lat: &lat, Lng: &lng}
	mockSvc := mocks.NewMockDestination(ctrl)
	mockSvc.EXPECT().Create(gomock.Any(), req).Return(&dto.DestinationResponse{Nome: req.Nome}, nil)

	app := newDestinationMineTestApp(mockSvc)
	resp := doDestinationRequest(t, app, http.MethodPost, "/me/destinations", req)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
}

func TestDestinationHandler_CreateMyDestination_ValidationError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockDestination(ctrl)
	app := newDestinationMineTestApp(mockSvc)

	// Missing required lat/lng.
	resp := doDestinationRequest(t, app, http.MethodPost, "/me/destinations", dto.DestinationRequest{Nome: "Senza coordinate"})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}
