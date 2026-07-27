package handlers_test

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"liquiswiss/config"
	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/adapter/email_adapter"
	"liquiswiss/internal/api"
	"liquiswiss/internal/events"
	"liquiswiss/internal/middleware"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/auth"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/reqctx"
	"liquiswiss/pkg/utils"
)

type eventsTestEnv struct {
	Server     *httptest.Server
	APIService api_service.IAPIService
	Hub        *events.Hub
	UserA      *models.User
	UserB      *models.User
	CurrencyID int64
}

func setupEventsTestEnvironment(t *testing.T) *eventsTestEnv {
	t.Helper()

	conn := SetupTestEnvironment(t)
	t.Cleanup(func() { conn.Close() })

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	emailService := email_adapter.NewEmailAdapter(config.Config{})
	apiService := api_service.NewAPIService(dbAdapter, emailService)
	middleware.InjectUserService(dbAdapter)

	hub := events.NewHub()
	apiService.SetEventHub(hub)
	apiHandler := api.NewAPI(dbAdapter, apiService, emailService)
	apiHandler.EventHub = hub

	userA, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "events-user-a@events-test.com", "test", "Events Org A",
	)
	require.NoError(t, err)
	userB, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "events-user-b@events-test.com", "test", "Events Org B",
	)
	require.NoError(t, err)

	currency, err := CreateCurrency(apiService, "CHF", "Schweizer Franken", "de-CH")
	require.NoError(t, err)

	server := httptest.NewServer(apiHandler.Router)
	t.Cleanup(server.Close)

	return &eventsTestEnv{
		Server:     server,
		APIService: apiService,
		Hub:        hub,
		UserA:      userA,
		UserB:      userB,
		CurrencyID: *currency.ID,
	}
}

func (env *eventsTestEnv) openStream(t *testing.T, user *models.User, clientID ...string) (*http.Response, *bufio.Reader) {
	t.Helper()
	accessToken, expiration, _, err := auth.GenerateAccessToken(*user)
	require.NoError(t, err)
	cookie := auth.GenerateCookie(utils.AccessTokenName, accessToken, expiration)

	url := env.Server.URL + "/api/events"
	if len(clientID) > 0 {
		url += "?client=" + clientID[0]
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	req.AddCookie(&cookie)

	resp, err := env.Server.Client().Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { resp.Body.Close() })
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, resp.Header.Get("Content-Type"), "text/event-stream")
	return resp, bufio.NewReader(resp.Body)
}

// readEventLine reads stream lines until a data line arrives or the timeout hits.
// Returns the data payload, or "" on timeout/stream end.
func readEventLine(reader *bufio.Reader, timeout time.Duration) string {
	result := make(chan string, 1)
	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				result <- ""
				return
			}
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "data:") {
				result <- strings.TrimSpace(strings.TrimPrefix(line, "data:"))
				return
			}
		}
	}()
	select {
	case payload := <-result:
		return payload
	case <-time.After(timeout):
		return ""
	}
}

func TestEventsStreamReceivesOwnOrganisationEvents(t *testing.T) {
	env := setupEventsTestEnvironment(t)
	_, reader := env.openStream(t, env.UserA)

	_, err := env.APIService.CreateBankAccount(context.Background(), models.CreateBankAccount{
		Name:     "Eventkonto",
		Amount:   1000,
		Currency: env.CurrencyID,
	}, env.UserA.ID)
	require.NoError(t, err)

	payload := readEventLine(reader, 3*time.Second)
	require.Contains(t, payload, `"entity":"bank_account"`)
	require.Contains(t, payload, `"action":"created"`)
}

// The acting tab (matching user + client id) sees own=true; every other
// connection of the same user sees own=false.
func TestEventsStreamOwnFlagForActingTab(t *testing.T) {
	env := setupEventsTestEnvironment(t)
	_, actingReader := env.openStream(t, env.UserA, "tab-1")
	_, otherTabReader := env.openStream(t, env.UserA, "tab-2")
	_, noClientReader := env.openStream(t, env.UserA)

	ctx := reqctx.WithClientID(context.Background(), "tab-1")
	_, err := env.APIService.CreateBankAccount(ctx, models.CreateBankAccount{
		Name:     "Eigenes Konto",
		Amount:   100,
		Currency: env.CurrencyID,
	}, env.UserA.ID)
	require.NoError(t, err)

	require.Contains(t, readEventLine(actingReader, 3*time.Second), `"own":true`)
	require.Contains(t, readEventLine(otherTabReader, 3*time.Second), `"own":false`)
	require.Contains(t, readEventLine(noClientReader, 3*time.Second), `"own":false`)
}

