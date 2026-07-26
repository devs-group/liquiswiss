package handlers_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"liquiswiss/config"
	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/adapter/email_adapter"
	"liquiswiss/internal/api"
	"liquiswiss/internal/middleware"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/auth"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

type oauthTestEnv struct {
	API       *api.API
	User      *models.User
	DBAdapter db_adapter.IDatabaseAdapter
}

func setupOAuthTestEnvironment(t *testing.T) *oauthTestEnv {
	t.Helper()

	conn := SetupTestEnvironment(t)
	t.Cleanup(func() { conn.Close() })

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	emailService := email_adapter.NewEmailAdapter(config.Config{})
	apiService := api_service.NewAPIService(dbAdapter, emailService)
	middleware.InjectUserService(dbAdapter)

	user, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "oauth-user@oauth-test.com", "test", "OAuth Test Org",
	)
	require.NoError(t, err)

	return &oauthTestEnv{
		API:       api.NewAPI(dbAdapter, apiService, emailService),
		User:      user,
		DBAdapter: dbAdapter,
	}
}

func (env *oauthTestEnv) registerClient(t *testing.T, redirectURI string) string {
	t.Helper()
	body := fmt.Sprintf(`{"client_name":"Test MCP Client","redirect_uris":[%q]}`, redirectURI)
	req, _ := http.NewRequest(http.MethodPost, "/api/oauth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.API.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	var response map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	clientID, _ := response["client_id"].(string)
	require.NotEmpty(t, clientID)
	return clientID
}

func (env *oauthTestEnv) sessionCookie(t *testing.T) *http.Cookie {
	t.Helper()
	accessToken, expiresAt, _, err := auth.GenerateAccessToken(models.User{ID: env.User.ID})
	require.NoError(t, err)
	cookie := auth.GenerateCookie(utils.AccessTokenName, accessToken, expiresAt)
	return &cookie
}

// approve runs the consent approval as the logged-in user and returns the authorization code
func (env *oauthTestEnv) approve(t *testing.T, clientID, redirectURI, challenge string) string {
	t.Helper()
	body := fmt.Sprintf(`{"clientId":%q,"redirectUri":%q,"codeChallenge":%q,"state":"xyz","approve":true}`, clientID, redirectURI, challenge)
	req, _ := http.NewRequest(http.MethodPost, "/api/oauth/approve", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(env.sessionCookie(t))
	w := httptest.NewRecorder()
	env.API.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var response map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	redirect, err := url.Parse(response["redirect"])
	require.NoError(t, err)
	code := redirect.Query().Get("code")
	require.NotEmpty(t, code)
	require.Equal(t, "xyz", redirect.Query().Get("state"))
	return code
}

func (env *oauthTestEnv) tokenRequest(t *testing.T, form url.Values) (int, map[string]any) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, "/api/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	env.API.Router.ServeHTTP(w, req)

	var response map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	return w.Code, response
}

func pkcePair() (string, string) {
	verifier := "test-verifier-with-enough-entropy-0123456789"
	sum := sha256.Sum256([]byte(verifier))
	return verifier, base64.RawURLEncoding.EncodeToString(sum[:])
}

func TestOAuthDiscoveryEndpoints(t *testing.T) {
	env := setupOAuthTestEnvironment(t)

	for _, path := range []string{"/.well-known/oauth-protected-resource", "/.well-known/oauth-authorization-server"} {
		req, _ := http.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		env.API.Router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code, path)
	}
}

