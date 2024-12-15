//go:generate mockgen -package=mocks -destination ../mocks/db_service.go liquiswiss/internal/service IDatabaseService
package service

import (
	"bytes"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/types"
	"liquiswiss/pkg/utils"
	"log"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

//go:embed queries/*.sql
var sqlQueries embed.FS

//go:embed mocks/*.sql
var sqlMocks embed.FS

var allowedSortOrders = map[string]bool{
	"ASC": true, "DESC": true,
}

type IDatabaseService interface {
	ApplyMocks() error

	CreateRegistration(email, code string) (int64, error)
	ValidateRegistration(email, code string, hours time.Duration) (int64, error)
	DeleteRegistration(registrationID int64, email string) error

	CreateUser(email, password string) (int64, error)
	GetUserPasswordByEMail(email string) (*models.Login, error)
	GetProfile(id string) (*models.User, error)
	UpdateProfile(payload models.UpdateUser, userID string) error
	UpdatePassword(password string, userID string) error
	ResetPassword(password string, userID string) error
	CheckUserExistance(id int64) (bool, error)
	SetUserCurrentOrganisation(userID int64, organisationID int64) error
	CreateResetPassword(email, code string, delay time.Duration) (bool, error)
	ValidateResetPassword(email, code string, hours time.Duration) (int64, error)
	DeleteResetPassword(email string) error

	ListTransactions(userID int64, page int64, limit int64, sortBy string, sortOrder string) ([]models.Transaction, int64, error)
	GetTransaction(userID int64, transactionID string) (*models.Transaction, error)
	CreateTransaction(payload models.CreateTransaction, userID int64) (int64, error)
	UpdateTransaction(payload models.UpdateTransaction, userID int64, transactionID string) error
	DeleteTransaction(userID int64, transactionID string) error

	ListOrganisations(userID int64, page int64, limit int64) ([]models.Organisation, int64, error)
	GetOrganisation(userID int64, id string) (*models.Organisation, error)
	CreateOrganisation(name string) (int64, error)
	UpdateOrganisation(payload models.UpdateOrganisation, userID int64, organisationID string) error
	AssignUserToOrganisation(userID int64, organisationID int64, role string, isDefault bool) error

	ListEmployees(userID int64, page int64, limit int64, sortBy string, sortOrder string) ([]models.Employee, int64, error)
	ListEmployeeHistory(userID int64, employeeID string, page int64, limit int64) ([]models.EmployeeHistory, int64, error)
	CountEmployees(userID int64, page int64, limit int64) (int64, error)
	GetEmployee(userID int64, id string) (*models.Employee, error)
	GetEmployeeHistory(userID int64, historyID string) (*models.EmployeeHistory, error)
	CreateEmployee(payload models.CreateEmployee, userID int64) (int64, error)
	CreateEmployeeHistory(payload models.CreateEmployeeHistory, userID int64, employeeID string) (int64, error)
	UpdateEmployee(payload models.UpdateEmployee, userID int64, employeeID string) error
	UpdateEmployeeHistory(payload models.UpdateEmployeeHistory, employeeID int64, historyID string) error
	DeleteEmployee(employeeID int64, userID int64) error
	DeleteEmployeeHistory(existingEmployeeHistory *models.EmployeeHistory, userID int64) error

	ListForecasts(userID int64, limit int64) ([]models.Forecast, error)
	ListForecastDetails(userID int64, limit int64) ([]models.ForecastDatabaseDetails, error)
	UpsertForecast(payload models.CreateForecast, userID int64) (int64, error)
	UpsertForecastDetail(payload models.CreateForecastDetail, userID, forecastID int64) (int64, error)
	ClearForecasts(userID int64) (int64, error)

	ListBankAccounts(userID int64) ([]models.BankAccount, error)
	GetBankAccount(userID int64, bankAccountID string) (*models.BankAccount, error)
	CreateBankAccount(payload models.CreateBankAccount, userID int64) (int64, error)
	UpdateBankAccount(payload models.UpdateBankAccount, userID int64, bankAccountID string) error
	DeleteBankAccount(userID int64, bankAccountID string) error

	ListVats(userID int64) ([]models.Vat, error)
	GetVat(userID int64, vatID string) (*models.Vat, error)
	CreateVat(payload models.CreateVat, userID int64) (int64, error)
	UpdateVat(payload models.UpdateVat, userID int64, vatID string) error
	DeleteVat(userID int64, vatID string) error

	ListCategories(userID, page, limit int64) ([]models.Category, int64, error)
	GetCategory(userID int64, id string) (*models.Category, error)
	CreateCategory(userID *int64, payload models.CreateCategory) (int64, error)
	UpdateCategory(userID int64, payload models.UpdateCategory, categoryID string) error

	ListCurrencies(page int64, limit int64) ([]models.Currency, int64, error)
	GetCurrency(id string) (*models.Currency, error)
	CreateCurrency(payload models.CreateCurrency) (int64, error)
	UpdateCurrency(payload models.UpdateCurrency, currencyID string) error

	StoreRefreshTokenID(userID int64, tokenId string, expirationTime time.Time, deviceName string) error
	CheckRefreshToken(tokenID string, userID int64) (bool, error)
	DeleteRefreshToken(tokenID string, userID int64) error

	ListFiatRates(base string) ([]models.FiatRate, error)
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

func (s *DatabaseService) ApplyMocks() error {
	return fs.WalkDir(sqlMocks, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if utils.IsProduction() && strings.Contains(path, "dev_") {
			logger.Logger.Infof(`Skipping mock "%s"`, path)
			return nil
		}
		if filepath.Ext(path) == ".sql" {
			query, err := sqlMocks.ReadFile(path)
			if err != nil {
				logger.Logger.Errorf("Could not read %s: %v", path, err)
			}
			_, err = s.db.Exec(string(query))
			if err != nil {
				logger.Logger.Infof("Failed to apply %s to DB: %s", path, err.Error())
			} else {
				logger.Logger.Infof("Applied %s to DB", path)
			}
		}
		return nil
	})
}

func (s *DatabaseService) CreateRegistration(email, code string) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_registration.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(email, code, email)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted registration
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *DatabaseService) ValidateRegistration(email, code string, hours time.Duration) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/validate_registration.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var id int64
	err = stmt.QueryRow(email, code, hours.Hours()).Scan(&id)
	if err != nil {
		return 0, err
	}
	if id <= 0 {
		return 0, errors.New("ungültiges Resultat für Registrierung")
	}

	return id, nil
}

