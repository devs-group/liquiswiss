//go:generate mockgen -package=mocks -destination ../mocks/db_service.go liquiswiss/internal/service IDatabaseService
package service

import (
	"database/sql"
	"embed"
	"io/fs"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/types"
	"path/filepath"
	"strings"
	"time"
)

//go:embed queries/*.sql
var sqlQueries embed.FS

//go:embed mocks/*.sql
var sqlMocks embed.FS

type IDatabaseService interface {
	ApplyMocks() error

	RegisterUser(email, password string) (int64, error)
	GetUserPasswordByEMail(email string) (*models.Login, error)
	GetProfile(id string) (*models.User, error)
	CheckUserExistance(id int64) (bool, error)

	ListTransactions(userID int64, page int64, limit int64) ([]models.Transaction, int64, error)
	GetTransaction(userID int64, id string) (*models.Transaction, error)
	CreateTransaction(payload models.CreateTransaction, userID int64) (int64, error)
	UpdateTransaction(payload models.UpdateTransaction, userID int64, transactionID string) error

	ListOrganisations(userID int64, page int64, limit int64) ([]models.Organisation, int64, error)
	GetOrganisation(userID int64, id string) (*models.Organisation, error)
	CreateOrganisation(name string, userID int64) (int64, error)
	UpdateOrganisation(payload models.UpdateOrganisation, userID int64, organisationID string) error
	AssignUserToOrganisation(userID int64, organisationID int64, role string) error

	ListEmployees(userID int64, page int64, limit int64) ([]models.Employee, int64, error)
	CountEmployees(userID int64, page int64, limit int64) (int64, error)
	GetEmployee(userID int64, id string) (*models.Employee, error)
	CreateEmployee(payload models.CreateEmployee, userID int64) (int64, error)
	UpdateEmployee(payload models.UpdateEmployee, userID int64, employeeID string) error
	DeleteEmployee(employeeID int64, userID int64) error

	ListCategories(page int64, limit int64) ([]models.Category, int64, error)
	GetCategory(id string) (*models.Category, error)
	CreateCategory(payload models.CreateCategory) (int64, error)
	UpdateCategory(payload models.UpdateCategory, categoryID string) error

	ListCurrencies(page int64, limit int64) ([]models.Currency, int64, error)
	GetCurrency(id string) (*models.Currency, error)
	CreateCurrency(payload models.CreateCurrency) (int64, error)
	UpdateCurrency(payload models.UpdateCurrency, currencyID string) error

	StoreRefreshTokenID(userID int64, tokenId string, expirationTime time.Time, deviceName string) error
	CheckRefreshToken(tokenID string, userID int64) (bool, error)
	DeleteRefreshToken(tokenID string, userID int64) error
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
		if filepath.Ext(path) == ".sql" {
			query, err := sqlMocks.ReadFile(path)
			if err != nil {
				logger.Logger.Errorf("Could not read %s: %v", path, err)
			}
			_, err = s.db.Exec(string(query))
			if err != nil {
				logger.Logger.Warnf("Failed to apply %s to DB: %s", path, err)
			} else {
				logger.Logger.Infof("Applied %s to DB", path)
			}
		}
		return nil
	})
}

func (s *DatabaseService) RegisterUser(email string, password string) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/register_user.sql")
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

	err = s.db.QueryRow(string(query), id).Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		return nil, err
	}

	return &user, nil
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
		payload.Category, payload.Currency, userID, nil,
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
	query := "UPDATE go_transactions SET "
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
		args = append(args, *payload.EndDate)
	} else {
		args = append(args, nil)
	}
	if payload.Category != nil {
		queryBuild = append(queryBuild, "category = ?")
		args = append(args, *payload.Category)
	}
	if payload.Currency != nil {
		queryBuild = append(queryBuild, "currency = ?")
		args = append(args, *payload.Currency)
	}

	// Add WHERE clause
	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ?"
	args = append(args, transactionID)

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