// Mutations without a client id (MCP, background jobs) are own=false for
// every tab: exactly these changes must trigger notifications.
func TestEventsStreamMCPMutationsAreNeverOwn(t *testing.T) {
	env := setupEventsTestEnvironment(t)
	_, reader := env.openStream(t, env.UserA, "tab-1")

	_, err := env.APIService.CreateBankAccount(context.Background(), models.CreateBankAccount{
		Name:     "MCP Konto",
		Amount:   200,
		Currency: env.CurrencyID,
	}, env.UserA.ID)
	require.NoError(t, err)

	require.Contains(t, readEventLine(reader, 3*time.Second), `"own":false`)
}

func TestEventsStreamDoesNotLeakAcrossOrganisations(t *testing.T) {
	env := setupEventsTestEnvironment(t)
	_, reader := env.openStream(t, env.UserB)

	// Mutation happens in user A's organisation; user B must not see it
	_, err := env.APIService.CreateBankAccount(context.Background(), models.CreateBankAccount{
		Name:     "Fremdes Konto",
		Amount:   500,
		Currency: env.CurrencyID,
	}, env.UserA.ID)
	require.NoError(t, err)

	payload := readEventLine(reader, 1500*time.Millisecond)
	require.Empty(t, payload, "user B received an event from user A's organisation")
}

func TestEventsStreamConnectionCapEvictsOldest(t *testing.T) {
	env := setupEventsTestEnvironment(t)
	_, oldestReader := env.openStream(t, env.UserA)
	for range events.MaxConnectionsPerUser {
		env.openStream(t, env.UserA)
	}

	// Oldest stream must be closed by the eviction; its body read terminates
	done := make(chan struct{})
	go func() {
		for {
			if _, err := oldestReader.ReadString('\n'); err != nil {
				close(done)
				return
			}
		}
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("oldest stream not evicted at connection cap")
	}
}

func TestEventsStreamClosedAtTokenExpiry(t *testing.T) {
	env := setupEventsTestEnvironment(t)

	// Craft an access token that expires almost immediately; the stream
	// deadline must follow the token expiry, not maxStreamLifetime
	expirationTime := time.Now().Add(1500 * time.Millisecond)
	claims := &auth.Claims{
		UserID: env.UserA.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(config.GetConfig().JWTKey)
	require.NoError(t, err)
	cookie := auth.GenerateCookie(utils.AccessTokenName, tokenString, expirationTime)

	req, err := http.NewRequest(http.MethodGet, env.Server.URL+"/api/events", nil)
	require.NoError(t, err)
	req.AddCookie(&cookie)
	resp, err := env.Server.Client().Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { resp.Body.Close() })
	require.Equal(t, http.StatusOK, resp.StatusCode)
	reader := bufio.NewReader(resp.Body)

	done := make(chan struct{})
	go func() {
		for {
			if _, err := reader.ReadString('\n'); err != nil {
				close(done)
				return
			}
		}
	}()
	select {
	case <-done:
	case <-time.After(4 * time.Second):
		t.Fatal("stream not closed at access token expiry")
	}
}

func TestEventsStreamClosedOnOrganisationSwitch(t *testing.T) {
	env := setupEventsTestEnvironment(t)

	// Give user A a second organisation to switch to
	organisation, err := env.APIService.CreateOrganisation(context.Background(), models.CreateOrganisation{
		Name: "Zweite Org",
	}, env.UserA.ID)
	require.NoError(t, err)

	resp, reader := env.openStream(t, env.UserA)
	_ = resp

	err = env.APIService.SetUserCurrentOrganisation(context.Background(), models.UpdateUserCurrentOrganisation{
		OrganisationID: organisation.ID,
	}, env.UserA.ID)
	require.NoError(t, err)

	// Stream must terminate; readEventLine returns "" when the body closes
	done := make(chan struct{})
	go func() {
		for {
			if _, err := reader.ReadString('\n'); err != nil {
				close(done)
				return
			}
		}
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("stream not closed after organisation switch")
	}
}
