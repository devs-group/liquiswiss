package utils

import "time"

const (
	AccessTokenValidity = 20 * time.Minute
	//AccessTokenValidity = 1 * time.Second

	RefreshTokenValidity = 90 * 24 * time.Hour // 3 months validity
	//RefreshTokenValidity = 3 * time.Second

	RegistrationCodeValidity = 1 * time.Hour

	ResetPasswordDelay = 1 * time.Hour

	InvitationValidity = 7 * 24 * time.Hour // 7 days validity

	MaxForecastYears = 3

	AccessTokenName  = "liq-access-token"
	RefreshTokenName = "liq-refresh-token"

	TransactionsTableName = "transactions"
	SalariesTableName     = "salaries"
	SalaryCostsTableName  = "salary_costs"
)