func (s *DatabaseService) DeleteRegistration(registrationID int64, email string) error {
	query, err := sqlQueries.ReadFile("queries/delete_registration.sql")
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(registrationID, email)
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) CreateUser(email string, password string) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_user.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(email, password)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted user
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *DatabaseService) GetUserPasswordByEMail(email string) (*models.Login, error) {
	var loginUser models.Login

	query, err := sqlQueries.ReadFile("queries/get_user_password_by_email.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), email).Scan(
		&loginUser.ID, &loginUser.Password,
	)
	if err != nil {
		return nil, err
	}

	return &loginUser, nil
}

func (s *DatabaseService) GetProfile(id string) (*models.User, error) {
	var user models.User

	query, err := sqlQueries.ReadFile("queries/get_profile.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), id).Scan(&user.ID, &user.Name, &user.Email, &user.CurrentOrganisationID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *DatabaseService) UpdateProfile(payload models.UpdateUser, userID string) error {
	// Base query
	query := "UPDATE users SET "
	queryBuild := []string{}
	args := []interface{}{}

	// Dynamically add fields that are not nil
	if payload.Name != nil {
		queryBuild = append(queryBuild, "name = ?")
		args = append(args, *payload.Name)
	}
	if payload.Email != nil {
		queryBuild = append(queryBuild, "email = ?")
		args = append(args, *payload.Email)
	}

	// Add WHERE clause
	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ?"
	args = append(args, userID)

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) UpdatePassword(password string, userID string) error {
	// Base query
	query := "UPDATE users SET "
	queryBuild := []string{}
	args := []interface{}{}

	// Dynamically add fields that are not nil
	queryBuild = append(queryBuild, "password = ?")
	args = append(args, password)

	// Add WHERE clause
	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ?"
	args = append(args, userID)

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}

	return nil
}

