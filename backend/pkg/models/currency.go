package models

type Currency struct {
	ID          int64  `db:"id" json:"id"`
	Code        string `db:"code" json:"code"`
	Description string `db:"description" json:"description"`
	LocaleCode  string `db:"locale_code" json:"localeCode"`
}

type CreateCurrency struct {
	Code        string `json:"code" validate:"required,len=3"`
	Description string `json:"description" validate:"max=30"`
	LocaleCode  string `json:"localeCode" validate:"required,len=2"`
}

type UpdateCurrency struct {
	Code        *string `json:"code" validate:"omitempty,len=3"`
	Description *string `json:"description" validate:"omitempty,max=30"`
	LocaleCode  *string `json:"localeCode" validate:"omitempty,len=2"`
}
