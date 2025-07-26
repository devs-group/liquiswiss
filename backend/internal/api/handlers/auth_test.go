package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"liquiswiss/internal/api"
	"liquiswiss/internal/mocks"
	"liquiswiss/internal/service/api_service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRegistrationSuccessful(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup Mocks
	mockDBService := mocks.NewMockIDatabaseAdapter(ctrl)
	mockSendgridService := mocks.NewMockISendgridAdapter(ctrl)

	apiService := api_service.NewAPIService(mockDBService, mockSendgridService)

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
	myAPI := api.NewAPI(mockDBService, apiService, mockSendgridService)

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
	mockDBService := mocks.NewMockIDatabaseAdapter(ctrl)
	mockSendgridService := mocks.NewMockISendgridAdapter(ctrl)

	apiService := api_service.NewAPIService(mockDBService, mockSendgridService)

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
	myAPI := api.NewAPI(mockDBService, apiService, mockSendgridService)

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
	mockDBService := mocks.NewMockIDatabaseAdapter(ctrl)
	mockSendgridService := mocks.NewMockISendgridAdapter(ctrl)

	apiService := api_service.NewAPIService(mockDBService, mockSendgridService)

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
	myAPI := api.NewAPI(mockDBService, apiService, mockSendgridService)

	// Create a request to the registration endpoint
	req, err := http.NewRequest(http.MethodPost, "/api/auth/registration/create", bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform the request using the Gin engine from the API struct
	w := httptest.NewRecorder()
	myAPI.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