// ResetPassword is only called internally and only after the reset password has been confirmed
func (s *DatabaseService) ResetPassword(password string, email string) error {
	// Base query
	query := "UPDATE users SET "
	queryBuild := []string{}
	args := []interface{}{}

	// Dynamically add fields that are not nil
	queryBuild = append(queryBuild, "password = ?")
	args = append(args, password)

	// Add WHERE clause
	query += strings.Join(queryBuild, ", ")
	query += " WHERE email = ?"
	args = append(args, email)

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) CheckUserExistance(id int64) (bool, error) {
	query, err := sqlQueries.ReadFile("queries/check_user_existence.sql")
	if err != nil {
		return false, err
	}

	var exists bool
	err = s.db.QueryRow(string(query), id).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *DatabaseService) SetUserCurrentOrganisation(userID int64, organisationID int64) error {
	query, err := sqlQueries.ReadFile("queries/set_user_current_organisation.sql")
	if err != nil {
		return err
	}

	res, err := s.db.Exec(string(query), organisationID, userID, organisationID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) CreateResetPassword(email, code string, delay time.Duration) (bool, error) {
	query, err := sqlQueries.ReadFile("queries/create_reset_password.sql")
	if err != nil {
		return false, err
	}

	res, err := s.db.Exec(string(query), email, code, email, email, delay.Minutes())
	if err != nil {
		return false, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

func (s *DatabaseService) ValidateResetPassword(email, code string, hours time.Duration) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/validate_reset_password.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var id int64
	err = stmt.QueryRow(email, code, hours.Hours()).Scan(&id)
	if err != nil {
		return 0, err
	}
	if id <= 0 {
		return 0, errors.New("ungültiges Resultat für Passwort Zurücksetzen")
	}

	return id, nil
}

func (s *DatabaseService) DeleteResetPassword(email string) error {
	query, err := sqlQueries.ReadFile("queries/delete_reset_password.sql")
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(email)
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) CreateTransaction(payload models.CreateTransaction, userID int64) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_transaction.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		payload.Name, payload.Amount, payload.Cycle, payload.Type, payload.StartDate, payload.EndDate,
		payload.Category, payload.Currency, payload.Employee, userID, payload.Vat, payload.VatIncluded,
	)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted transaction
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *DatabaseService) UpdateTransaction(payload models.UpdateTransaction, userID int64, transactionID string) error {
	// Base query
	query := "UPDATE transactions SET "
	queryBuild := []string{}
	args := []interface{}{}

	// Dynamically add fields that are not nil
	if payload.Name != nil {
		queryBuild = append(queryBuild, "name = ?")
		args = append(args, *payload.Name)
	}
	if payload.Amount != nil {
		queryBuild = append(queryBuild, "amount = ?")
		args = append(args, *payload.Amount)
	}
	// Cycle is also always considered
	queryBuild = append(queryBuild, "cycle = ?")
	if payload.Cycle != nil {
		args = append(args, *payload.Cycle)
	} else {
		args = append(args, nil)
	}
	if payload.Type != nil {
		queryBuild = append(queryBuild, "type = ?")
		args = append(args, *payload.Type) // `type` might be a reserved keyword, hence the backticks
	}
	if payload.StartDate != nil {
		queryBuild = append(queryBuild, "start_date = ?")
		args = append(args, *payload.StartDate)
	}
	// Always consider EndDate in case it is set back to null
	queryBuild = append(queryBuild, "end_date = ?")
	if payload.EndDate != nil {
		endDate, err := time.Parse(utils.InternalDateFormat, *payload.EndDate)
		if err != nil {
			return err
		}
		args = append(args, endDate)
	} else {
		args = append(args, nil)
	}
	if payload.Category != nil {
		queryBuild = append(queryBuild, "category_id = ?")
		args = append(args, *payload.Category)
	}
	if payload.Currency != nil {
		queryBuild = append(queryBuild, "currency_id = ?")
		args = append(args, *payload.Currency)
	}
	queryBuild = append(queryBuild, "employee_id = ?")
	if payload.Employee != nil {
		args = append(args, *payload.Employee)
	} else {
		args = append(args, nil)
	}
	queryBuild = append(queryBuild, "vat_id = ?")
	if payload.Vat != nil {
		args = append(args, *payload.Vat)
	} else {
		args = append(args, nil)
	}
	if payload.VatIncluded != nil {
		queryBuild = append(queryBuild, "vat_included = ?")
		args = append(args, *payload.VatIncluded)
	}

	// Add WHERE clause
	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ? AND organisation_id = (SELECT current_organisation FROM users u WHERE u.id = ?)"
	args = append(args, transactionID)
	args = append(args, userID)

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) DeleteTransaction(userID int64, transactionID string) error {
	query, err := sqlQueries.ReadFile("queries/delete_transaction.sql")
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(transactionID, userID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) ListTransactions(userID int64, page int64, limit int64, sortBy string, sortOrder string) ([]models.Transaction, int64, error) {
	transactions := []models.Transaction{}
	var totalCount int64
	sortByMap := map[string]string{
		"name": "r.name", "startDate": "r.start_date", "endDate": "r.end_date", "amount": "r.amount",
		"cycle": "r.cycle", "category": "c.name", "employee": "emp.name",
	}

	// Validate inputs
	sortBy = sortByMap[sortBy]
	if sortBy == "" || !allowedSortOrders[sortOrder] {
		return nil, 0, fmt.Errorf("invalid sort by or sort order")
	}

	parsed, err := template.ParseFS(sqlQueries, "queries/list_transactions.sql")
	if err != nil {
		return nil, 0, err
	}

	var query bytes.Buffer
	err = parsed.Execute(&query, map[string]string{
		"sortBy":    sortBy,
		"sortOrder": sortOrder,
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(query.String(), userID, (page)*limit, 0)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var transaction models.Transaction
		// These are required for proper date convertion afterwards
		var startDate time.Time
		var endDate sql.NullTime
		var nextExecutionDate sql.NullTime
		var transactionEmployeeID sql.NullInt64
		var transactionEmployeeName sql.NullString
		var vatID sql.NullInt64
		var vatValue sql.NullInt64
		var vatFormattedValue sql.NullString
		var vatCanEdit sql.NullBool

		err := rows.Scan(
			&transaction.ID, &transaction.Name, &transaction.Amount, &transaction.VatAmount, &transaction.VatIncluded,
			&transaction.Cycle, &transaction.Type, &startDate, &endDate,
			&transaction.Category.ID, &transaction.Category.Name,
			&transaction.Currency.ID, &transaction.Currency.Code, &transaction.Currency.Description, &transaction.Currency.LocaleCode,
			&transactionEmployeeID, &transactionEmployeeName,
			&vatID, &vatValue, &vatFormattedValue, &vatCanEdit,
			&nextExecutionDate, &totalCount,
		)
		if err != nil {
			return nil, 0, err
		}

		transaction.StartDate = types.AsDate(startDate)

		if endDate.Valid {
			convertedDate := types.AsDate(endDate.Time)
			transaction.EndDate = &convertedDate
		}

		if nextExecutionDate.Valid {
			convertedDate := types.AsDate(nextExecutionDate.Time)
			transaction.NextExecutionDate = &convertedDate
		}

		if transactionEmployeeID.Valid {
			transaction.Employee = &models.TransactionEmployee{
				ID:   transactionEmployeeID.Int64,
				Name: transactionEmployeeName.String,
			}
		}

		if vatID.Valid {
			transaction.Vat = &models.Vat{
				ID:             vatID.Int64,
				Value:          vatValue.Int64,
				FormattedValue: vatFormattedValue.String,
				CanEdit:        vatCanEdit.Bool,
			}
		}

		transactions = append(transactions, transaction)
	}

	return transactions, totalCount, nil
}

func (s *DatabaseService) GetTransaction(userID int64, transactionID string) (*models.Transaction, error) {
	var transaction models.Transaction
	// These are required for proper date convertion afterwards
	var startDate time.Time
	var endDate sql.NullTime
	var transactionEmployeeID sql.NullInt64
	var transactionEmployeeName sql.NullString
	var vatID sql.NullInt64
	var vatValue sql.NullInt64
	var vatFormattedValue sql.NullString
	var vatCanEdit sql.NullBool

	query, err := sqlQueries.ReadFile("queries/get_transaction.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), transactionID, userID).Scan(
		&transaction.ID, &transaction.Name, &transaction.Amount, &transaction.Cycle, &transaction.Type, &startDate, &endDate,
		&transaction.Category.ID, &transaction.Category.Name,
		&transaction.Currency.ID, &transaction.Currency.Code, &transaction.Currency.Description, &transaction.Currency.LocaleCode,
		&transactionEmployeeID, &transactionEmployeeName,
		&vatID, &vatValue, &vatFormattedValue, &vatCanEdit,
	)
	if err != nil {
		return nil, err
	}

	transaction.StartDate = types.AsDate(startDate)

	if endDate.Valid {
		convertedDate := types.AsDate(endDate.Time)
		transaction.EndDate = &convertedDate
	}

	if transactionEmployeeID.Valid {
		transaction.Employee = &models.TransactionEmployee{
			ID:   transactionEmployeeID.Int64,
			Name: transactionEmployeeName.String,
		}
	}

	if vatID.Valid {
		transaction.Vat = &models.Vat{
			ID:             vatID.Int64,
			Value:          vatValue.Int64,
			FormattedValue: vatFormattedValue.String,
			CanEdit:        vatCanEdit.Bool,
		}
	}

	return &transaction, nil
}

func (s *DatabaseService) ListOrganisations(userID int64, page int64, limit int64) ([]models.Organisation, int64, error) {
	organisations := []models.Organisation{}
	var totalCount int64

	query, err := sqlQueries.ReadFile("queries/list_organisations.sql")
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(string(query), userID, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var organisation models.Organisation

		err := rows.Scan(
			&organisation.ID, &organisation.Name, &organisation.MemberCount,
			&organisation.Role, &organisation.IsDefault,
			&totalCount,
		)
		if err != nil {
			return nil, 0, err
		}

		organisations = append(organisations, organisation)
	}

	return organisations, totalCount, nil
}

func (s *DatabaseService) GetOrganisation(userID int64, id string) (*models.Organisation, error) {
	var organisation models.Organisation

	query, err := sqlQueries.ReadFile("queries/get_organisation.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), id, userID).Scan(
		&organisation.ID, &organisation.Name, &organisation.MemberCount, &organisation.Role,
	)
	if err != nil {
		return nil, err
	}

	return &organisation, nil
}

// CreateOrganisation implements IDatabaseService.
func (s *DatabaseService) CreateOrganisation(name string) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_organisation.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(name)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted user
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *DatabaseService) UpdateOrganisation(payload models.UpdateOrganisation, userID int64, organisationID string) error {
	// Base query
	query := "UPDATE organisations SET "
	queryBuild := []string{}
	args := []interface{}{}

	// Dynamically add fields that are not nil
	if payload.Name != nil {
		queryBuild = append(queryBuild, "name = ?")
		args = append(args, *payload.Name)
	}

	// Add WHERE clause
	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ?"
	args = append(args, organisationID)

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) AssignUserToOrganisation(userID int64, organisationID int64, role string, isDefault bool) error {
	query, err := sqlQueries.ReadFile("queries/assign_user_to_organisation.sql")
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, organisationID, role, isDefault)
	if err != nil {
		return err
	}

	return nil
}

