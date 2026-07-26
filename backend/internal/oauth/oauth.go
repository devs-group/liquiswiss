package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"liquiswiss/config"
	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/pkg/auth"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

const (
	AuthCodeValidity = 10 * time.Minute
	// MCPAudience marks access tokens issued for the MCP resource so the
	// bearer middleware can reject regular web session tokens
	MCPAudience = "liquiswiss-mcp"

	authCodePrefix     = "lsw_ac_"
	refreshTokenPrefix = "lsw_rt_"
)

type Handler struct {
	dbService db_adapter.IDatabaseAdapter
}

func NewHandler(dbService db_adapter.IDatabaseAdapter) *Handler {
	return &Handler{dbService: dbService}
}

func randomToken(prefix string) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(raw), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// mcpResourceURL is the canonical resource identifier of the MCP endpoint
func mcpResourceURL() string {
	cfg := config.GetConfig()
	return strings.TrimSuffix(cfg.APIHost, "/") + "/api/mcp"
}

// ProtectedResourceMetadata handles GET /.well-known/oauth-protected-resource
func (h *Handler) ProtectedResourceMetadata(c *gin.Context) {
	cfg := config.GetConfig()
	c.JSON(http.StatusOK, gin.H{
		"resource":              mcpResourceURL(),
		"authorization_servers": []string{strings.TrimSuffix(cfg.APIHost, "/")},
	})
}

// AuthorizationServerMetadata handles GET /.well-known/oauth-authorization-server (RFC 8414)
func (h *Handler) AuthorizationServerMetadata(c *gin.Context) {
	cfg := config.GetConfig()
	base := strings.TrimSuffix(cfg.APIHost, "/")
	c.JSON(http.StatusOK, gin.H{
		"issuer":                                base,
		"authorization_endpoint":                base + "/api/oauth/authorize",
		"token_endpoint":                        base + "/api/oauth/token",
		"registration_endpoint":                 base + "/api/oauth/register",
		"revocation_endpoint":                   base + "/api/oauth/revoke",
		"response_types_supported":              []string{"code"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
		"code_challenge_methods_supported":      []string{"S256"},
		"token_endpoint_auth_methods_supported": []string{"none"},
	})
}

type registerRequest struct {
	ClientName   string   `json:"client_name"`
	RedirectURIs []string `json:"redirect_uris"`
}

// Register handles POST /api/oauth/register (RFC 7591 dynamic client registration, public clients only)
func (h *Handler) Register(c *gin.Context) {
	var payload registerRequest
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_client_metadata"})
		return
	}
	if len(payload.RedirectURIs) == 0 || len(payload.RedirectURIs) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_redirect_uri"})
		return
	}
	for _, uri := range payload.RedirectURIs {
		parsed, err := url.Parse(uri)
		if err != nil || !parsed.IsAbs() || parsed.Fragment != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_redirect_uri"})
			return
		}
		// Only loopback redirects may use plain http (native clients per RFC 8252)
		if parsed.Scheme == "http" {
			host := parsed.Hostname()
			if host != "localhost" && host != "127.0.0.1" && host != "::1" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_redirect_uri"})
				return
			}
		} else if parsed.Scheme != "https" && !strings.Contains(parsed.Scheme, ".") {
			// custom reverse-domain schemes (e.g. com.example.app:/callback) are allowed for native apps
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_redirect_uri"})
			return
		}
	}

	clientName := strings.TrimSpace(payload.ClientName)
	if clientName == "" {
		clientName = "Unbenannte Anwendung"
	}
	if len(clientName) > 255 {
		clientName = clientName[:255]
	}

	clientID := utils.GenerateUUID()
	if err := h.dbService.CreateOAuthClient(clientID, clientName, payload.RedirectURIs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"client_id":                  clientID,
		"client_name":                clientName,
		"redirect_uris":              payload.RedirectURIs,
		"token_endpoint_auth_method": "none",
		"grant_types":                []string{"authorization_code", "refresh_token"},
		"response_types":             []string{"code"},
	})
}

