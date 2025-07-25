//go:generate mockgen -package=mocks -destination ../mocks/db_service.go liquiswiss/internal/service/db_service IDatabaseService
package db_service

import (
	"database/sql"
	"embed"
	"liquiswiss/pkg/models"
	"time"
)

//go:embed queries/*.sql
var sqlQueries embed.FS

var allowedSortOrders = map[string]bool{
	"ASC": true, "DESC": true,
}

type IDatabaseService interface {
	CreateRegistration(email, code string) (int64, error)
	ValidateRegistration(email, code string, hours time.Duration) (int64, error)
	DeleteRegistration(registrationID int64, email string) error

	CreateUser(email, password string) (int64, error)
	GetUserPasswordByEMail(email string) (*models.Login, error)
	GetProfile(userID int64) (*models.User, error)
	UpdateProfile(payload models.UpdateUser, userID int64) error
	UpdatePassword(password string, userID int64) error
	ResetPassword(password string, email string) error
	CheckUserExistence(id int64) (bool, error)
	SetUserCurrentOrganisation(userID int64, organisationID int64) error
	CreateResetPassword(email, code string, delay time.Duration) (bool, error)
	ValidateResetPassword(email, code string, hours time.Duration) (int64, error)
	DeleteResetPassword(email string) error

	ListTransactions(userID int64, page int64, limit int64, sortBy string, sortOrder string) ([]models.Transaction, int64, error)
	GetTransaction(userID int64, transactionID int64) (*models.Transaction, error)
	CreateTransaction(payload models.CreateTransaction, userID int64) (int64, error)
	UpdateTransaction(payload models.UpdateTransaction, userID int64, transactionID int64) error
	DeleteTransaction(userID int64, transactionID int64) error

	ListOrganisations(userID int64, page int64, limit int64) ([]models.Organisation, int64, error)
	GetOrganisation(userID int64, organisationID int64) (*models.Organisation, error)
	CreateOrganisation(name string) (int64, error)
	UpdateOrganisation(payload models.UpdateOrganisation, userID int64, organisationID int64) error
	AssignUserToOrganisation(userID int64, organisationID int64, role string, isDefault bool) error

	ListEmployees(userID int64, page int64, limit int64, sortBy string, sortOrder string) ([]models.Employee, int64, error)
	GetEmployee(userID int64, employeeID int64) (*models.Employee, error)
	CreateEmployee(payload models.CreateEmployee, userID int64) (int64, error)
	UpdateEmployee(payload models.UpdateEmployee, userID int64, employeeID int64) error
	DeleteEmployee(employeeID int64, userID int64) error
	CountEmployees(userID int64, page int64, limit int64) (int64, error)

	ListSalaries(userID int64, employeeID int64, page int64, limit int64) ([]models.Salary, int64, error)
	GetSalary(userID int64, salaryID int64) (*models.Salary, error)
	CreateSalary(payload models.CreateSalary, userID int64, employeeID int64) (int64, *int64, *int64, error)
	UpdateSalary(payload models.UpdateSalary, employeeID int64, salaryID int64) (*int64, *int64, error)
	DeleteSalary(existingSalary *models.Salary, userID int64) (*int64, *int64, error)

	ListSalaryCosts(userID int64, salaryID int64, page int64, limit int64) ([]models.SalaryCost, int64, error)
	GetSalaryCost(userID int64, salaryCostID int64) (*models.SalaryCost, error)
	CreateSalaryCost(payload models.CreateSalaryCost, userID int64, salaryID int64) (int64, error)
	CopySalaryCosts(payload models.CopySalaryCosts, userID int64, salaryID int64) error
	UpdateSalaryCost(payload models.CreateSalaryCost, userID int64, salaryCostID int64) error
	DeleteSalaryCost(salaryCostID int64, userID int64) error

	ListSalaryCostDetails(salaryCostID int64) ([]models.SalaryCostDetail, error)
	CalculateSalaryCostDetails(salaryCostID int64, userID int64) error
	UpsertSalaryCostDetails(payload models.CreateSalaryCostDetail) (int64, error)
	RefreshSalaryCostDetails(userID int64, salaryID int64) error

	ListSalaryCostLabels(userID int64, page int64, limit int64) ([]models.SalaryCostLabel, int64, error)
	GetSalaryCostLabel(userID int64, salaryCostLabelID int64) (*models.SalaryCostLabel, error)
	CreateSalaryCostLabel(payload models.CreateSalaryCostLabel, userID int64) (int64, error)
	UpdateSalaryCostLabel(payload models.CreateSalaryCostLabel, userID int64, salaryCostLabelID int64) error
	DeleteSalaryCostLabel(salaryCostLabelID int64, userID int64) error

	ListForecasts(userID int64, limit int64) ([]models.Forecast, error)
	ListForecastDetails(userID int64, limit int64) ([]models.ForecastDatabaseDetails, error)
	UpsertForecast(payload models.CreateForecast, userID int64) (int64, error)
	UpsertForecastDetail(payload models.CreateForecastDetail, userID, forecastID int64) (int64, error)
	ListForecastExclusions(userID, relatedID int64, relatedTable string) (map[string]bool, error)
	CreateForecastExclusion(payload models.CreateForecastExclusion, userID int64) (int64, error)
	DeleteForecastExclusion(payload models.CreateForecastExclusion, userID int64) (int64, error)
	ClearForecasts(userID int64) (int64, error)

	ListBankAccounts(userID int64) ([]models.BankAccount, error)
	GetBankAccount(userID int64, bankAccountID int64) (*models.BankAccount, error)
	CreateBankAccount(payload models.CreateBankAccount, userID int64) (int64, error)
	UpdateBankAccount(payload models.UpdateBankAccount, userID int64, bankAccountID int64) error
	DeleteBankAccount(userID int64, bankAccountID int64) error

	ListVats(userID int64) ([]models.Vat, error)
	GetVat(userID int64, vatID int64) (*models.Vat, error)
	CreateVat(payload models.CreateVat, userID int64) (int64, error)
	UpdateVat(payload models.UpdateVat, userID int64, vatID int64) error
	DeleteVat(userID int64, vatID int64) error

	ListCategories(userID, page, limit int64) ([]models.Category, int64, error)
	GetCategory(userID int64, categoryID int64) (*models.Category, error)
	CreateCategory(userID *int64, payload models.CreateCategory) (int64, error)
	UpdateCategory(userID int64, payload models.UpdateCategory, categoryID int64) error

	ListCurrencies(userID int64) ([]models.Currency, error)
	GetCurrency(currencyID int64) (*models.Currency, error)
	CreateCurrency(payload models.CreateCurrency) (int64, error)
	UpdateCurrency(payload models.UpdateCurrency, currencyID int64) error
	CountCurrencies() (int64, error)

	StoreRefreshTokenID(userID int64, tokenId string, expirationTime time.Time, deviceName string) error
	CheckRefreshToken(tokenID string, userID int64) (bool, error)
	DeleteRefreshToken(tokenID string, userID int64) error

	ListFiatRates(base string) ([]models.FiatRate, error)
	CountUniqueCurrenciesInFiatRates() (int64, error)
	GetFiatRate(base, target string) (*models.FiatRate, error)
	UpsertFiatRate(payload models.CreateFiatRate) error
}

type DatabaseService struct {
	db *sql.DB
}

func NewDatabaseService(db *sql.DB) IDatabaseService {
	return &DatabaseService{
		db: db,
	}
}