// ListEmployees implements employee listing with pagination
func (s *DatabaseService) ListEmployees(userID int64, page int64, limit int64, sortBy string, sortOrder string) ([]models.Employee, int64, error) {
	employees := make([]models.Employee, 0)
	var totalCount int64
	sortByMap := map[string]string{
		"name": "e.name", "hoursPerMonth": "h.hours_per_month", "salaryPerMonth": "h.salary_per_month", "vacationDaysPerYear": "h.vacation_days_per_year",
		"fromDate": "h.from_date", "toDate": "h.to_date",
	}

	// Validate inputs
	sortBy = sortByMap[sortBy]
	if sortBy == "" || !allowedSortOrders[sortOrder] {
		return nil, 0, fmt.Errorf("invalid sort by or sort order")
	}

	query, err := sqlQueries.ReadFile("queries/list_employees.sql")
	if err != nil {
		return nil, 0, err
	}

	queryString := fmt.Sprintf(string(query), sortBy, sortBy, sortOrder)

	rows, err := s.db.Query(queryString, userID, (page)*limit, 0)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var employee models.Employee
		var fromDate sql.NullTime
		var toDate sql.NullTime
		var isInFuture sql.NullBool

		employee.Currency = &models.Currency{}

		err := rows.Scan(
			&employee.ID, &employee.Name, &employee.HoursPerMonth, &employee.SalaryPerMonth,
			&employee.Currency.ID, &employee.Currency.LocaleCode, &employee.Currency.Description,
			&employee.Currency.Code, &employee.VacationDaysPerYear, &fromDate, &toDate, &isInFuture, &totalCount,
		)
		if err != nil {
			return nil, 0, err
		}

		if fromDate.Valid {
			convertedDate := types.AsDate(fromDate.Time)
			employee.FromDate = &convertedDate
		}
		if toDate.Valid {
			convertedDate := types.AsDate(toDate.Time)
			employee.ToDate = &convertedDate
		}
		if isInFuture.Valid {
			employee.IsInFuture = isInFuture.Bool
		} else {
			employee.IsInFuture = false
		}

		employees = append(employees, employee)
	}

	return employees, totalCount, nil
}

// ListEmployeeHistory implements listing the history of an employee
func (s *DatabaseService) ListEmployeeHistory(userID int64, employeeID string, page int64, limit int64) ([]models.EmployeeHistory, int64, error) {
	employeeHistories := make([]models.EmployeeHistory, 0)
	var totalCount int64

	query, err := sqlQueries.ReadFile("queries/list_employee_history.sql")
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(string(query), employeeID, userID, (page)*limit, 0)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var employeeHistory models.EmployeeHistory
		var fromDate sql.NullTime
		var toDate sql.NullTime

		employeeHistory.Currency = models.Currency{}

		err := rows.Scan(
			&employeeHistory.ID, &employeeHistory.EmployeeID, &employeeHistory.HoursPerMonth, &employeeHistory.SalaryPerMonth,
			&employeeHistory.Currency.ID, &employeeHistory.Currency.LocaleCode, &employeeHistory.Currency.Description,
			&employeeHistory.Currency.Code,
			&employeeHistory.VacationDaysPerYear, &fromDate, &toDate, &totalCount,
		)
		if err != nil {
			return nil, 0, err
		}

		if fromDate.Valid {
			convertedDate := types.AsDate(fromDate.Time)
			employeeHistory.FromDate = convertedDate
		}
		if toDate.Valid {
			convertedDate := types.AsDate(toDate.Time)
			employeeHistory.ToDate = &convertedDate
		}

		employeeHistories = append(employeeHistories, employeeHistory)
	}

	return employeeHistories, totalCount, nil
}

// CountEmployees implements employee listing with pagination
func (s *DatabaseService) CountEmployees(userID int64, page int64, limit int64) (int64, error) {
	var totalCount int64

	query, err := sqlQueries.ReadFile("queries/count_employees.sql")
	if err != nil {
		return 0, err
	}

	rows, err := s.db.Query(string(query), userID, limit, (page-1)*limit)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {

		err := rows.Scan(&totalCount)
		if err != nil {
			return 0, err
		}
	}

	return totalCount, nil
}

// GetEmployee implements fetching an employee by ID
func (s *DatabaseService) GetEmployee(userID int64, id string) (*models.Employee, error) {
	var employee models.Employee
	var fromDate sql.NullTime
	var toDate sql.NullTime
	var isInFuture sql.NullBool

	employee.Currency = &models.Currency{}

	query, err := sqlQueries.ReadFile("queries/get_employee.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), id, userID).Scan(
		&employee.ID, &employee.Name, &employee.HoursPerMonth, &employee.SalaryPerMonth,
		&employee.Currency.ID, &employee.Currency.LocaleCode, &employee.Currency.Description,
		&employee.Currency.Code, &employee.VacationDaysPerYear, &fromDate, &toDate, &isInFuture, &employee.HistoryID,
	)
	if err != nil {
		return nil, err
	}

	if fromDate.Valid {
		convertedDate := types.AsDate(fromDate.Time)
		employee.FromDate = &convertedDate
	}
	if toDate.Valid {
		convertedDate := types.AsDate(toDate.Time)
		employee.ToDate = &convertedDate
	}
	if isInFuture.Valid {
		employee.IsInFuture = isInFuture.Bool
	} else {
		employee.IsInFuture = false
	}

	return &employee, nil
}

