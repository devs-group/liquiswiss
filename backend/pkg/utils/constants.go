package utils

import "time"

const (
	AccessTokenValidity = 20 * time.Minute
	//AccessTokenValidity = 1 * time.Second

	RefreshTokenValidity = 90 * 24 * time.Hour // 3 months validity
	//RefreshTokenValidity = 3 * time.Second

	RegistrationCodeValidity = 1 * time.Hour

	ResetPasswordDelay = 1 * time.Hour

	AccessTokenName  = "liq-access-token"
	RefreshTokenName = "liq-refresh-token"

	BaseCurrency = "CHF"

	RegistrationMailTemplate = "d-7b6d32452f134c6583bc09260ab26275"
)
