package handlers_test

import (
	"bytes"
	"encoding/json"
	"liquiswiss/internal/api"
	"liquiswiss/internal/mocks"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestMain(m *testing.M) {
	utils.InitValidator()
	code := m.Run()
	os.Exit(code)
}

func TestRegisterSuccessfully(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock instance of the IDatabaseService
	mockDBService := mocks.NewMockIDatabaseService(ctrl)

	// Set up expectations for the mock
	mockDBService.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).Return(int64(1), nil)
	mockDBService.EXPECT().GetProfile("1").Return(&models.User{
		ID:    1,
		Email: "test@example.com",
		Name:  "",
	}, nil)

	// Initialize the API struct with the mocked service
	myAPI := api.NewAPI(mockDBService)

	// Prepare the payload for the registration request
	payload := map[string]string{
		"email":    "test@example.com",
		"password": "securepassword",
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Create a request to the register endpoint
	req, err := http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform the request using the Gin engine from the API struct
	w := httptest.NewRecorder()
	myAPI.Router.ServeHTTP(w, req)

	// Check the response status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK; got %v, body: %v", w.Code, w.Body.String())
	}

	var user models.User
	err = json.Unmarshal(w.Body.Bytes(), &user)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	assert.Equal(t, "test@example.com", user.Email)
}
