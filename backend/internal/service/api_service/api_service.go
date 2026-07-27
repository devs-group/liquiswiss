//go:generate mockgen -package=mocks -destination ../../mocks/api_service.go liquiswiss/internal/service/api_service IAPIService
package api_service

import (
	"context"
	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/adapter/email_adapter"
	"liquiswiss/internal/events"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/reqctx"
	"time"
)

type IAPIService interface {
	Login(ctx context.Context, payload models.Login, deviceName string, existingRefreshToken string) (*models.User, *string, *time.Time, *string, *time.Time, error)
	Logout(ctx context.Context, existingRefreshToken string)
	ForgotPassword(ctx context.Context, payload models.ForgotPassword, code string) error
	ResetPassword(ctx context.Context, payload models.ResetPassword) error
	CheckResetPasswordCode(ctx context.Context, payload models.CheckResetPasswordCode) error
	CreateRegistration(ctx context.Context, payload models.CreateRegistration, code string) (int64, error)
	CheckRegistrationCode(ctx context.Context, payload models.CheckRegistrationCode, validity time.Duration) (int64, error)
	FinishRegistration(ctx context.Context, payload models.FinishRegistration, deviceName string, validity time.Duration) (*models.User, *string, *time.Time, *string, *time.Time, error)
	DeleteRegistration(ctx context.Context, registrationID int64, email string) error

	GetProfile(ctx context.Context, userID int64) (*models.User, error)
	UpdateProfile(ctx context.Context, payload models.UpdateUser, userID int64) (*models.User, error)
	UpdatePassword(ctx context.Context, payload models.UpdateUserPassword, userID int64) error
	SetUserCurrentOrganisation(ctx context.Context, payload models.UpdateUserCurrentOrganisation, userID int64) error
	GetCurrentOrganisation(ctx context.Context, userID int64) (*models.Organisation, error)

	ListTransactions(ctx context.Context, userID int64, page int64, limit int64, sortBy string, sortOrder string, search string, hideDisabled bool, hideExpired bool) ([]models.Transaction, int64, error)
	GetTransaction(ctx context.Context, userID int64, transactionID int64) (*models.Transaction, error)
	CreateTransaction(ctx context.Context, payload models.CreateTransaction, userID int64) (*models.Transaction, error)
	UpdateTransaction(ctx context.Context, payload models.UpdateTransaction, userID int64, transactionID int64) (*models.Transaction, error)
	DeleteTransaction(ctx context.Context, userID int64, transactionID int64) error

	ListOrganisations(ctx context.Context, userID int64, page int64, limit int64) ([]models.Organisation, int64, error)
	GetOrganisation(ctx context.Context, userID int64, organisationID int64) (*models.Organisation, error)
	CreateOrganisation(ctx context.Context, payload models.CreateOrganisation, userID int64) (*models.Organisation, error)
	UpdateOrganisation(ctx context.Context, payload models.UpdateOrganisation, userID int64, organisationID int64) (*models.Organisation, error)

	ListEmployees(ctx context.Context, userID int64, page int64, limit int64, sortBy string, sortOrder string, search string, hideTerminated bool) ([]models.Employee, int64, error)
	GetEmployee(ctx context.Context, userID int64, employeeID int64) (*models.Employee, error)
	CreateEmployee(ctx context.Context, payload models.CreateEmployee, userID int64) (*models.Employee, error)
	UpdateEmployee(ctx context.Context, payload models.UpdateEmployee, userID int64, employeeID int64) (*models.Employee, error)
	DeleteEmployee(ctx context.Context, userID int64, employeeID int64) error
	CountEmployees(ctx context.Context, userID int64, page int64, limit int64) (int64, error)

	ListSalaries(ctx context.Context, userID int64, employeeID int64, page int64, limit int64) ([]models.Salary, int64, error)
	GetSalary(ctx context.Context, userID int64, salaryID int64) (*models.Salary, error)
	CreateSalary(ctx context.Context, payload models.CreateSalary, userID int64, employeeID int64) (*models.Salary, error)
	UpdateSalary(ctx context.Context, payload models.UpdateSalary, userID int64, salaryID int64) (*models.Salary, error)
	DeleteSalary(ctx context.Context, userID int64, salaryID int64) error

	ListSalaryCosts(ctx context.Context, userID int64, salaryID int64, page int64, limit int64, skipPrevious bool) ([]models.SalaryCost, int64, error)
	GetSalaryCost(ctx context.Context, userID int64, salaryCostID int64, skipPrevious bool) (*models.SalaryCost, error)
	CreateSalaryCost(ctx context.Context, payload models.CreateSalaryCost, userID int64, salaryID int64) (*models.SalaryCost, error)
	UpdateSalaryCost(ctx context.Context, payload models.CreateSalaryCost, userID int64, salaryCostID int64) (*models.SalaryCost, error)
	DeleteSalaryCost(ctx context.Context, userID int64, salaryCostID int64) error
	CopySalaryCosts(ctx context.Context, payload models.CopySalaryCosts, userID int64, salaryID int64) error

	//ListSalaryCostDetails(salaryCostID int64) ([]models.SalaryCostDetail, error)
	//CalculateSalaryCostDetails(salaryCostID int64, userID int64) error
	//UpsertSalaryCostDetails(payload models.CreateSalaryCostDetail) (int64, error)
	//RefreshSalaryCostDetails(userID int64, salaryID int64) error

	ListSalaryCostLabels(ctx context.Context, userID int64, page int64, limit int64) ([]models.SalaryCostLabel, int64, error)
	GetSalaryCostLabel(ctx context.Context, userID int64, salaryCostLabelID int64) (*models.SalaryCostLabel, error)
	CreateSalaryCostLabel(ctx context.Context, payload models.CreateSalaryCostLabel, userID int64) (*models.SalaryCostLabel, error)
	UpdateSalaryCostLabel(ctx context.Context, payload models.CreateSalaryCostLabel, userID int64, salaryCostLabelID int64) (*models.SalaryCostLabel, error)
	DeleteSalaryCostLabel(ctx context.Context, userID int64, salaryCostLabelID int64) error

	ListForecasts(ctx context.Context, userID int64, limit int64) ([]models.Forecast, error)
	ListForecastDetails(ctx context.Context, userID int64, limit int64) ([]models.ForecastDatabaseDetails, error)
	ListForecastExclusions(ctx context.Context, userID int64, relatedID int64, relatedTable string) (map[string]bool, error)
	ListAllForecastExclusions(ctx context.Context, userID int64) ([]models.ForecastExclusionInfo, error)
	CreateForecastExclusion(ctx context.Context, payload models.CreateForecastExclusion, userID int64) (int64, error)
	DeleteForecastExclusion(ctx context.Context, payload models.CreateForecastExclusion, userID int64) (int64, error)
	UpdateForecastExclusions(ctx context.Context, payload models.UpdateForecastExclusions, userID int64) error
	CalculateForecast(ctx context.Context, userID int64) ([]models.Forecast, error)

	ListBankAccounts(ctx context.Context, userID int64, page int64, limit int64, sortBy string, sortOrder string, search string) ([]models.BankAccount, int64, error)
	GetBankAccount(ctx context.Context, userID int64, bankAccountID int64) (*models.BankAccount, error)
	CreateBankAccount(ctx context.Context, payload models.CreateBankAccount, userID int64) (*models.BankAccount, error)
	UpdateBankAccount(ctx context.Context, payload models.UpdateBankAccount, userID int64, bankAccountID int64) (*models.BankAccount, error)
	DeleteBankAccount(ctx context.Context, userID int64, bankAccountID int64) error

	ListVats(ctx context.Context, userID int64) ([]models.Vat, error)
	GetVat(ctx context.Context, userID int64, vatID int64) (*models.Vat, error)
	CreateVat(ctx context.Context, payload models.CreateVat, userID int64) (*models.Vat, error)
	UpdateVat(ctx context.Context, payload models.UpdateVat, userID int64, vatID int64) (*models.Vat, error)
	DeleteVat(ctx context.Context, userID int64, vatID int64) error

	GetVatSetting(ctx context.Context, userID int64) (*models.VatSetting, error)
	CreateVatSetting(ctx context.Context, payload models.CreateVatSetting, userID int64) (*models.VatSetting, error)
	UpdateVatSetting(ctx context.Context, payload models.UpdateVatSetting, userID int64) (*models.VatSetting, error)
	DeleteVatSetting(ctx context.Context, userID int64) error

	GetUserSetting(ctx context.Context, userID int64) (*models.UserSetting, error)
	UpdateUserSetting(ctx context.Context, payload models.UpdateUserSetting, userID int64) (*models.UserSetting, error)

	GetUserOrganisationSetting(ctx context.Context, userID int64) (*models.UserOrganisationSetting, error)
	UpdateUserOrganisationSetting(ctx context.Context, payload models.UpdateUserOrganisationSetting, userID int64) (*models.UserOrganisationSetting, error)

	ListCategories(ctx context.Context, userID, page, limit int64) ([]models.Category, int64, error)
	GetCategory(ctx context.Context, userID int64, categoryID int64) (*models.Category, error)
	CreateCategory(ctx context.Context, payload models.CreateCategory, userID *int64) (*models.Category, error)
	UpdateCategory(ctx context.Context, payload models.UpdateCategory, userID int64, categoryID int64) (*models.Category, error)
	DeleteCategory(ctx context.Context, userID int64, categoryID int64) error
	ReassignCategoryTransactions(ctx context.Context, userID int64, fromCategoryID int64, toCategoryID int64) (int64, error)

	ListCurrencies(ctx context.Context, userID int64) ([]models.Currency, error)
	GetCurrency(ctx context.Context, currencyID int64) (*models.Currency, error)
	CreateCurrency(ctx context.Context, payload models.CreateCurrency) (*models.Currency, error)
	UpdateCurrency(ctx context.Context, payload models.UpdateCurrency, currencyID int64) (*models.Currency, error)
	CountCurrencies(ctx context.Context) (int64, error)

	ListFiatRates(ctx context.Context, base string) ([]models.FiatRate, error)
	GetFiatRate(ctx context.Context, base, target string) (*models.FiatRate, error)
	UpsertFiatRate(ctx context.Context, payload models.CreateFiatRate) error
	CountUniqueCurrenciesInFiatRates(ctx context.Context) (int64, error)

	ListOrganisationInvitations(ctx context.Context, userID int64, organisationID int64) ([]models.Invitation, error)
	CreateOrganisationInvitation(ctx context.Context, payload models.CreateInvitation, userID int64, organisationID int64) (*models.Invitation, error)
	DeleteOrganisationInvitation(ctx context.Context, userID int64, organisationID int64, invitationID int64) error
	ResendOrganisationInvitation(ctx context.Context, userID int64, organisationID int64, invitationID int64) error
	CheckInvitation(ctx context.Context, token string) (*models.CheckInvitationResponse, error)
	AcceptInvitation(ctx context.Context, payload models.AcceptInvitation, deviceName string, authenticatedUserID int64) (*models.User, *string, *time.Time, *string, *time.Time, error)
	ListMyPendingInvitations(ctx context.Context, userID int64) ([]models.UserPendingInvitation, error)
	DeclineMyInvitation(ctx context.Context, userID int64, invitationID int64) error

	ListOrganisationMembers(ctx context.Context, userID int64, organisationID int64) ([]models.OrganisationMember, error)
	UpdateOrganisationMember(ctx context.Context, payload models.UpdateMember, userID int64, organisationID int64, memberUserID int64) error
	RemoveOrganisationMember(ctx context.Context, userID int64, organisationID int64, memberUserID int64) error

	SetEventHub(hub *events.Hub)
}

