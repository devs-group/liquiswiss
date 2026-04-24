package db_adapter

import (
	"database/sql"
	"liquiswiss/pkg/models"
)

func (d *DatabaseAdapter) ListBankAccountsByOrganisation(orgID int64) ([]models.BankAccount, error) {
	query, err := sqlQueries.ReadFile("queries/list_bank_accounts_by_organisation.sql")
	if err != nil {
		return nil, err
	}

	rows, err := d.db.Query(string(query), orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.BankAccount
	for rows.Next() {
		var ba models.BankAccount
		if err := rows.Scan(
			&ba.ID, &ba.Name, &ba.Amount,
			&ba.Currency.ID, &ba.Currency.Code, &ba.Currency.Description, &ba.Currency.LocaleCode,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, ba)
	}
	return accounts, rows.Err()
}

func (d *DatabaseAdapter) ListEmployeesByOrganisation(orgID int64) ([]models.Employee, error) {
	query, err := sqlQueries.ReadFile("queries/list_employees_by_organisation.sql")
	if err != nil {
		return nil, err
	}

	rows, err := d.db.Query(string(query), orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []models.Employee
	for rows.Next() {
		var e models.Employee
		if err := rows.Scan(&e.ID, &e.Name); err != nil {
			return nil, err
		}
		employees = append(employees, e)
	}
	return employees, rows.Err()
}

func (d *DatabaseAdapter) ListTransactionsByOrganisation(orgID int64) ([]models.Transaction, error) {
	query, err := sqlQueries.ReadFile("queries/list_transactions_by_organisation.sql")
	if err != nil {
		return nil, err
	}

	rows, err := d.db.Query(string(query), orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		var currencyCode string
		if err := rows.Scan(
			&t.ID, &t.Name, &t.Amount, &t.Cycle, &t.Type, &t.IsDisabled, &currencyCode,
		); err != nil {
			return nil, err
		}
		t.Currency.Code = &currencyCode
		transactions = append(transactions, t)
	}
	return transactions, rows.Err()
}

func (d *DatabaseAdapter) GetOrganisationByID(orgID int64) (*models.Organisation, error) {
	query, err := sqlQueries.ReadFile("queries/get_organisation_by_id.sql")
	if err != nil {
		return nil, err
	}

	var org models.Organisation
	var currencyID sql.NullInt64
	var currencyCode, currencyDesc, currencyLocale sql.NullString

	err = d.db.QueryRow(string(query), orgID).Scan(
		&org.ID, &org.Name,
		&currencyID, &currencyCode, &currencyDesc, &currencyLocale,
		&org.MemberCount,
	)
	if err != nil {
		return nil, err
	}

	if currencyID.Valid {
		id := currencyID.Int64
		org.Currency.ID = &id
		org.Currency.Code = &currencyCode.String
		org.Currency.Description = &currencyDesc.String
		org.Currency.LocaleCode = &currencyLocale.String
	}

	return &org, nil
}