func (s *DatabaseService) ListTransactions(userID int64, page int64, limit int64) ([]models.Transaction, int64, error) {
	transactions := []models.Transaction{}
	var totalCount int64

	query, err := sqlQueries.ReadFile("queries/list_transactions.sql")
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(string(query), userID, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var transaction models.Transaction
		// These are required for proper date convertion afterwards
		var startDate time.Time
		var endDate sql.NullTime

		err := rows.Scan(
			&transaction.ID, &transaction.Name, &transaction.Amount, &transaction.Cycle, &transaction.Type, &startDate, &endDate,
			&transaction.Category.ID, &transaction.Category.Name,
			&transaction.Currency.ID, &transaction.Currency.Code, &transaction.Currency.Description, &transaction.Currency.LocaleCode,
			&totalCount,
		)
		if err != nil {
			return nil, 0, err
		}

		transaction.StartDate = types.AsDate(startDate)

		if endDate.Valid {
			convertedDate := types.AsDate(endDate.Time)
			transaction.EndDate = &convertedDate
		}

		transactions = append(transactions, transaction)
	}

	return transactions, totalCount, nil
}

func (s *DatabaseService) GetTransaction(userID int64, id string) (*models.Transaction, error) {
	var transaction models.Transaction
	// These are required for proper date convertion afterwards
	var startDate time.Time
	var endDate sql.NullTime

	query, err := sqlQueries.ReadFile("queries/get_transaction.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), id, userID).Scan(
		&transaction.ID, &transaction.Name, &transaction.Amount, &transaction.Cycle, &transaction.Type, &startDate, &endDate,
		&transaction.Category.ID, &transaction.Category.Name,
		&transaction.Currency.ID, &transaction.Currency.Code, &transaction.Currency.Description, &transaction.Currency.LocaleCode,
	)
	if err != nil {
		return nil, err
	}

	transaction.StartDate = types.AsDate(startDate)

	if endDate.Valid {
		convertedDate := types.AsDate(endDate.Time)
		transaction.EndDate = &convertedDate
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
			&organisation.ID, &organisation.Name, &organisation.MemberCount, &organisation.Role,
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
func (s *DatabaseService) CreateOrganisation(name string, userID int64) (int64, error) {
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
	query := "UPDATE go_organisations SET "
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

func (s *DatabaseService) AssignUserToOrganisation(userID int64, organisationID int64, role string) error {
	query, err := sqlQueries.ReadFile("queries/assign_user_to_organisation.sql")
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, organisationID, role)
	if err != nil {
		return err
	}

	return nil
}

// ListEmployees implements employee listing with pagination
func (s *DatabaseService) ListEmployees(userID int64, page int64, limit int64) ([]models.Employee, int64, error) {
	employees := make([]models.Employee, 0)
	var totalCount int64

	query, err := sqlQueries.ReadFile("queries/list_employees.sql")
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(string(query), userID, (page)*limit, 0)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var employee models.Employee
		var entryDate time.Time
		var exitDate sql.NullTime

		err := rows.Scan(&employee.ID, &employee.Name, &employee.HoursPerMonth, &employee.VacationDaysPerYear, &entryDate, &exitDate, &totalCount)
		if err != nil {
			return nil, 0, err
		}

		employee.EntryDate = types.AsDate(entryDate)

		if exitDate.Valid {
			convertedDate := types.AsDate(exitDate.Time)
			employee.ExitDate = &convertedDate
		}

		employees = append(employees, employee)
	}

	return employees, totalCount, nil
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
	var entryDate time.Time
	var exitDate sql.NullTime

	query, err := sqlQueries.ReadFile("queries/get_employee.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), id, userID).Scan(
		&employee.ID, &employee.Name, &employee.HoursPerMonth, &employee.VacationDaysPerYear,
		&entryDate, &exitDate,
	)
	if err != nil {
		return nil, err
	}

	employee.EntryDate = types.AsDate(entryDate)

	if exitDate.Valid {
		convertedDate := types.AsDate(exitDate.Time)
		employee.ExitDate = &convertedDate
	}

	return &employee, nil
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

	// Prepare entry and exit date
	entryDate, err := time.Parse("2006-01-02", payload.EntryDate)
	if err != nil {
		return 0, err
	}

	var exitDate sql.NullTime
	if payload.ExitDate != nil {
		parsedExitDate, err := time.Parse("2006-01-02", *payload.ExitDate)
		if err != nil {
			return 0, err
		}
		exitDate = sql.NullTime{Time: parsedExitDate, Valid: true}
	} else {
		exitDate = sql.NullTime{Valid: false}
	}

	res, err := stmt.Exec(
		payload.Name, payload.HoursPerMonth, payload.VacationDaysPerYear, entryDate, exitDate, userID,
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

// UpdateEmployee implements updating employee details
func (s *DatabaseService) UpdateEmployee(payload models.UpdateEmployee, userID int64, employeeID string) error {
	// Base query
	query := "UPDATE go_employees SET "
	queryBuild := []string{}
	args := []interface{}{}

	// Dynamically add fields that are not nil
	if payload.Name != nil {
		queryBuild = append(queryBuild, "name = ?")
		args = append(args, *payload.Name)
	}
	if payload.HoursPerMonth != nil {
		queryBuild = append(queryBuild, "hours_per_month = ?")
		args = append(args, *payload.HoursPerMonth)
	}
	if payload.VacationDaysPerYear != nil {
		queryBuild = append(queryBuild, "vacation_days_per_year = ?")
		args = append(args, *payload.VacationDaysPerYear)
	}
	if payload.EntryDate != nil {
		queryBuild = append(queryBuild, "entry_date = ?")
		entryDate, err := time.Parse("2006-01-02", *payload.EntryDate)
		if err != nil {
			return err
		}
		args = append(args, entryDate)
	}
	// Always consider ExitDate in case it is set back to null
	queryBuild = append(queryBuild, "exit_date = ?")
	if payload.ExitDate != nil {
		exitDate, err := time.Parse("2006-01-02", *payload.ExitDate)
		if err != nil {
			return err
		}
		args = append(args, exitDate)
	} else {
		args = append(args, nil)
	}

	// Add WHERE clause
	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ?"
	args = append(args, employeeID)

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

// ListCategories implements listing categories
func (s *DatabaseService) ListCategories(page int64, limit int64) ([]models.Category, int64, error) {
	categories := []models.Category{}
	var totalCount int64

	query, err := sqlQueries.ReadFile("queries/list_categories.sql")
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(string(query), limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var category models.Category

		err := rows.Scan(&category.ID, &category.Name, &totalCount)
		if err != nil {
			return nil, 0, err
		}

		categories = append(categories, category)
	}

	return categories, totalCount, nil
}

// GetCategory implements fetching a category by ID
func (s *DatabaseService) GetCategory(id string) (*models.Category, error) {
	var category models.Category

	query, err := sqlQueries.ReadFile("queries/get_category.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), id).Scan(&category.ID, &category.Name)
	if err != nil {
		return nil, err
	}

	return &category, nil
}

// CreateCategory implements the creation of a new category
func (s *DatabaseService) CreateCategory(payload models.CreateCategory) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_category.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(payload.Name)
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
func (s *DatabaseService) UpdateCategory(payload models.UpdateCategory, categoryID string) error {
	query := "UPDATE go_categories SET "
	queryBuild := []string{}
	args := []interface{}{}

	if payload.Name != nil {
		queryBuild = append(queryBuild, "name = ?")
		args = append(args, *payload.Name)
	}

	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ?"
	args = append(args, categoryID)

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
	query := "UPDATE go_currencies SET "
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
