package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/adapter/sendgrid_adapter"
	"liquiswiss/internal/api/handlers"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
)

func TestCreateScenarioWithoutParent(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	user, organisation, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "create-scenario@example.com", "secret", "Scenario Org",
	)
	assert.NoError(t, err)

	scenario := createScenarioViaHandler(t, apiService, user.ID, `{"name":"Baseline","isDefault":true}`)

	assert.Equal(t, "Baseline", scenario.Name)
	assert.True(t, scenario.IsDefault)
	assert.Nil(t, scenario.ParentScenarioID)
	assert.Equal(t, organisation.ID, scenario.OrganisationID)
}

func TestCreateScenarioWithParent(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	user, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "create-scenario-parent@example.com", "secret", "Scenario Org",
	)
	assert.NoError(t, err)

	parent := createScenarioViaHandler(t, apiService, user.ID, `{"name":"Baseline","isDefault":true}`)
	childPayload := fmt.Sprintf(`{"name":"Upside","parentScenarioId":%d}`, parent.ID)

	child := createScenarioViaHandler(t, apiService, user.ID, childPayload)

	if assert.NotNil(t, child.ParentScenarioID) {
		assert.Equal(t, parent.ID, *child.ParentScenarioID)
	}
	assert.False(t, child.IsDefault)
}

func TestListScenariosIntegration(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	user, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "list-scenario@example.com", "secret", "Scenario Org",
	)
	assert.NoError(t, err)

	parent := createScenarioViaHandler(t, apiService, user.ID, `{"name":"Baseline","isDefault":true}`)
	createScenarioViaHandler(t, apiService, user.ID, fmt.Sprintf(`{"name":"Upside","parentScenarioId":%d}`, parent.ID))

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodGet, "/scenarios?limit=10&page=1", nil)
	ctx.Request = req
	ctx.Set("userID", user.ID)

	handlers.ListScenarios(apiService, ctx)

	assert.Equal(t, http.StatusOK, ctx.Writer.Status())

	var response models.ListResponse[models.Scenario]
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Data, 3)
	names := make([]string, 0, len(response.Data))
	for _, scenario := range response.Data {
		names = append(names, scenario.Name)
	}
	assert.ElementsMatch(t, []string{"Default", "Baseline", "Upside"}, names)
	assert.Equal(t, int64(3), response.Pagination.TotalCount)
	assert.Equal(t, int64(1), response.Pagination.CurrentPage)
	assert.Equal(t, int64(0), response.Pagination.TotalRemaining)
}

func TestGetScenarioIntegration(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	user, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "get-scenario@example.com", "secret", "Scenario Org",
	)
	assert.NoError(t, err)

	scenario := createScenarioViaHandler(t, apiService, user.ID, `{"name":"Baseline","isDefault":true}`)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/scenarios/%d", scenario.ID), nil)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "scenarioID", Value: fmt.Sprintf("%d", scenario.ID)}}
	ctx.Set("userID", user.ID)

	handlers.GetScenario(apiService, ctx)

	assert.Equal(t, http.StatusOK, ctx.Writer.Status())

	var fetched models.Scenario
	err = json.Unmarshal(recorder.Body.Bytes(), &fetched)
	assert.NoError(t, err)
	assert.Equal(t, scenario.ID, fetched.ID)
	assert.Equal(t, "Baseline", fetched.Name)
}

func TestUpdateScenarioWithoutParent(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	user, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "update-scenario@example.com", "secret", "Scenario Org",
	)
	assert.NoError(t, err)

	scenario := createScenarioViaHandler(t, apiService, user.ID, `{"name":"Baseline","isDefault":true}`)

	updated := updateScenarioViaHandler(t, apiService, user.ID, scenario.ID, `{"name":"Updated Baseline","isDefault":false}`)

	assert.Equal(t, "Updated Baseline", updated.Name)
	assert.False(t, updated.IsDefault)
	assert.Nil(t, updated.ParentScenarioID)
}

