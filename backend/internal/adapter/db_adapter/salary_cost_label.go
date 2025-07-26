package db_adapter

import (
	"database/sql"
	"liquiswiss/pkg/models"
)

func (d *DatabaseAdapter) ListSalaryCostLabels(userID int64, page int64, limit int64) ([]models.SalaryCostLabel, int64, error) {
	salaryCostLabels := make([]models.SalaryCostLabel, 0)
	var totalCount int64

	query, err := sqlQueries.ReadFile("queries/list_salary_cost_labels.sql")
	if err != nil {
		return nil, 0, err
	}

	rows, err := d.db.Query(string(query), userID, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var salaryCostLabel models.SalaryCostLabel

		err := rows.Scan(
			&salaryCostLabel.ID,
			&salaryCostLabel.Name,
			&totalCount,
		)
		if err != nil {
			return nil, 0, err
		}

		salaryCostLabels = append(salaryCostLabels, salaryCostLabel)
	}

	return salaryCostLabels, totalCount, nil
}

func (d *DatabaseAdapter) GetSalaryCostLabel(userID int64, salaryCostLabelID int64) (*models.SalaryCostLabel, error) {
	var salaryCostLabel models.SalaryCostLabel

	query, err := sqlQueries.ReadFile("queries/get_salary_cost_label.sql")
	if err != nil {
		return nil, err
	}

	err = d.db.QueryRow(string(query), salaryCostLabelID, userID).Scan(
		&salaryCostLabel.ID,
		&salaryCostLabel.Name,
	)
	if err != nil {
		return nil, err
	}

	return &salaryCostLabel, nil
}

func (d *DatabaseAdapter) CreateSalaryCostLabel(payload models.CreateSalaryCostLabel, userID int64) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_salary_cost_labels.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		payload.Name,
		userID,
	)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted salary cost label
	salaryCostLabelID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if salaryCostLabelID == 0 {
		return 0, sql.ErrNoRows
	}

	return salaryCostLabelID, nil
}

func (d *DatabaseAdapter) UpdateSalaryCostLabel(payload models.CreateSalaryCostLabel, userID int64, salaryCostLabelID int64) error {
	query, err := sqlQueries.ReadFile("queries/update_salary_cost_label.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		payload.Name,
		salaryCostLabelID,
		userID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (d *DatabaseAdapter) DeleteSalaryCostLabel(userID int64, salaryCostLabelID int64) error {
	query, err := sqlQueries.ReadFile("queries/delete_salary_cost_label.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(salaryCostLabelID, userID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}
