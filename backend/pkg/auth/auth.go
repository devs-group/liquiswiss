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

func GenerateCookie(name, token string, expiration time.Time) http.Cookie {
	sameSite := http.SameSiteLaxMode
	if utils.IsProduction() {
		sameSite = http.SameSiteNoneMode
	}
	return http.Cookie{
		Name:     name,
		Value:    token,
		Expires:  expiration,
		HttpOnly: true,
		Path:     "/",
		Secure:   utils.IsProduction(),
		SameSite: sameSite,
	}
}

func GenerateDeleteCookie(name string) http.Cookie {
	sameSite := http.SameSiteLaxMode
	if utils.IsProduction() {
		sameSite = http.SameSiteNoneMode
	}
	return http.Cookie{
		Name:     name,
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
		Secure:   utils.IsProduction(),
		SameSite: sameSite,
	}
}

func ClearAuthCookies(c *gin.Context) {
	accessCookie := GenerateDeleteCookie(utils.AccessTokenName)
	refreshCookie := GenerateDeleteCookie(utils.RefreshTokenName)
	http.SetCookie(c.Writer, &accessCookie)
	http.SetCookie(c.Writer, &refreshCookie)
}