// Authorize handles GET /api/oauth/authorize: validates the request and forwards
// the user to the frontend consent page, which completes the flow via Approve
func (h *Handler) Authorize(c *gin.Context) {
	cfg := config.GetConfig()

	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	client, err := h.dbService.GetOAuthClient(clientID)
	if err != nil {
		c.String(http.StatusInternalServerError, "server error")
		return
	}
	// Invalid client or redirect_uri: never redirect, respond directly (RFC 6749 §4.1.2.1)
	if client == nil || !contains(client.RedirectURIs, redirectURI) {
		c.String(http.StatusBadRequest, "invalid client_id or redirect_uri")
		return
	}

	redirectErr := func(code string) {
		target, _ := url.Parse(redirectURI)
		q := target.Query()
		q.Set("error", code)
		if state := c.Query("state"); state != "" {
			q.Set("state", state)
		}
		target.RawQuery = q.Encode()
		c.Redirect(http.StatusFound, target.String())
	}

	if c.Query("response_type") != "code" {
		redirectErr("unsupported_response_type")
		return
	}
	if c.Query("code_challenge_method") != "S256" || c.Query("code_challenge") == "" {
		redirectErr("invalid_request")
		return
	}

	// Hand over to the frontend consent page with the validated query intact
	consentURL := strings.TrimSuffix(cfg.WebHost, "/") + "/oauth/consent?" + c.Request.URL.RawQuery +
		"&client_name=" + url.QueryEscape(client.ClientName)
	c.Redirect(http.StatusFound, consentURL)
}

type approveRequest struct {
	ClientID      string `json:"clientId" binding:"required"`
	RedirectURI   string `json:"redirectUri" binding:"required"`
	CodeChallenge string `json:"codeChallenge" binding:"required"`
	State         string `json:"state"`
	Resource      string `json:"resource"`
	Approve       bool   `json:"approve"`
}

// Approve handles POST /api/oauth/approve (cookie-authenticated): issues the
// authorization code after the user consented and returns the redirect target
func (h *Handler) Approve(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var payload approveRequest
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	client, err := h.dbService.GetOAuthClient(payload.ClientID)
	if err != nil || client == nil || !contains(client.RedirectURIs, payload.RedirectURI) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client_id or redirect_uri"})
		return
	}

	target, err := url.Parse(payload.RedirectURI)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid redirect_uri"})
		return
	}
	q := target.Query()
	if payload.State != "" {
		q.Set("state", payload.State)
	}

	if !payload.Approve {
		q.Set("error", "access_denied")
		target.RawQuery = q.Encode()
		c.JSON(http.StatusOK, gin.H{"redirect": target.String()})
		return
	}

	code, err := randomToken(authCodePrefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}
	err = h.dbService.CreateOAuthAuthCode(models.OAuthAuthCode{
		CodeHash:      hashToken(code),
		ClientID:      payload.ClientID,
		UserID:        userID,
		CodeChallenge: payload.CodeChallenge,
		RedirectURI:   payload.RedirectURI,
		Resource:      payload.Resource,
		ExpiresAt:     time.Now().Add(AuthCodeValidity),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}

	q.Set("code", code)
	target.RawQuery = q.Encode()
	c.JSON(http.StatusOK, gin.H{"redirect": target.String()})
}

// Token handles POST /api/oauth/token (form-encoded, public client + PKCE)
func (h *Handler) Token(c *gin.Context) {
	switch c.PostForm("grant_type") {
	case "authorization_code":
		h.tokenAuthorizationCode(c)
	case "refresh_token":
		h.tokenRefresh(c)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_grant_type"})
	}
}

