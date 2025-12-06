package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/adapter/sendgrid_adapter"
	"liquiswiss/internal/api"
	"liquiswiss/internal/mocks"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
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

func TestFinishRegistrationCreatesOrganisationAndScenario(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	// Create a registration
	registrationID, err := dbAdapter.CreateRegistration("newuser@example.com", "test-code-123")
	assert.NoError(t, err)
	assert.NotEqual(t, int64(0), registrationID)

	// Finish registration
	user, accessToken, accessExpirationTime, refreshToken, refreshExpirationTime, err := apiService.FinishRegistration(
		models.FinishRegistration{
			Email:    "newuser@example.com",
			Code:     "test-code-123",
			Password: "securepassword123",
		},
		"Test Device",
		utils.RegistrationCodeValidity,
	)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotNil(t, accessToken)
	assert.NotNil(t, accessExpirationTime)
	assert.NotNil(t, refreshToken)
	assert.NotNil(t, refreshExpirationTime)

	// Verify user was created
	assert.NotEqual(t, int64(0), user.ID)
	assert.Equal(t, "newuser@example.com", user.Email)

	// Verify user has an organisation assigned
	assert.NotNil(t, user.CurrentOrganisationID)
	assert.NotEqual(t, int64(0), user.CurrentOrganisationID)

	// Verify the organisation exists and has the correct name
	organisations, totalCount, err := dbAdapter.ListOrganisations(user.ID, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), totalCount)
	assert.Equal(t, 1, len(organisations))
	assert.Equal(t, "Meine Organisation", organisations[0].Name)
	assert.Equal(t, user.CurrentOrganisationID, organisations[0].ID)
	assert.True(t, organisations[0].IsDefault)

	// Verify user has a scenario assigned
	assert.NotNil(t, user.CurrentScenarioID)
	assert.NotEqual(t, int64(0), user.CurrentScenarioID)

	// Verify the scenario exists and has the correct name
	scenarios, err := apiService.ListScenarios(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(scenarios))
	assert.Equal(t, "Standardszenario", scenarios[0].Name)
	assert.Equal(t, user.CurrentScenarioID, scenarios[0].ID)
	assert.True(t, scenarios[0].IsDefault)
}