// GetEmployeeHistory implements fetching an employees history by ID
func (s *DatabaseService) GetEmployeeHistory(userID int64, historyID string) (*models.EmployeeHistory, error) {
	var employeeHistory models.EmployeeHistory
	var fromDate time.Time
	var toDate sql.NullTime

	employeeHistory.Currency = models.Currency{}

	query, err := sqlQueries.ReadFile("queries/get_employee_history.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), historyID, userID).Scan(
		&employeeHistory.ID, &employeeHistory.EmployeeID, &employeeHistory.HoursPerMonth, &employeeHistory.SalaryPerMonth,
		&employeeHistory.Currency.ID, &employeeHistory.Currency.LocaleCode, &employeeHistory.Currency.Description,
		&employeeHistory.Currency.Code, &employeeHistory.VacationDaysPerYear, &fromDate, &toDate,
	)
	if err != nil {
		return nil, err
	}

	employeeHistory.FromDate = types.AsDate(fromDate)

	if toDate.Valid {
		convertedDate := types.AsDate(toDate.Time)
		employeeHistory.ToDate = &convertedDate
	}

	return &employeeHistory, nil
}

// CreateEmployee implements the creation of a new employee
func (s *DatabaseService) CreateEmployee(payload models.CreateEmployee, userID int64) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_employee.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		payload.Name, userID,
	)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted employee
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// CreateEmployeeHistory implements the creation of a new employee
func (s *DatabaseService) CreateEmployeeHistory(payload models.CreateEmployeeHistory, userID int64, employeeID string) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	query, err := sqlQueries.ReadFile("queries/create_employee_history.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := tx.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	// Prepare entry and exit date
	fromDate, err := time.Parse(utils.InternalDateFormat, payload.FromDate)
	if err != nil {
		return 0, err
	}

	var toDate sql.NullTime
	if payload.ToDate != nil {
		parsedToDate, err := time.Parse(utils.InternalDateFormat, *payload.ToDate)
		if err != nil {
			return 0, err
		}
		toDate = sql.NullTime{Time: parsedToDate, Valid: true}
	} else {
		toDate = sql.NullTime{Valid: false}
	}

	res, err := stmt.Exec(
		employeeID, payload.HoursPerMonth, payload.SalaryPerMonth, payload.CurrencyID, payload.VacationDaysPerYear,
		fromDate, toDate, employeeID, userID,
	)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted history
	historyID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if historyID == 0 {
		return 0, sql.ErrNoRows
	}

	// Fetch and adjust previous entry if required
	var previous models.EmployeeHistory
	if err := tx.QueryRow(`
        SELECT id, from_date, to_date FROM employee_history 
        WHERE employee_id = ? AND from_date < ? AND (to_date IS NULL OR to_date >= ?) AND id != ?
        ORDER BY from_date DESC LIMIT 1
    `, employeeID, payload.FromDate, payload.FromDate, historyID).Scan(&previous.ID, &previous.FromDate, &previous.ToDate); err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if previous.ID != 0 {
		logger.Logger.Debugf("Checking previous history entry")
		currentFromDate, err := time.Parse(utils.InternalDateFormat, payload.FromDate)
		if err != nil {
			return 0, err
		}
		needsAdjustment := false
		if previous.ToDate == nil {
			needsAdjustment = true
		} else {
			previousToDate := time.Time(*previous.ToDate)
			if !previousToDate.Before(currentFromDate) {
				needsAdjustment = true
			}
		}

		if needsAdjustment {
			previousFromDate := time.Time(previous.FromDate)
			newToDate := utils.GetNextAvailableDate(previousFromDate, currentFromDate)
			logger.Logger.Debugf("Adjusting previous history entry to: %s", newToDate)
			_, err := tx.Exec(`
            UPDATE employee_history SET to_date = ? WHERE id = ?
        `, newToDate, previous.ID)
			if err != nil {
				return 0, err
			}
		}
	}

	// Fetch and adjust current entry if required by next
	var next models.EmployeeHistory
	if err := tx.QueryRow(`
	   SELECT id, from_date, to_date FROM employee_history
       WHERE employee_id = ? AND from_date > ? AND id != ?
	   ORDER BY from_date ASC LIMIT 1
	`, employeeID, payload.FromDate, historyID).Scan(&next.ID, &next.FromDate, &next.ToDate); err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if next.ID != 0 {
		logger.Logger.Debugf("Checking next history entry")
		needsAdjustment := false

		nextFromDate := time.Time(next.FromDate)
		if payload.ToDate == nil {
			needsAdjustment = true
		} else {
			currentToDate, err := time.Parse(utils.InternalDateFormat, *payload.ToDate)
			if err != nil {
				return 0, err
			}
			if !currentToDate.Before(nextFromDate) {
				needsAdjustment = true
			}
		}

		if needsAdjustment {
			currentFromDate, err := time.Parse(utils.InternalDateFormat, payload.FromDate)
			if err != nil {
				return 0, err
			}
			newToDate := utils.GetNextAvailableDate(currentFromDate, nextFromDate)
			logger.Logger.Debugf("Adjusting current history entry to: %s", newToDate)
			_, err = tx.Exec(`
            UPDATE employee_history SET to_date = ? WHERE id = ?
        `, newToDate, historyID)
			if err != nil {
				return 0, err
			}
		}
	}

	return historyID, nil
}

// UpdateEmployee implements updating employee details
func (s *DatabaseService) UpdateEmployee(payload models.UpdateEmployee, userID int64, employeeID string) error {
	// Base query
	query := "UPDATE employees SET "
	queryBuild := []string{}
	args := []interface{}{}

	// Dynamically add fields that are not nil
	if payload.Name != nil {
		queryBuild = append(queryBuild, "name = ?")
		args = append(args, *payload.Name)
	}

	// Add WHERE clause
	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ? AND organisation_id = (SELECT current_organisation FROM users u WHERE u.id = ?)"
	args = append(args, employeeID, userID)

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}

	return nil
}