func TestUpdateScenarioWithParent(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	user, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "update-scenario-parent@example.com", "secret", "Scenario Org",
	)
	assert.NoError(t, err)

	parent := createScenarioViaHandler(t, apiService, user.ID, `{"name":"Parent","isDefault":true}`)
	child := createScenarioViaHandler(t, apiService, user.ID, `{"name":"Child","isDefault":false}`)

	payload := fmt.Sprintf(`{"parentScenarioId":%d}`, parent.ID)
	updated := updateScenarioViaHandler(t, apiService, user.ID, child.ID, payload)

	if assert.NotNil(t, updated.ParentScenarioID) {
		assert.Equal(t, parent.ID, *updated.ParentScenarioID)
	}
	assert.Equal(t, "Child", updated.Name)
}

func TestUpdateScenarioWithoutChanges(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	user, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "update-scenario-empty@example.com", "secret", "Scenario Org",
	)
	assert.NoError(t, err)

	scenario := createScenarioViaHandler(t, apiService, user.ID, `{"name":"Baseline","isDefault":true}`)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/scenarios/%d", scenario.ID), bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "scenarioID", Value: fmt.Sprintf("%d", scenario.ID)}}
	ctx.Set("userID", user.ID)

	handlers.UpdateScenario(apiService, ctx)

	assert.Equal(t, http.StatusBadRequest, ctx.Writer.Status())
}

func TestDeleteScenarioIntegration(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	user, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "delete-scenario@example.com", "secret", "Scenario Org",
	)
	assert.NoError(t, err)

	scenario := createScenarioViaHandler(t, apiService, user.ID, `{"name":"Baseline","isDefault":false}`)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/scenarios/%d", scenario.ID), nil)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "scenarioID", Value: fmt.Sprintf("%d", scenario.ID)}}
	ctx.Set("userID", user.ID)

	handlers.DeleteScenario(apiService, ctx)

	assert.Equal(t, http.StatusNoContent, ctx.Writer.Status())

	recorder = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(recorder)
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/scenarios/%d", scenario.ID), nil)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "scenarioID", Value: fmt.Sprintf("%d", scenario.ID)}}
	ctx.Set("userID", user.ID)

	handlers.GetScenario(apiService, ctx)

	assert.Equal(t, http.StatusNotFound, ctx.Writer.Status())
}

func TestListScenariosUnauthorized(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodGet, "/scenarios?limit=5&page=1", nil)
	ctx.Request = req

	handlers.ListScenarios(nil, ctx)

	assert.Equal(t, http.StatusUnauthorized, ctx.Writer.Status())
}

func createScenarioViaHandler(t *testing.T, apiService api_service.IAPIService, userID int64, payload string) models.Scenario {
	t.Helper()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPost, "/scenarios", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req
	ctx.Set("userID", userID)

	handlers.CreateScenario(apiService, ctx)

	if ctx.Writer.Status() != http.StatusCreated {
		t.Fatalf("expected status %d, got %d; response: %s", http.StatusCreated, ctx.Writer.Status(), recorder.Body.String())
	}

	var scenario models.Scenario
	if err := json.Unmarshal(recorder.Body.Bytes(), &scenario); err != nil {
		t.Fatalf("failed to unmarshal scenario response: %v", err)
	}

	return scenario
}

func updateScenarioViaHandler(t *testing.T, apiService api_service.IAPIService, userID, scenarioID int64, payload string) models.Scenario {
	t.Helper()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/scenarios/%d", scenarioID), bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "scenarioID", Value: fmt.Sprintf("%d", scenarioID)}}
	ctx.Set("userID", userID)

	handlers.UpdateScenario(apiService, ctx)

	if ctx.Writer.Status() != http.StatusOK {
		t.Fatalf("expected status %d, got %d; response: %s", http.StatusOK, ctx.Writer.Status(), recorder.Body.String())
	}

	var scenario models.Scenario
	if err := json.Unmarshal(recorder.Body.Bytes(), &scenario); err != nil {
		t.Fatalf("failed to unmarshal scenario response: %v", err)
	}

	return scenario
}
