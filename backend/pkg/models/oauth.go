package models

import "time"

type OAuthClient struct {
	ClientID     string   `json:"client_id"`
	ClientName   string   `json:"client_name"`
	RedirectURIs []string `json:"redirect_uris"`
}

type OAuthAuthCode struct {
	CodeHash      string
	ClientID      string
	UserID        int64
	CodeChallenge string
	RedirectURI   string
	Resource      string
	ExpiresAt     time.Time
	UsedAt        *time.Time
}

type OAuthRefreshToken struct {
	TokenHash   string
	ClientID    string
	UserID      int64
	ExpiresAt   time.Time
	RevokedAt   *time.Time
	RotatedFrom *string
}

type OAuthConnection struct {
	ClientID   string    `json:"clientId"`
	ClientName string    `json:"clientName"`
	CreatedAt  time.Time `json:"createdAt"`
	LastUsedAt time.Time `json:"lastUsedAt"`
}
