package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"liquiswiss/internal/api"
	"liquiswiss/internal/mocks"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestRegistrationSuccessful(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup Mocks
	mockDBService := mocks.NewMockIDatabaseAdapter(ctrl)
	mockEmailService := mocks.NewMockIEmailAdapter(ctrl)

	apiService := api_service.NewAPIService(mockDBService, mockEmailService)

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
	mockEmailService.EXPECT().
		SendRegistrationMail("test@example.com", gomock.AssignableToTypeOf("string")).
		Return(nil)

	// Initialize the API struct with the mocked service
	myAPI := api.NewAPI(mockDBService, apiService, mockEmailService)

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
	mockEmailService := mocks.NewMockIEmailAdapter(ctrl)

	apiService := api_service.NewAPIService(mockDBService, mockEmailService)

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
	myAPI := api.NewAPI(mockDBService, apiService, mockEmailService)

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
	mockEmailService := mocks.NewMockIEmailAdapter(ctrl)

	apiService := api_service.NewAPIService(mockDBService, mockEmailService)

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
	mockEmailService.EXPECT().
		SendRegistrationMail("test@example.com", gomock.AssignableToTypeOf("string")).
		Return(errors.New("error sending email"))
	mockDBService.EXPECT().
		DeleteRegistration(int64(2001), gomock.AssignableToTypeOf("string")).
		Return(nil)

	// Initialize the API struct with the mocked service
	myAPI := api.NewAPI(mockDBService, apiService, mockEmailService)

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

// Failed logins must answer 401, never 500: a 500 signals a server fault and
// (paired with a 401 for unknown emails) would leak which accounts exist.
func TestLoginWrongPasswordReturnsUnauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDBService := mocks.NewMockIDatabaseAdapter(ctrl)
	mockEmailService := mocks.NewMockIEmailAdapter(ctrl)
	apiService := api_service.NewAPIService(mockDBService, mockEmailService)

	// bcrypt hash of "correct-password"
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	assert.NoError(t, err)
	mockDBService.EXPECT().
		GetUserPasswordByEMail("user@example.com").
		Return(&models.Login{ID: 1, Email: "user@example.com", Password: string(hash)}, nil)

	payloadBytes, err := json.Marshal(map[string]string{
		"email":    "user@example.com",
		"password": "wrong-password",
	})
	assert.NoError(t, err)

	myAPI := api.NewAPI(mockDBService, apiService, mockEmailService)
	req, err := http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(payloadBytes))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	myAPI.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Unknown emails must be indistinguishable from wrong passwords (no user enumeration)
func TestLoginUnknownEmailReturnsUnauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDBService := mocks.NewMockIDatabaseAdapter(ctrl)
	mockEmailService := mocks.NewMockIEmailAdapter(ctrl)
	apiService := api_service.NewAPIService(mockDBService, mockEmailService)

	mockDBService.EXPECT().
		GetUserPasswordByEMail("nobody@example.com").
		Return(nil, sql.ErrNoRows)

	payloadBytes, err := json.Marshal(map[string]string{
		"email":    "nobody@example.com",
		"password": "irrelevant",
	})
	assert.NoError(t, err)

	myAPI := api.NewAPI(mockDBService, apiService, mockEmailService)
	req, err := http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(payloadBytes))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	myAPI.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// A genuine infrastructure failure must still surface as 500
func TestLoginDatabaseErrorReturnsInternalServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDBService := mocks.NewMockIDatabaseAdapter(ctrl)
	mockEmailService := mocks.NewMockIEmailAdapter(ctrl)
	apiService := api_service.NewAPIService(mockDBService, mockEmailService)

	mockDBService.EXPECT().
		GetUserPasswordByEMail("user@example.com").
		Return(nil, errors.New("database is on fire"))

	payloadBytes, err := json.Marshal(map[string]string{
		"email":    "user@example.com",
		"password": "irrelevant",
	})
	assert.NoError(t, err)

	myAPI := api.NewAPI(mockDBService, apiService, mockEmailService)
	req, err := http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(payloadBytes))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	myAPI.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