// UpdateEmployeeHistory implements updating employee history details
func (s *DatabaseService) UpdateEmployeeHistory(payload models.UpdateEmployeeHistory, employeeID int64, historyID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Base query
	query := "UPDATE employee_history SET "
	queryBuild := []string{}
	args := []interface{}{}

	// Dynamically add fields that are not nil
	if payload.HoursPerMonth != nil {
		queryBuild = append(queryBuild, "hours_per_month = ?")
		args = append(args, *payload.HoursPerMonth)
	}
	if payload.SalaryPerMonth != nil {
		queryBuild = append(queryBuild, "salary_per_month = ?")
		args = append(args, *payload.SalaryPerMonth)
	}
	if payload.CurrencyID != nil {
		queryBuild = append(queryBuild, "currency_id = ?")
		args = append(args, *payload.CurrencyID)
	}
	if payload.VacationDaysPerYear != nil {
		queryBuild = append(queryBuild, "vacation_days_per_year = ?")
		args = append(args, *payload.VacationDaysPerYear)
	}
	if payload.FromDate != nil {
		queryBuild = append(queryBuild, "from_date = ?")
		entryDate, err := time.Parse(utils.InternalDateFormat, *payload.FromDate)
		if err != nil {
			return err
		}
		args = append(args, entryDate)
	}
	// Always consider ToDate in case it is set back to null
	queryBuild = append(queryBuild, "to_date = ?")
	if payload.ToDate != nil {
		exitDate, err := time.Parse(utils.InternalDateFormat, *payload.ToDate)
		if err != nil {
			return err
		}
		args = append(args, exitDate)
	} else {
		args = append(args, nil)
	}

	// Add WHERE clause
	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ? AND employee_id = ?"
	args = append(args, historyID, employeeID)

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}

	// Fetch and adjust previous entry if required
	var previous models.EmployeeHistory
	if err := tx.QueryRow(`
        SELECT id, from_date, to_date FROM employee_history 
        WHERE employee_id = ? AND from_date < ? AND (to_date IS NULL OR to_date >= ?) AND id != ?
        ORDER BY from_date DESC LIMIT 1
    `, employeeID, payload.FromDate, payload.FromDate, historyID).Scan(&previous.ID, &previous.FromDate, &previous.ToDate); err != nil && err != sql.ErrNoRows {
		return err
	}

	if previous.ID != 0 {
		logger.Logger.Debugf("Checking previous history entry")
		currentFromDate, err := time.Parse(utils.InternalDateFormat, *payload.FromDate)
		if err != nil {
			return err
		}
		needsAdjustment := false
		if previous.ToDate == nil {
			needsAdjustment = true
		} else {
			previousToDate := time.Time(*previous.ToDate)
			if !previousToDate.Before(currentFromDate) {
				needsAdjustment = true
			}
		}

		if needsAdjustment {
			previousFromDate := time.Time(previous.FromDate)
			newToDate := utils.GetNextAvailableDate(previousFromDate, currentFromDate)
			logger.Logger.Debugf("Adjusting previous history entry to: %s", newToDate)
			_, err := tx.Exec(`
            UPDATE employee_history SET to_date = ? WHERE id = ?
        `, newToDate, previous.ID)
			if err != nil {
				return err
			}
		}
	}

	// Fetch and adjust current entry if required by next
	var next models.EmployeeHistory
	if err := tx.QueryRow(`
	   SELECT id, from_date, to_date FROM employee_history
       WHERE employee_id = ? AND from_date > ? AND id != ?
	   ORDER BY from_date ASC LIMIT 1
	`, employeeID, payload.FromDate, historyID).Scan(&next.ID, &next.FromDate, &next.ToDate); err != nil && err != sql.ErrNoRows {
		return err
	}

	if next.ID != 0 {
		logger.Logger.Debugf("Checking next history entry")
		needsAdjustment := false

		nextFromDate := time.Time(next.FromDate)
		if payload.ToDate == nil {
			needsAdjustment = true
		} else {
			currentToDate, err := time.Parse(utils.InternalDateFormat, *payload.ToDate)
			if err != nil {
				return err
			}
			if !currentToDate.Before(nextFromDate) {
				needsAdjustment = true
			}
		}

		if needsAdjustment {
			currentFromDate, err := time.Parse(utils.InternalDateFormat, *payload.FromDate)
			if err != nil {
				return err
			}
			newToDate := utils.GetNextAvailableDate(currentFromDate, nextFromDate)
			logger.Logger.Debugf("Adjusting current history entry to: %s", newToDate)
			_, err = tx.Exec(`
            UPDATE employee_history SET to_date = ? WHERE id = ?
        `, newToDate, historyID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DeleteEmployee implements updating employee details
func (s *DatabaseService) DeleteEmployee(employeeID int64, userID int64) error {
	query, err := sqlQueries.ReadFile("queries/delete_employee.sql")
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(employeeID, userID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

// DeleteEmployeeHistory implements deleting an employee history entry
func (s *DatabaseService) DeleteEmployeeHistory(toDeleteEmployeeHistory *models.EmployeeHistory, userID int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	query, err := sqlQueries.ReadFile("queries/delete_employee_history.sql")
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(toDeleteEmployeeHistory.ID, userID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	// In case there is a previous and next entry, adjust the previous toDate related to the next one
	currentFromDate := time.Time(toDeleteEmployeeHistory.FromDate)
	currentFromDateFormatted := currentFromDate.Format(utils.InternalDateFormat)
	var previous models.EmployeeHistory
	if err := tx.QueryRow(`
        SELECT id, from_date, to_date FROM employee_history 
        WHERE employee_id = ? AND from_date < ? AND id != ?
        ORDER BY from_date DESC LIMIT 1
    `, toDeleteEmployeeHistory.EmployeeID, currentFromDateFormatted, toDeleteEmployeeHistory.ID,
	).Scan(&previous.ID, &previous.FromDate, &previous.ToDate); err != nil && err != sql.ErrNoRows {
		return err
	}
	var next models.EmployeeHistory
	if err := tx.QueryRow(`
	   SELECT id, from_date, to_date FROM employee_history
       WHERE employee_id = ? AND from_date > ? AND id != ?
	   ORDER BY from_date ASC LIMIT 1
	`, toDeleteEmployeeHistory.EmployeeID, currentFromDateFormatted, toDeleteEmployeeHistory.ID,
	).Scan(&next.ID, &next.FromDate, &next.ToDate); err != nil && err != sql.ErrNoRows {
		return err
	}

	if previous.ID != 0 && next.ID != 0 {
		previousFromDate := time.Time(previous.FromDate)
		nextFromDate := time.Time(next.FromDate)
		newToDate := utils.GetNextAvailableDate(previousFromDate, nextFromDate)
		logger.Logger.Debugf("Adjusting previous history entry to: %s", newToDate)
		_, err := tx.Exec(`
            UPDATE employee_history SET to_date = ? WHERE id = ?
        `, newToDate, previous.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

// ListForecasts lists all forecasts
func (s *DatabaseService) ListForecasts(userID int64, limit int64) ([]models.Forecast, error) {
	forecasts := make([]models.Forecast, 0)

	query, err := sqlQueries.ReadFile("queries/list_forecasts.sql")
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(string(query), limit, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var forecast models.Forecast

		err := rows.Scan(
			&forecast.Data.Month, &forecast.Data.Revenue, &forecast.Data.Expense, &forecast.Data.Cashflow,
			&forecast.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		forecasts = append(forecasts, forecast)
	}

	return forecasts, nil
}

// ListForecastDetails lists all forecasts
func (s *DatabaseService) ListForecastDetails(userID int64, limit int64) ([]models.ForecastDatabaseDetails, error) {
	forecastDetails := make([]models.ForecastDatabaseDetails, 0)

	query, err := sqlQueries.ReadFile("queries/list_forecast_details.sql")
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(string(query), limit, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var forecastDetail models.ForecastDatabaseDetails
		var revenueJSON, expenseJSON []byte

		if err := rows.Scan(&forecastDetail.Month, &revenueJSON, &expenseJSON, &forecastDetail.ForecastID); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(revenueJSON, &forecastDetail.Revenue); err != nil {
			log.Fatal(err)
		}
		if err := json.Unmarshal(expenseJSON, &forecastDetail.Expense); err != nil {
			log.Fatal(err)
		}

		forecastDetails = append(forecastDetails, forecastDetail)
	}

	return forecastDetails, nil
}

// UpsertForecast Inserts or updates a forecast
func (s *DatabaseService) UpsertForecast(payload models.CreateForecast, userID int64) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/upsert_forecast.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		payload.Month, payload.Revenue, payload.Expense, payload.Cashflow, userID,
	)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted employee
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// UpsertForecastDetail Inserts or updates a forecast detail that belongs to a forecast
func (s *DatabaseService) UpsertForecastDetail(payload models.CreateForecastDetail, userID, forecastID int64) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/upsert_forecast_detail.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	revenueJson, err := json.Marshal(payload.Revenue)
	if err != nil {
		return 0, err
	}

	expenseJson, err := json.Marshal(payload.Expense)
	if err != nil {
		return 0, err
	}

	res, err := stmt.Exec(
		payload.Month, revenueJson, expenseJson, forecastID, userID,
	)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted employee
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// ClearForecasts clears all records of the user
func (s *DatabaseService) ClearForecasts(userID int64) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/clear_forecasts.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(userID)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted employee
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}

// ListBankAccounts returns the available bank accounts
func (s *DatabaseService) ListBankAccounts(userID int64) ([]models.BankAccount, error) {
	bankAccounts := make([]models.BankAccount, 0)

	query, err := sqlQueries.ReadFile("queries/list_bank_accounts.sql")
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(string(query), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bankAccount models.BankAccount

		err := rows.Scan(
			&bankAccount.ID, &bankAccount.Name, &bankAccount.Amount,
			&bankAccount.Currency.ID, &bankAccount.Currency.Code, &bankAccount.Currency.Description, &bankAccount.Currency.LocaleCode,
		)
		if err != nil {
			return nil, err
		}

		bankAccounts = append(bankAccounts, bankAccount)
	}

	return bankAccounts, nil
}

func (s *DatabaseService) GetBankAccount(userID int64, bankAccountID string) (*models.BankAccount, error) {
	var bankAccount models.BankAccount

	query, err := sqlQueries.ReadFile("queries/get_bank_account.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), bankAccountID, userID).Scan(
		&bankAccount.ID, &bankAccount.Name, &bankAccount.Amount,
		&bankAccount.Currency.ID, &bankAccount.Currency.Code, &bankAccount.Currency.Description, &bankAccount.Currency.LocaleCode,
	)
	if err != nil {
		return nil, err
	}

	return &bankAccount, nil
}

func (s *DatabaseService) CreateBankAccount(payload models.CreateBankAccount, userID int64) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_bank_account.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		payload.Name, payload.Amount, payload.Currency, userID,
	)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted bank account
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *DatabaseService) UpdateBankAccount(payload models.UpdateBankAccount, userID int64, bankAccountID string) error {
	// Base query
	query := "UPDATE bank_accounts SET "
	queryBuild := []string{}
	args := []interface{}{}

	// Dynamically add fields that are not nil
	if payload.Name != nil {
		queryBuild = append(queryBuild, "name = ?")
		args = append(args, *payload.Name)
	}
	if payload.Amount != nil {
		queryBuild = append(queryBuild, "amount = ?")
		args = append(args, *payload.Amount)
	}
	if payload.Currency != nil {
		queryBuild = append(queryBuild, "currency_id = ?")
		args = append(args, *payload.Currency)
	}

	// Add WHERE clause
	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ? AND organisation_id = (SELECT current_organisation FROM users u WHERE u.id = ?)"
	args = append(args, bankAccountID, userID)

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(args...)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) DeleteBankAccount(userID int64, bankAccountID string) error {
	query, err := sqlQueries.ReadFile("queries/delete_bank_account.sql")
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(bankAccountID, userID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

// ListVats returns generally available vats and the vats created by the user
func (s *DatabaseService) ListVats(userID int64) ([]models.Vat, error) {
	vats := make([]models.Vat, 0)

	query, err := sqlQueries.ReadFile("queries/list_vats.sql")
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(string(query), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var vat models.Vat

		err := rows.Scan(
			&vat.ID, &vat.Value, &vat.FormattedValue, &vat.CanEdit,
		)
		if err != nil {
			return nil, err
		}

		vats = append(vats, vat)
	}

	return vats, nil
}

func (s *DatabaseService) GetVat(userID int64, vatID string) (*models.Vat, error) {
	var vat models.Vat

	query, err := sqlQueries.ReadFile("queries/get_vat.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), vatID, userID).Scan(
		&vat.ID, &vat.Value, &vat.FormattedValue, &vat.CanEdit,
	)
	if err != nil {
		return nil, err
	}

	return &vat, nil
}

func (s *DatabaseService) CreateVat(payload models.CreateVat, userID int64) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_vat.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		payload.Value, userID,
	)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted bank account
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *DatabaseService) UpdateVat(payload models.UpdateVat, userID int64, vatID string) error {
	// Base query
	query := "UPDATE vats SET "
	queryBuild := []string{}
	args := []interface{}{}

	// Dynamically add fields that are not nil
	if payload.Value != nil {
		queryBuild = append(queryBuild, "value = ?")
		args = append(args, *payload.Value)
	}

	// Add WHERE clause
	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ? AND organisation_id = (SELECT current_organisation FROM users u WHERE u.id = ?)"
	args = append(args, vatID, userID)

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(args...)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) DeleteVat(userID int64, vatID string) error {
	query, err := sqlQueries.ReadFile("queries/delete_vat.sql")
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(vatID, userID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

// ListCategories implements listing categories
func (s *DatabaseService) ListCategories(userID, page, limit int64) ([]models.Category, int64, error) {
	categories := []models.Category{}
	var totalCount int64

	query, err := sqlQueries.ReadFile("queries/list_categories.sql")
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(string(query), userID, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var category models.Category

		err := rows.Scan(&category.ID, &category.Name, &category.CanEdit, &totalCount)
		if err != nil {
			return nil, 0, err
		}

		categories = append(categories, category)
	}

	return categories, totalCount, nil
}

// GetCategory implements fetching a category by ID
func (s *DatabaseService) GetCategory(userID int64, id string) (*models.Category, error) {
	var category models.Category

	query, err := sqlQueries.ReadFile("queries/get_category.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), id, userID).Scan(&category.ID, &category.Name, &category.CanEdit)
	if err != nil {
		return nil, err
	}

	return &category, nil
}

// CreateCategory implements the creation of a new category
func (s *DatabaseService) CreateCategory(userID *int64, payload models.CreateCategory) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_category.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(payload.Name, userID)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// UpdateCategory implements updating a category
func (s *DatabaseService) UpdateCategory(userID int64, payload models.UpdateCategory, categoryID string) error {
	query := "UPDATE categories SET "
	queryBuild := []string{}
	args := []interface{}{}

	if payload.Name != nil {
		queryBuild = append(queryBuild, "name = ?")
		args = append(args, *payload.Name)
	}

	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ? AND organisation_id = (SELECT current_organisation FROM users u WHERE u.id = ?)"
	args = append(args, categoryID, userID)

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}

	return nil
}

// ListCurrencies implements listing currencies
func (s *DatabaseService) ListCurrencies(page int64, limit int64) ([]models.Currency, int64, error) {
	currencies := []models.Currency{}
	var totalCount int64

	query, err := sqlQueries.ReadFile("queries/list_currencies.sql")
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(string(query), limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var currency models.Currency

		err := rows.Scan(&currency.ID, &currency.Code, &currency.Description, &currency.LocaleCode, &totalCount)
		if err != nil {
			return nil, 0, err
		}

		currencies = append(currencies, currency)
	}

	return currencies, totalCount, nil
}

// GetCurrency implements fetching a currency by ID
func (s *DatabaseService) GetCurrency(id string) (*models.Currency, error) {
	var currency models.Currency

	query, err := sqlQueries.ReadFile("queries/get_currency.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), id).Scan(
		&currency.ID, &currency.Code, &currency.Description, &currency.LocaleCode,
	)
	if err != nil {
		return nil, err
	}

	return &currency, nil
}

// CreateCurrency implements the creation of a new currency
func (s *DatabaseService) CreateCurrency(payload models.CreateCurrency) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_currency.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(payload.Code, payload.Description, payload.LocaleCode)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// UpdateCurrency implements updating a currency
func (s *DatabaseService) UpdateCurrency(payload models.UpdateCurrency, currencyID string) error {
	query := "UPDATE currencies SET "
	queryBuild := []string{}
	args := []interface{}{}

	if payload.Code != nil {
		queryBuild = append(queryBuild, "code = ?")
		args = append(args, *payload.Code)
	}
	if payload.Description != nil {
		queryBuild = append(queryBuild, "description = ?")
		args = append(args, *payload.Description)
	}
	if payload.LocaleCode != nil {
		queryBuild = append(queryBuild, "locale_code = ?")
		args = append(args, *payload.LocaleCode)
	}

	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ?"
	args = append(args, currencyID)

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}

	return nil
}

// StoreRefreshTokenID stores the refresh token's token ID, user ID, device name and expiration time in the database
func (s *DatabaseService) StoreRefreshTokenID(userID int64, tokenId string, expirationTime time.Time, deviceName string) error {
	query, err := sqlQueries.ReadFile("queries/create_refresh_token.sql")
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, tokenId, expirationTime, deviceName)
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) CheckRefreshToken(tokenID string, userID int64) (bool, error) {
	query, err := sqlQueries.ReadFile("queries/get_refresh_token.sql")
	if err != nil {
		return false, err
	}

	var exists bool
	err = s.db.QueryRow(string(query), tokenID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *DatabaseService) DeleteRefreshToken(tokenID string, userID int64) error {
	query, err := sqlQueries.ReadFile("queries/delete_refresh_token.sql")
	if err != nil {
		return err
	}

	_, err = s.db.Exec(string(query), tokenID, userID)

	return err
}

func (s *DatabaseService) ListFiatRates(base string) ([]models.FiatRate, error) {
	fiatRates := []models.FiatRate{}

	query, err := sqlQueries.ReadFile("queries/list_fiat_rates.sql")
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(string(query), base)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var fiatRate models.FiatRate

		err := rows.Scan(&fiatRate.ID, &fiatRate.Base, &fiatRate.Target, &fiatRate.Rate, &fiatRate.UpdatedAt)
		if err != nil {
			return nil, err
		}

		fiatRates = append(fiatRates, fiatRate)
	}

	return fiatRates, nil
}

func (s *DatabaseService) GetFiatRate(base, target string) (*models.FiatRate, error) {
	var fiatRate models.FiatRate

	query, err := sqlQueries.ReadFile("queries/get_fiat_rate.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), base, target).Scan(
		&fiatRate.ID, &fiatRate.Base, &fiatRate.Target, &fiatRate.Rate, &fiatRate.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &fiatRate, nil
}

// UpsertFiatRate Inserts or updates a fiat rate
func (s *DatabaseService) UpsertFiatRate(payload models.CreateFiatRate) error {
	query, err := sqlQueries.ReadFile("queries/upsert_fiat_rate.sql")
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		payload.Base, payload.Target, payload.Rate,
	)
	if err != nil {
		return err
	}

	return nil
}
