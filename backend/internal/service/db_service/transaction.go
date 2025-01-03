package db_service

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/types"
	"liquiswiss/pkg/utils"
	"strings"
	"time"
)

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
			&transaction.ID,
			&transaction.Name,
			&transaction.Amount,
			&transaction.VatAmount,
			&transaction.VatIncluded,
			&transaction.Cycle,
			&transaction.Type,
			&startDate,
			&endDate,
			&transaction.Category.ID,
			&transaction.Category.Name,
			&transaction.Currency.ID,
			&transaction.Currency.Code,
			&transaction.Currency.Description,
			&transaction.Currency.LocaleCode,
			&transactionEmployeeID,
			&transactionEmployeeName,
			&vatID,
			&vatValue,
			&vatFormattedValue,
			&vatCanEdit,
			&nextExecutionDate,
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

func (s *DatabaseService) GetTransaction(userID int64, transactionID int64) (*models.Transaction, error) {
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
	var nextExecutionDate sql.NullTime

	query, err := sqlQueries.ReadFile("queries/get_transaction.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), transactionID, userID).Scan(
		&transaction.ID,
		&transaction.Name,
		&transaction.Amount,
		&transaction.VatAmount,
		&transaction.VatIncluded,
		&transaction.Cycle,
		&transaction.Type,
		&startDate,
		&endDate,
		&transaction.Category.ID,
		&transaction.Category.Name,
		&transaction.Currency.ID,
		&transaction.Currency.Code,
		&transaction.Currency.Description,
		&transaction.Currency.LocaleCode,
		&transactionEmployeeID,
		&transactionEmployeeName,
		&vatID,
		&vatValue,
		&vatFormattedValue,
		&vatCanEdit,
		&nextExecutionDate,
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

	if nextExecutionDate.Valid {
		convertedDate := types.AsDate(nextExecutionDate.Time)
		transaction.NextExecutionDate = &convertedDate
	}

	return &transaction, nil
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

func (s *DatabaseService) UpdateTransaction(payload models.UpdateTransaction, userID int64, transactionID int64) error {
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
	query += " WHERE id = ? AND organisation_id = get_current_organisation(?)"
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

func (s *DatabaseService) DeleteTransaction(userID int64, transactionID int64) error {
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
