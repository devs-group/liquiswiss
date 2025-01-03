package db_service

import (
	"database/sql"
	"liquiswiss/pkg/models"
)

func (s *DatabaseService) ListEmployeeHistoryCostLabels(userID int64, page int64, limit int64) ([]models.EmployeeHistoryCostLabel, int64, error) {
	employeeHistoryCostLabels := make([]models.EmployeeHistoryCostLabel, 0)
	var totalCount int64

	query, err := sqlQueries.ReadFile("queries/list_employee_history_cost_labels.sql")
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(string(query), userID, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var employeeHistoryCostLabel models.EmployeeHistoryCostLabel

		err := rows.Scan(
			&employeeHistoryCostLabel.ID,
			&employeeHistoryCostLabel.Name,
			&totalCount,
		)
		if err != nil {
			return nil, 0, err
		}

		employeeHistoryCostLabels = append(employeeHistoryCostLabels, employeeHistoryCostLabel)
	}

	return employeeHistoryCostLabels, totalCount, nil
}

func (s *DatabaseService) GetEmployeeHistoryCostLabel(userID int64, historyCostLabelID int64) (*models.EmployeeHistoryCostLabel, error) {
	var historyCostLabel models.EmployeeHistoryCostLabel

	query, err := sqlQueries.ReadFile("queries/get_employee_history_cost_label.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), historyCostLabelID, userID).Scan(
		&historyCostLabel.ID,
		&historyCostLabel.Name,
	)
	if err != nil {
		return nil, err
	}

	return &historyCostLabel, nil
}

func (s *DatabaseService) CreateEmployeeHistoryCostLabel(payload models.CreateEmployeeHistoryCostLabel, userID int64) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_employee_history_cost_labels.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
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

	// Get the ID of the newly inserted history
	historyCostLabelID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if historyCostLabelID == 0 {
		return 0, sql.ErrNoRows
	}

	return historyCostLabelID, nil
}

func (s *DatabaseService) UpdateEmployeeHistoryCostLabel(payload models.CreateEmployeeHistoryCostLabel, userID int64, historyCostLabelID int64) error {
	query, err := sqlQueries.ReadFile("queries/update_employee_history_cost_label.sql")
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		payload.Name,
		historyCostLabelID,
		userID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) DeleteEmployeeHistoryCostLabel(historyCostLabelID int64, userID int64) error {
	query, err := sqlQueries.ReadFile("queries/delete_employee_history_cost_label.sql")
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(historyCostLabelID, userID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}
