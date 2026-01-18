package db_adapter

import (
	"liquiswiss/pkg/models"
	"strings"
)

func (d *DatabaseAdapter) ListBankAccounts(userID int64) ([]models.BankAccount, error) {
	bankAccounts := make([]models.BankAccount, 0)

	query, err := sqlQueries.ReadFile("queries/list_bank_accounts.sql")
	if err != nil {
		return nil, err
	}

	rows, err := d.db.Query(string(query), userID)
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

func (d *DatabaseAdapter) GetBankAccount(userID int64, bankAccountID int64) (*models.BankAccount, error) {
	var bankAccount models.BankAccount

	query, err := sqlQueries.ReadFile("queries/get_bank_account.sql")
	if err != nil {
		return nil, err
	}

	err = d.db.QueryRow(string(query), bankAccountID, userID).Scan(
		&bankAccount.ID, &bankAccount.Name, &bankAccount.Amount,
		&bankAccount.Currency.ID, &bankAccount.Currency.Code, &bankAccount.Currency.Description, &bankAccount.Currency.LocaleCode,
	)
	if err != nil {
		return nil, err
	}

	return &bankAccount, nil
}

func (d *DatabaseAdapter) CreateBankAccount(payload models.CreateBankAccount, userID int64) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_bank_account.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := d.db.Prepare(string(query))
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

func (d *DatabaseAdapter) UpdateBankAccount(payload models.UpdateBankAccount, userID int64, bankAccountID int64) error {
	// Base query
	query := "UPDATE bank_accounts SET "
	queryBuild := []string{}
	args := []any{}

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
	query += " WHERE id = ? AND organisation_id = get_current_user_organisation_id(?)"
	args = append(args, bankAccountID, userID)

	stmt, err := d.db.Prepare(query)
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

func (d *DatabaseAdapter) DeleteBankAccount(userID int64, bankAccountID int64) error {
	query, err := sqlQueries.ReadFile("queries/delete_bank_account.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
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