func (h *Handler) tokenAuthorizationCode(c *gin.Context) {
	code := c.PostForm("code")
	verifier := c.PostForm("code_verifier")
	clientID := c.PostForm("client_id")
	redirectURI := c.PostForm("redirect_uri")

	if code == "" || verifier == "" || clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	authCode, err := h.dbService.GetOAuthAuthCode(hashToken(code))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}
	if authCode == nil || authCode.ClientID != clientID || time.Now().After(authCode.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
		return
	}
	if redirectURI != authCode.RedirectURI {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
		return
	}

	// PKCE S256: challenge must equal BASE64URL(SHA256(verifier))
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	if subtle.ConstantTimeCompare([]byte(challenge), []byte(authCode.CodeChallenge)) != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
		return
	}

	// Single use: the first consumer wins, replays get rejected
	consumed, err := h.dbService.MarkOAuthAuthCodeUsed(authCode.CodeHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}
	if !consumed {
		// Code replay: revoke everything issued to this client/user pair
		if err := h.dbService.RevokeOAuthConnection(authCode.UserID, authCode.ClientID); err != nil {
			logger.Logger.Error("Failed to revoke OAuth connection after code replay", err)
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
		return
	}

	h.issueTokenPair(c, authCode.UserID, authCode.ClientID, nil)
}

func (h *Handler) tokenRefresh(c *gin.Context) {
	refreshToken := c.PostForm("refresh_token")
	clientID := c.PostForm("client_id")
	if refreshToken == "" || clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	tokenHash := hashToken(refreshToken)
	stored, err := h.dbService.GetOAuthRefreshToken(tokenHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}
	if stored == nil || stored.ClientID != clientID || time.Now().After(stored.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
		return
	}
	if stored.RevokedAt != nil {
		// Reuse of a rotated/revoked token: assume theft, kill the whole connection
		if err := h.dbService.RevokeOAuthConnection(stored.UserID, stored.ClientID); err != nil {
			logger.Logger.Error("Failed to revoke OAuth connection after refresh token reuse", err)
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
		return
	}

	if err := h.dbService.RevokeOAuthRefreshToken(tokenHash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	h.issueTokenPair(c, stored.UserID, stored.ClientID, &tokenHash)
}

func (h *Handler) issueTokenPair(c *gin.Context, userID int64, clientID string, rotatedFrom *string) {
	accessToken, expiresAt, err := GenerateAccessToken(userID, clientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	refreshToken, err := randomToken(refreshTokenPrefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}
	err = h.dbService.CreateOAuthRefreshToken(models.OAuthRefreshToken{
		TokenHash:   hashToken(refreshToken),
		ClientID:    clientID,
		UserID:      userID,
		ExpiresAt:   time.Now().Add(utils.RefreshTokenValidity),
		RotatedFrom: rotatedFrom,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    int(time.Until(expiresAt).Seconds()),
		"refresh_token": refreshToken,
	})
}

// Revoke handles POST /api/oauth/revoke (RFC 7009). Always returns 200.
func (h *Handler) Revoke(c *gin.Context) {
	token := c.PostForm("token")
	if strings.HasPrefix(token, refreshTokenPrefix) {
		_ = h.dbService.RevokeOAuthRefreshToken(hashToken(token))
	}
	c.Status(http.StatusOK)
}

// ListConnections handles GET /api/oauth/connections (cookie-authenticated)
func (h *Handler) ListConnections(c *gin.Context) {
	userID := c.GetInt64("userID")
	connections, err := h.dbService.ListOAuthConnections(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}
	c.JSON(http.StatusOK, connections)
}

// RevokeConnection handles DELETE /api/oauth/connections/:clientId (cookie-authenticated)
func (h *Handler) RevokeConnection(c *gin.Context) {
	userID := c.GetInt64("userID")
	if err := h.dbService.RevokeOAuthConnection(userID, c.Param("clientId")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}
	c.Status(http.StatusNoContent)
}

// GenerateAccessToken creates a short-lived JWT bound to the MCP audience
func GenerateAccessToken(userID int64, clientID string) (string, time.Time, error) {
	token, expiresAt, err := auth.GenerateAudienceAccessToken(userID, clientID, MCPAudience)
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
