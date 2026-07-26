package auth

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"liquiswiss/config"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64 `json:"userId"`
	jwt.RegisteredClaims
}

// GenerateAccessToken generates a new JWT token
func GenerateAccessToken(user models.User) (string, time.Time, *Claims, error) {
	expirationTime := time.Now().Add(utils.AccessTokenValidity)
	claims := &Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	cfg := config.GetConfig()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(cfg.JWTKey)

	return tokenString, expirationTime, claims, err
}

func GenerateRefreshToken(user models.User) (string, string, time.Time, error) {
	expirationTime := time.Now().Add(utils.RefreshTokenValidity) // Refresh token valid for 3 months
	tokenID := utils.GenerateUUID()                              // Generate a unique ID for this token

	claims := &Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			ID:        tokenID,
		},
	}

	cfg := config.GetConfig()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(cfg.JWTKey)

	return tokenString, tokenID, expirationTime, err
}

// GenerateAudienceAccessToken generates a JWT access token bound to a specific
// audience (e.g. the MCP resource) and OAuth client
func GenerateAudienceAccessToken(userID int64, clientID string, audience string) (string, time.Time, error) {
	expirationTime := time.Now().Add(utils.AccessTokenValidity)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Audience:  jwt.ClaimStrings{audience},
			Subject:   clientID,
		},
	}

	cfg := config.GetConfig()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(cfg.JWTKey)

	return tokenString, expirationTime, err
}

// VerifyToken verifies the given token and returns the user ID
func VerifyToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	cfg := config.GetConfig()
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return cfg.JWTKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// All browser traffic reaches the API through the Nuxt proxy (same origin), so
// auth cookies are strictly SameSite=Lax. Cross-site requests must never carry
// them (blocks CSRF and cross-site SSE/WebSocket hijacking).
func GenerateCookie(name, token string, expiration time.Time) http.Cookie {
	return http.Cookie{
		Name:     name,
		Value:    token,
		Expires:  expiration,
		HttpOnly: true,
		Path:     "/",
		Secure:   utils.IsProduction(),
		SameSite: http.SameSiteLaxMode,
	}
}

func GenerateDeleteCookie(name string) http.Cookie {
	return http.Cookie{
		Name:     name,
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
		Secure:   utils.IsProduction(),
		SameSite: http.SameSiteLaxMode,
	}
}

func ClearAuthCookies(c *gin.Context) {
	accessCookie := GenerateDeleteCookie(utils.AccessTokenName)
	refreshCookie := GenerateDeleteCookie(utils.RefreshTokenName)
	http.SetCookie(c.Writer, &accessCookie)
	http.SetCookie(c.Writer, &refreshCookie)
}
