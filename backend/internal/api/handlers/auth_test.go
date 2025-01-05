package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"liquiswiss/internal/api"
	"liquiswiss/internal/service/mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRegistrationSuccessful(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock instance of the IDatabaseService
	mockDBService := mocks.NewMockIDatabaseService(ctrl)
	mockSendgridService := mocks.NewMockISendgridService(ctrl)
	mockForecastService := mocks.NewMockIForecastService(ctrl)

	// Prepare the payload for the registration request
	payload := map[string]string{
		"email": "test@example.com",
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Set up expectations for the mock
	mockDBService.EXPECT().
		CreateRegistration("test@example.com", gomock.AssignableToTypeOf("string")).
		Return(int64(2001), nil)
	mockSendgridService.EXPECT().
		SendRegistrationMail("test@example.com", gomock.AssignableToTypeOf("string")).
		Return(nil)

	// Initialize the API struct with the mocked service
	myAPI := api.NewAPI(mockDBService, mockSendgridService, mockForecastService)

	// Create a request to the registration endpoint
	req, err := http.NewRequest(http.MethodPost, "/api/auth/registration/create", bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform the request using the Gin engine from the API struct
	w := httptest.NewRecorder()
	myAPI.Router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusOK)
}

func TestRegistrationCreationFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock instance of the IDatabaseService
	mockDBService := mocks.NewMockIDatabaseService(ctrl)
	mockSendgridService := mocks.NewMockISendgridService(ctrl)
	mockForecastService := mocks.NewMockIForecastService(ctrl)

	// Prepare the payload for the registration request
	payload := map[string]string{
		"email": "test@example.com",
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Set up expectations for the mock
	mockDBService.EXPECT().
		CreateRegistration("test@example.com", gomock.AssignableToTypeOf("string")).
		Return(int64(0), errors.New("creation error occurred"))

	// Initialize the API struct with the mocked service
	myAPI := api.NewAPI(mockDBService, mockSendgridService, mockForecastService)

	// Create a request to the registration endpoint
	req, err := http.NewRequest(http.MethodPost, "/api/auth/registration/create", bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform the request using the Gin engine from the API struct
	w := httptest.NewRecorder()
	myAPI.Router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusInternalServerError)
}

func TestRegistrationEmailFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock instance of the IDatabaseService
	mockDBService := mocks.NewMockIDatabaseService(ctrl)
	mockSendgridService := mocks.NewMockISendgridService(ctrl)
	mockForecastService := mocks.NewMockIForecastService(ctrl)

	// Prepare the payload for the registration request
	payload := map[string]string{
		"email": "test@example.com",
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Set up expectations for the mock
	mockDBService.EXPECT().
		CreateRegistration("test@example.com", gomock.AssignableToTypeOf("string")).
		Return(int64(2001), nil)
	mockSendgridService.EXPECT().
		SendRegistrationMail("test@example.com", gomock.AssignableToTypeOf("string")).
		Return(errors.New("error sending email"))
	mockDBService.EXPECT().
		DeleteRegistration(int64(2001), gomock.AssignableToTypeOf("string")).
		Return(nil)

	// Initialize the API struct with the mocked service
	myAPI := api.NewAPI(mockDBService, mockSendgridService, mockForecastService)

	// Create a request to the registration endpoint
	req, err := http.NewRequest(http.MethodPost, "/api/auth/registration/create", bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform the request using the Gin engine from the API struct
	w := httptest.NewRecorder()
	myAPI.Router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusInternalServerError)
}
