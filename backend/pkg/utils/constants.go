package utils

import "time"

const (
	AccessTokenValidity = 20 * time.Minute
	//AccessTokenValidity = 1 * time.Second

	RefreshTokenValidity = 90 * 24 * time.Hour // 3 months validity
	//RefreshTokenValidity = 3 * time.Second

	AccessTokenName  = "liq-access-token"
	RefreshTokenName = "liq-refresh-token"
)
