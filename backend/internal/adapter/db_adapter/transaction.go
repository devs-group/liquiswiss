package db_adapter

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

func (d *DatabaseAdapter) ListTransactions(userID int64, page int64, limit int64, sortBy string, sortOrder string, search string) ([]models.Transaction, int64, error) {
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
	err = parsed.Execute(&query, map[string]any{
		"sortBy":    sortBy,
		"sortOrder": sortOrder,
		"hasSearch": search != "",
	})
	if err != nil {
		return nil, 0, err
	}

	var rows *sql.Rows
	if search != "" {
		searchPattern := "%" + search + "%"
		rows, err = d.db.Query(query.String(), userID, searchPattern, (page)*limit, 0)
	} else {
		rows, err = d.db.Query(query.String(), userID, (page)*limit, 0)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var transactionID int64

		err := rows.Scan(
			&transactionID,
			&totalCount,
		)
		if err != nil {
			return nil, 0, err
		}

		transaction, err := d.GetTransaction(userID, transactionID)
		if err != nil {
			return nil, 0, err
		}

		transactions = append(transactions, *transaction)
	}

	return transactions, totalCount, nil
}

func (d *DatabaseAdapter) GetTransaction(userID int64, transactionID int64) (*models.Transaction, error) {
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

	err = d.db.QueryRow(string(query), transactionID, userID).Scan(
		&transaction.ID,
		&transaction.Name,
		&transaction.Link,
		&transaction.Amount,
		&transaction.VatAmount,
		&transaction.VatIncluded,
		&transaction.IsDisabled,
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
		&transaction.DBDate,
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

	nextExecutionDate := d.CalculateSalaryExecutionDate(transaction.StartDate, transaction.EndDate, transaction.Cycle, transaction.DBDate, 1, true)
	if nextExecutionDate != nil {
		nextExecutionDateAsDate := types.AsDate(*nextExecutionDate)
		transaction.NextExecutionDate = &nextExecutionDateAsDate
	}

	return &transaction, nil
}

func (d *DatabaseAdapter) CreateTransaction(payload models.CreateTransaction, userID int64) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_transaction.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		payload.Name, payload.Link, payload.Amount, payload.Cycle, payload.Type, payload.StartDate, payload.EndDate,
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

func (d *DatabaseAdapter) UpdateTransaction(payload models.UpdateTransaction, userID int64, transactionID int64) error {
	// Base query
	query := "UPDATE transactions SET "
	queryBuild := []string{}
	args := []any{}

	// Dynamically add fields that are not nil
	if payload.Name != nil {
		queryBuild = append(queryBuild, "name = ?")
		args = append(args, *payload.Name)
	}
	if payload.Link != nil {
		queryBuild = append(queryBuild, "link = ?")
		args = append(args, *payload.Link)
	} else if payload.IsDisabled == nil {
		queryBuild = append(queryBuild, "link = ?")
		args = append(args, nil)
	}
	if payload.Amount != nil {
		queryBuild = append(queryBuild, "amount = ?")
		args = append(args, *payload.Amount)
	}
	if payload.Cycle != nil {
		queryBuild = append(queryBuild, "cycle = ?")
		args = append(args, *payload.Cycle)
	} else if payload.IsDisabled == nil {
		queryBuild = append(queryBuild, "cycle = ?")
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
	if payload.EndDate != nil {
		queryBuild = append(queryBuild, "end_date = ?")
		endDate, err := time.Parse(utils.InternalDateFormat, *payload.EndDate)
		if err != nil {
			return err
		}
		args = append(args, endDate)
	} else if payload.IsDisabled == nil {
		queryBuild = append(queryBuild, "end_date = ?")
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
	if payload.Employee != nil {
		queryBuild = append(queryBuild, "employee_id = ?")
		args = append(args, *payload.Employee)
	} else if payload.IsDisabled == nil {
		queryBuild = append(queryBuild, "employee_id = ?")
		args = append(args, nil)
	}
	if payload.Vat != nil {
		queryBuild = append(queryBuild, "vat_id = ?")
		args = append(args, *payload.Vat)
	} else if payload.IsDisabled == nil {
		queryBuild = append(queryBuild, "vat_id = ?")
		args = append(args, nil)
	}
	if payload.VatIncluded != nil {
		queryBuild = append(queryBuild, "vat_included = ?")
		args = append(args, *payload.VatIncluded)
	}
	if payload.IsDisabled != nil {
		queryBuild = append(queryBuild, "is_disabled = ?")
		args = append(args, *payload.IsDisabled)
	}

	// Add WHERE clause
	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ? AND organisation_id = get_current_user_organisation_id(?)"
	args = append(args, transactionID)
	args = append(args, userID)

	stmt, err := d.db.Prepare(query)
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

func (d *DatabaseAdapter) DeleteTransaction(userID int64, transactionID int64) error {
	query, err := sqlQueries.ReadFile("queries/delete_transaction.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
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