type APIService struct {
	dbService    db_adapter.IDatabaseAdapter
	emailAdapter email_adapter.IEmailAdapter
	eventHub     *events.Hub
}

func NewAPIService(dbService db_adapter.IDatabaseAdapter, emailAdapter email_adapter.IEmailAdapter) IAPIService {
	return &APIService{
		dbService:    dbService,
		emailAdapter: emailAdapter,
	}
}

// SetEventHub wires the real-time event hub. A nil hub (default, e.g. in tests)
// turns all event publishing into a no-op.
func (a *APIService) SetEventHub(hub *events.Hub) {
	a.eventHub = hub
}

// notifyChange publishes a change event scoped to the acting user's current
// organisation. Failures only affect real-time refresh, never the mutation.
func (a *APIService) notifyChange(ctx context.Context, userID int64, entity string, action string, id int64) {
	a.notifyChangeWithParent(ctx, userID, entity, action, id, 0)
}

// notifyChangeWithParent additionally links the event to a parent entity id
// (e.g. salary cost → owning salary) for targeted client-side highlighting
func (a *APIService) notifyChangeWithParent(ctx context.Context, userID int64, entity string, action string, id int64, parentID int64) {
	if a.eventHub == nil {
		return
	}
	user, err := a.dbService.GetProfile(userID)
	if err != nil {
		logger.Logger.Warnf("events: could not resolve organisation for user %d: %v", userID, err)
		return
	}
	a.eventHub.Publish(events.Event{
		Entity:         entity,
		Action:         action,
		ID:             id,
		ParentID:       parentID,
		OrganisationID: user.CurrentOrganisationID,
		OriginUserID:   userID,
		OriginClientID: reqctx.ClientID(ctx),
	})
}

// notifyOrganisationChange publishes a change event for an explicit
// organisation (invitation/member mutations carry the org id directly).
// actorUserID is the user who performed the mutation (event origin).
func (a *APIService) notifyOrganisationChange(ctx context.Context, actorUserID int64, organisationID int64, entity string, action string, id int64) {
	if a.eventHub == nil || organisationID == 0 {
		return
	}
	a.eventHub.Publish(events.Event{
		Entity:         entity,
		Action:         action,
		ID:             id,
		OrganisationID: organisationID,
		OriginUserID:   actorUserID,
		OriginClientID: reqctx.ClientID(ctx),
	})
}

// closeUserStreams force-closes a user's event streams (logout, org switch,
// membership changes) so stale subscriptions cannot leak events.
func (a *APIService) closeUserStreams(userID int64) {
	if a.eventHub == nil {
		return
	}
	a.eventHub.CloseUser(userID)
}
