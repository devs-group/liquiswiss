package models

import "time"

type VatSetting struct {
	ID              int64     `db:"id" json:"id"`
	OrganisationID  int64     `db:"organisation_id" json:"organisationId"`
	Enabled         bool      `db:"enabled" json:"enabled"`
	BillingDate     time.Time `db:"billing_date" json:"billingDate"`         // Rechnungszeitpunkt
	TransactionDate time.Time `db:"transaction_date" json:"transactionDate"` // Transaktionszeitpunkt
	Interval        string    `db:"interval" json:"interval"`                // monthly, quarterly, biannually, yearly
	CreatedAt       time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt       time.Time `db:"updated_at" json:"updatedAt"`
}

type CreateVatSetting struct {
	Enabled         bool   `json:"enabled" validate:"required"`
	BillingDate     string `json:"billingDate" validate:"required,datetime=2006-01-02"`
	TransactionDate string `json:"transactionDate" validate:"required,datetime=2006-01-02"`
	Interval        string `json:"interval" validate:"required,oneof=monthly quarterly biannually yearly"`
}

type UpdateVatSetting struct {
	Enabled         *bool   `json:"enabled" validate:"omitempty"`
	BillingDate     *string `json:"billingDate" validate:"omitempty,datetime=2006-01-02"`
	TransactionDate *string `json:"transactionDate" validate:"omitempty,datetime=2006-01-02"`
	Interval        *string `json:"interval" validate:"omitempty,oneof=monthly quarterly biannually yearly"`
}
