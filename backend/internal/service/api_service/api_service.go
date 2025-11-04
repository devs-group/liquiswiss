//go:generate mockgen -package=mocks -destination ../../mocks/api_service.go liquiswiss/internal/service/api_service IAPIService
package api_service

import (
	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/adapter/sendgrid_adapter"
	"liquiswiss/pkg/models"
	"time"
)

type IAPIService interface {
	Login(payload models.Login, deviceName string, existingRefreshToken string) (*models.User, *string, *time.Time, *string, *time.Time, error)
	Logout(existingRefreshToken string)
	ForgotPassword(payload models.ForgotPassword, code string) error
	ResetPassword(payload models.ResetPassword) error
	CheckResetPasswordCode(payload models.CheckResetPasswordCode) error
	CreateRegistration(payload models.CreateRegistration, code string) (int64, error)
	CheckRegistrationCode(payload models.CheckRegistrationCode, validity time.Duration) (int64, error)
	FinishRegistration(payload models.FinishRegistration, deviceName string, validity time.Duration) (*models.User, *string, *time.Time, *string, *time.Time, error)
	DeleteRegistration(registrationID int64, email string) error

	GetProfile(userID int64) (*models.User, error)
	UpdateProfile(payload models.UpdateUser, userID int64) (*models.User, error)
	UpdatePassword(payload models.UpdateUserPassword, userID int64) error
	SetUserCurrentOrganisation(payload models.UpdateUserCurrentOrganisation, userID int64) error
	SetUserCurrentScenario(payload models.UpdateUserCurrentScenario, userID int64) error
	GetCurrentOrganisation(userID int64) (*models.Organisation, error)

	ListTransactions(userID int64, page int64, limit int64, sortBy string, sortOrder string) ([]models.Transaction, int64, error)
	GetTransaction(userID int64, transactionID int64) (*models.Transaction, error)
	CreateTransaction(payload models.CreateTransaction, userID int64) (*models.Transaction, error)
	UpdateTransaction(payload models.UpdateTransaction, userID int64, transactionID int64) (*models.Transaction, error)
	DeleteTransaction(userID int64, transactionID int64) error

	ListOrganisations(userID int64, page int64, limit int64) ([]models.Organisation, int64, error)
	GetOrganisation(userID int64, organisationID int64) (*models.Organisation, error)
	CreateOrganisation(payload models.CreateOrganisation, userID int64) (*models.Organisation, error)
	UpdateOrganisation(payload models.UpdateOrganisation, userID int64, organisationID int64) (*models.Organisation, error)

	ListEmployees(userID int64, page int64, limit int64, sortBy string, sortOrder string) ([]models.Employee, int64, error)
	GetEmployee(userID int64, employeeID int64) (*models.Employee, error)
	CreateEmployee(payload models.CreateEmployee, userID int64) (*models.Employee, error)
	UpdateEmployee(payload models.UpdateEmployee, userID int64, employeeID int64) (*models.Employee, error)
	DeleteEmployee(userID int64, employeeID int64) error
	CountEmployees(userID int64, page int64, limit int64) (int64, error)

	ListSalaries(userID int64, employeeID int64, page int64, limit int64) ([]models.Salary, int64, error)
	GetSalary(userID int64, salaryID int64) (*models.Salary, error)
	CreateSalary(payload models.CreateSalary, userID int64, employeeID int64) (*models.Salary, error)
	UpdateSalary(payload models.UpdateSalary, userID int64, salaryID int64) (*models.Salary, error)
	DeleteSalary(userID int64, salaryID int64) error

	ListSalaryCosts(userID int64, salaryID int64, page int64, limit int64, skipPrevious bool) ([]models.SalaryCost, int64, error)
	GetSalaryCost(userID int64, salaryCostID int64, skipPrevious bool) (*models.SalaryCost, error)
	CreateSalaryCost(payload models.CreateSalaryCost, userID int64, salaryID int64) (*models.SalaryCost, error)
	UpdateSalaryCost(payload models.CreateSalaryCost, userID int64, salaryCostID int64) (*models.SalaryCost, error)
	DeleteSalaryCost(userID int64, salaryCostID int64) error
	CopySalaryCosts(payload models.CopySalaryCosts, userID int64, salaryID int64) error

	//ListSalaryCostDetails(salaryCostID int64) ([]models.SalaryCostDetail, error)
	//CalculateSalaryCostDetails(salaryCostID int64, userID int64) error
	//UpsertSalaryCostDetails(payload models.CreateSalaryCostDetail) (int64, error)
	//RefreshSalaryCostDetails(userID int64, salaryID int64) error

	ListSalaryCostLabels(userID int64, page int64, limit int64) ([]models.SalaryCostLabel, int64, error)
	GetSalaryCostLabel(userID int64, salaryCostLabelID int64) (*models.SalaryCostLabel, error)
	CreateSalaryCostLabel(payload models.CreateSalaryCostLabel, userID int64) (*models.SalaryCostLabel, error)
	UpdateSalaryCostLabel(payload models.CreateSalaryCostLabel, userID int64, salaryCostLabelID int64) (*models.SalaryCostLabel, error)
	DeleteSalaryCostLabel(userID int64, salaryCostLabelID int64) error

	ListForecasts(userID int64, limit int64) ([]models.Forecast, error)
	ListForecastDetails(userID int64, limit int64) ([]models.ForecastDatabaseDetails, error)
	ListForecastExclusions(userID int64, relatedID int64, relatedTable string) (map[string]bool, error)
	CreateForecastExclusion(payload models.CreateForecastExclusion, userID int64) (int64, error)
	DeleteForecastExclusion(payload models.CreateForecastExclusion, userID int64) (int64, error)
	UpdateForecastExclusions(payload models.UpdateForecastExclusions, userID int64) error
	CalculateForecast(userID int64) ([]models.Forecast, error)

	ListBankAccounts(userID int64) ([]models.BankAccount, error)
	GetBankAccount(userID int64, bankAccountID int64) (*models.BankAccount, error)
	CreateBankAccount(payload models.CreateBankAccount, userID int64) (*models.BankAccount, error)
	UpdateBankAccount(payload models.UpdateBankAccount, userID int64, bankAccountID int64) (*models.BankAccount, error)
	DeleteBankAccount(userID int64, bankAccountID int64) error

	ListVats(userID int64) ([]models.Vat, error)
	GetVat(userID int64, vatID int64) (*models.Vat, error)
	CreateVat(payload models.CreateVat, userID int64) (*models.Vat, error)
	UpdateVat(payload models.UpdateVat, userID int64, vatID int64) (*models.Vat, error)
	DeleteVat(userID int64, vatID int64) error

	ListCategories(userID, page, limit int64) ([]models.Category, int64, error)
	GetCategory(userID int64, categoryID int64) (*models.Category, error)
	CreateCategory(payload models.CreateCategory, userID *int64) (*models.Category, error)
	UpdateCategory(payload models.UpdateCategory, userID int64, categoryID int64) (*models.Category, error)

	ListCurrencies(userID int64) ([]models.Currency, error)
	GetCurrency(currencyID int64) (*models.Currency, error)
	CreateCurrency(payload models.CreateCurrency) (*models.Currency, error)
	UpdateCurrency(payload models.UpdateCurrency, currencyID int64) (*models.Currency, error)
	CountCurrencies() (int64, error)

	ListFiatRates(base string) ([]models.FiatRate, error)
	GetFiatRate(base, target string) (*models.FiatRate, error)
	UpsertFiatRate(payload models.CreateFiatRate) error
	CountUniqueCurrenciesInFiatRates() (int64, error)

	ListScenarios(userID int64) ([]models.ScenarioListItem, error)
	GetScenario(userID int64, scenarioID int64) (*models.Scenario, error)
	GetDefaultScenario(userID int64) (*models.Scenario, error)
	CreateScenario(payload models.CreateScenario, userID int64) (*models.Scenario, error)
	UpdateScenario(payload models.UpdateScenario, userID int64, scenarioID int64) (*models.Scenario, error)
	DeleteScenario(userID int64, scenarioID int64) error
}

type APIService struct {
	dbService       db_adapter.IDatabaseAdapter
	sendgridAdapter sendgrid_adapter.ISendgridAdapter
}

func NewAPIService(dbService db_adapter.IDatabaseAdapter, sendgridAdapter sendgrid_adapter.ISendgridAdapter) IAPIService {
	return &APIService{
		dbService:       dbService,
		sendgridAdapter: sendgridAdapter,
	}
}