func TestOAuthAuthorizeRedirectsToConsent(t *testing.T) {
	env := setupOAuthTestEnvironment(t)
	redirectURI := "http://localhost:33418/callback"
	clientID := env.registerClient(t, redirectURI)
	_, challenge := pkcePair()

	query := url.Values{
		"response_type":         {"code"},
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
		"state":                 {"xyz"},
	}
	req, _ := http.NewRequest(http.MethodGet, "/api/oauth/authorize?"+query.Encode(), nil)
	w := httptest.NewRecorder()
	env.API.Router.ServeHTTP(w, req)

	require.Equal(t, http.StatusFound, w.Code)
	location := w.Header().Get("Location")
	require.Contains(t, location, "/oauth/consent")
	require.Contains(t, location, "client_name=Test+MCP+Client")

	// Unknown client must not redirect anywhere
	query.Set("client_id", "unknown")
	req, _ = http.NewRequest(http.MethodGet, "/api/oauth/authorize?"+query.Encode(), nil)
	w = httptest.NewRecorder()
	env.API.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOAuthFullFlowAndMCPAccess(t *testing.T) {
	env := setupOAuthTestEnvironment(t)
	redirectURI := "http://localhost:33418/callback"
	clientID := env.registerClient(t, redirectURI)
	verifier, challenge := pkcePair()

	code := env.approve(t, clientID, redirectURI, challenge)

	// Exchange code for tokens
	status, tokens := env.tokenRequest(t, url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"code_verifier": {verifier},
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
	})
	require.Equal(t, http.StatusOK, status)
	accessToken, _ := tokens["access_token"].(string)
	refreshToken, _ := tokens["refresh_token"].(string)
	require.NotEmpty(t, accessToken)
	require.NotEmpty(t, refreshToken)

	// MCP endpoint without token responds 401 with resource metadata pointer
	initialize := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}`
	req, _ := http.NewRequest(http.MethodPost, "/api/mcp", strings.NewReader(initialize))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.API.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Contains(t, w.Header().Get("WWW-Authenticate"), "oauth-protected-resource")

	// MCP endpoint with the OAuth access token works
	req, _ = http.NewRequest(http.MethodPost, "/api/mcp", strings.NewReader(initialize))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w = httptest.NewRecorder()
	env.API.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), "liquiswiss")

	// The MCP access token must NOT work as a web session cookie
	req, _ = http.NewRequest(http.MethodGet, "/api/profile", nil)
	req.AddCookie(&http.Cookie{Name: utils.AccessTokenName, Value: accessToken})
	w = httptest.NewRecorder()
	env.API.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	// Refresh rotation: old refresh token gets a new pair
	status, rotated := env.tokenRequest(t, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {clientID},
	})
	require.Equal(t, http.StatusOK, status)
	newRefreshToken, _ := rotated["refresh_token"].(string)
	require.NotEmpty(t, newRefreshToken)
	require.NotEqual(t, refreshToken, newRefreshToken)

	// Reusing the rotated (now revoked) refresh token must fail AND kill the whole chain
	status, reuse := env.tokenRequest(t, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {clientID},
	})
	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, "invalid_grant", reuse["error"])

	status, _ = env.tokenRequest(t, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {newRefreshToken},
		"client_id":     {clientID},
	})
	require.Equal(t, http.StatusBadRequest, status, "reuse detection must revoke the successor token too")
}

func TestOAuthPKCEMismatchRejected(t *testing.T) {
	env := setupOAuthTestEnvironment(t)
	redirectURI := "http://localhost:33418/callback"
	clientID := env.registerClient(t, redirectURI)
	_, challenge := pkcePair()

	code := env.approve(t, clientID, redirectURI, challenge)

	status, response := env.tokenRequest(t, url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"code_verifier": {"totally-wrong-verifier-aaaaaaaaaaaaaaaaaaaa"},
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
	})
	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, "invalid_grant", response["error"])
}

func TestOAuthConnectionsListAndRevoke(t *testing.T) {
	env := setupOAuthTestEnvironment(t)
	redirectURI := "http://localhost:33418/callback"
	clientID := env.registerClient(t, redirectURI)
	verifier, challenge := pkcePair()
	code := env.approve(t, clientID, redirectURI, challenge)

	status, tokens := env.tokenRequest(t, url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"code_verifier": {verifier},
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
	})
	require.Equal(t, http.StatusOK, status)
	refreshToken, _ := tokens["refresh_token"].(string)

	// Connection appears in the list
	req, _ := http.NewRequest(http.MethodGet, "/api/oauth/connections", nil)
	req.AddCookie(env.sessionCookie(t))
	w := httptest.NewRecorder()
	env.API.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "Test MCP Client")

	// Revoke it
	req, _ = http.NewRequest(http.MethodDelete, "/api/oauth/connections/"+clientID, nil)
	req.AddCookie(env.sessionCookie(t))
	w = httptest.NewRecorder()
	env.API.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusNoContent, w.Code)

	// Refresh token no longer works
	status, _ = env.tokenRequest(t, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {clientID},
	})
	require.Equal(t, http.StatusBadRequest, status)
}

func TestOAuthRevokedConnectionBlocksMCPImmediately(t *testing.T) {
	env := setupOAuthTestEnvironment(t)
	token := env.mcpAccessToken(t)

	// Works while connected
	result := env.mcpCall(t, token, "get_organisation", `{}`)
	require.False(t, result.IsError, result.Text)

	// Revoke the connection (as the web UI does)
	connections, err := env.DBAdapter.ListOAuthConnections(env.User.ID)
	require.NoError(t, err)
	require.Len(t, connections, 1)
	require.NoError(t, env.DBAdapter.RevokeOAuthConnection(env.User.ID, connections[0].ClientID))

	// Same still-valid access token must now be rejected
	req, _ := http.NewRequest(http.MethodPost, "/api/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	env.API.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}
