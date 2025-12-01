package db_adapter

import (
	"database/sql"
	"liquiswiss/pkg/models"
	"strings"
)

func (d *DatabaseAdapter) GetVatSetting(userID int64) (*models.VatSetting, error) {
	var vatSetting models.VatSetting

	query, err := sqlQueries.ReadFile("queries/get_vat_setting.sql")
	if err != nil {
		return nil, err
	}

	err = d.db.QueryRow(string(query), userID).Scan(
		&vatSetting.ID,
		&vatSetting.OrganisationID,
		&vatSetting.Enabled,
		&vatSetting.BillingDate,
		&vatSetting.TransactionMonthOffset,
		&vatSetting.Interval,
		&vatSetting.CreatedAt,
		&vatSetting.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &vatSetting, nil
}

func (d *DatabaseAdapter) CreateVatSetting(payload models.CreateVatSetting, userID int64) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_vat_setting.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		userID,
		payload.Enabled,
		payload.BillingDate,
		payload.TransactionMonthOffset,
		payload.Interval,
	)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted vat setting
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (d *DatabaseAdapter) UpdateVatSetting(payload models.UpdateVatSetting, userID int64) error {
	// Base query
	query := "UPDATE vat_settings SET "
	queryBuild := []string{}
	args := []interface{}{}

	// Dynamically add fields that are not nil
	if payload.Enabled != nil {
		queryBuild = append(queryBuild, "enabled = ?")
		args = append(args, *payload.Enabled)
	}

	if payload.BillingDate != nil {
		queryBuild = append(queryBuild, "billing_date = ?")
		args = append(args, *payload.BillingDate)
	}

	if payload.TransactionMonthOffset != nil {
		queryBuild = append(queryBuild, "transaction_month_offset = ?")
		args = append(args, *payload.TransactionMonthOffset)
	}

	if payload.Interval != nil {
		queryBuild = append(queryBuild, "`interval` = ?")
		args = append(args, *payload.Interval)
	}

	// Add WHERE clause
	query += strings.Join(queryBuild, ", ")
	query += " WHERE organisation_id = get_current_user_organisation_id(?)"
	args = append(args, userID)

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

func (d *DatabaseAdapter) DeleteVatSetting(userID int64) error {
	query, err := sqlQueries.ReadFile("queries/delete_vat_setting.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(userID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}
