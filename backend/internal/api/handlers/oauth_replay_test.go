package handlers_test

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"liquiswiss/pkg/models"
)

func hashTokenForTest(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func TestOAuthRevokeConnectionAdapter(t *testing.T) {
	env := setupOAuthTestEnvironment(t)

	require.NoError(t, env.DBAdapter.CreateOAuthClient("debug-client", "Debug", []string{"http://localhost/cb"}))
	require.NoError(t, env.DBAdapter.CreateOAuthRefreshToken(models.OAuthRefreshToken{
		TokenHash: "debughash",
		ClientID:  "debug-client",
		UserID:    env.User.ID,
		ExpiresAt: time.Now().Add(time.Hour),
	}))

	require.NoError(t, env.DBAdapter.RevokeOAuthConnection(env.User.ID, "debug-client"))

	stored, err := env.DBAdapter.GetOAuthRefreshToken("debughash")
	require.NoError(t, err)
	require.NotNil(t, stored)
	require.NotNil(t, stored.RevokedAt, "refresh token should be revoked")
}

// TestOAuthCodeReplayRevokesTokensHTTP verifies that replaying an already-used
// authorization code both fails and revokes the tokens issued from it
func TestOAuthCodeReplayRevokesTokensHTTP(t *testing.T) {
	env := setupOAuthTestEnvironment(t)
	redirectURI := "http://localhost:33418/callback"
	clientID := env.registerClient(t, redirectURI)
	verifier, challenge := pkcePair()
	code := env.approve(t, clientID, redirectURI, challenge)

	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"code_verifier": {verifier},
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
	}
	status, tokens := env.tokenRequest(t, form)
	require.Equal(t, 200, status)
	refreshToken, _ := tokens["refresh_token"].(string)

	authCode, err := env.DBAdapter.GetOAuthAuthCode(hashTokenForTest(code))
	require.NoError(t, err)
	require.NotNil(t, authCode)
	require.NotNil(t, authCode.UsedAt, "auth code must be marked used after first exchange")

	status, replayResponse := env.tokenRequest(t, form)
	require.Equal(t, 400, status)
	require.Equal(t, "invalid_grant", replayResponse["error"])

	stored, err := env.DBAdapter.GetOAuthRefreshToken(hashTokenForTest(refreshToken))
	require.NoError(t, err)
	require.NotNil(t, stored)
	require.NotNil(t, stored.RevokedAt, "code replay must revoke issued refresh tokens")
}
