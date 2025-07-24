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

	ListEmployeeHistory(userID int64, employeeID int64, page int64, limit int64) ([]models.EmployeeHistory, int64, error)
	GetEmployeeHistory(userID int64, historyID int64) (*models.EmployeeHistory, error)
	CreateEmployeeHistory(payload models.CreateEmployeeHistory, userID int64, employeeID int64) (int64, *int64, *int64, error)
	UpdateEmployeeHistory(payload models.UpdateEmployeeHistory, employeeID int64, historyID int64) (*int64, *int64, error)
	DeleteEmployeeHistory(existingEmployeeHistory *models.EmployeeHistory, userID int64) error

	ListEmployeeHistoryCosts(userID int64, historyID int64, page int64, limit int64) ([]models.EmployeeHistoryCost, int64, error)
	GetEmployeeHistoryCost(userID int64, historyCostID int64) (*models.EmployeeHistoryCost, error)
	CreateEmployeeHistoryCost(payload models.CreateEmployeeHistoryCost, userID int64, historyID int64) (int64, error)
	CopyEmployeeHistoryCosts(payload models.CopyEmployeeHistoryCosts, userID int64, historyID int64) error
	UpdateEmployeeHistoryCost(payload models.CreateEmployeeHistoryCost, userID int64, historyCostID int64) error
	DeleteEmployeeHistoryCost(historyCostID int64, userID int64) error

	ListEmployeeHistoryCostDetails(historyCostID int64) ([]models.EmployeeHistoryCostDetail, error)
	CalculateEmployeeHistoryCostDetails(historyCostID int64, userID int64) error
	UpsertEmployeeHistoryCostDetails(payload models.CreateEmployeeHistoryCostDetail) (int64, error)
	RefreshCostDetails(userID int64, historyID int64) error

	ListEmployeeHistoryCostLabels(userID int64, page int64, limit int64) ([]models.EmployeeHistoryCostLabel, int64, error)
	GetEmployeeHistoryCostLabel(userID int64, historyCostLabelID int64) (*models.EmployeeHistoryCostLabel, error)
	CreateEmployeeHistoryCostLabel(payload models.CreateEmployeeHistoryCostLabel, userID int64) (int64, error)
	UpdateEmployeeHistoryCostLabel(payload models.CreateEmployeeHistoryCostLabel, userID int64, historyCostLabelID int64) error
	DeleteEmployeeHistoryCostLabel(historyCostLabelID int64, userID int64) error

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
