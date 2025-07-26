package db_adapter

import (
	"liquiswiss/pkg/models"
	"strings"
)

func (d *DatabaseAdapter) ListVats(userID int64) ([]models.Vat, error) {
	vats := make([]models.Vat, 0)

	query, err := sqlQueries.ReadFile("queries/list_vats.sql")
	if err != nil {
		return nil, err
	}

	rows, err := d.db.Query(string(query), userID)
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

func (d *DatabaseAdapter) GetVat(userID int64, vatID int64) (*models.Vat, error) {
	var vat models.Vat

	query, err := sqlQueries.ReadFile("queries/get_vat.sql")
	if err != nil {
		return nil, err
	}

	err = d.db.QueryRow(string(query), vatID, userID).Scan(
		&vat.ID, &vat.Value, &vat.FormattedValue, &vat.CanEdit,
	)
	if err != nil {
		return nil, err
	}

	return &vat, nil
}

func (d *DatabaseAdapter) CreateVat(payload models.CreateVat, userID int64) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_vat.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := d.db.Prepare(string(query))
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

func (d *DatabaseAdapter) UpdateVat(payload models.UpdateVat, userID int64, vatID int64) error {
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
	query += " WHERE id = ? AND organisation_id = get_current_user_organisation_id(?)"
	args = append(args, vatID, userID)

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

func (d *DatabaseAdapter) DeleteVat(userID int64, vatID int64) error {
	query, err := sqlQueries.ReadFile("queries/delete_vat.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
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
