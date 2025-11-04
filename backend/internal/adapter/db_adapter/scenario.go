package db_adapter

import (
	"liquiswiss/pkg/models"
)

func (d *DatabaseAdapter) ListScenarios(userID int64) ([]models.ScenarioListItem, error) {
	query := `
		SELECT
			s.id,
			s.name,
			s.type,
			s.is_default,
			s.parent_scenario_id,
			s.created_at,
			s.updated_at
		FROM scenarios s
		WHERE s.organisation_id = get_current_user_organisation_id(?)
		ORDER BY s.is_default DESC, s.created_at ASC
	`

	rows, err := d.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scenarios []models.ScenarioListItem
	for rows.Next() {
		var scenario models.ScenarioListItem
		err := rows.Scan(
			&scenario.ID,
			&scenario.Name,
			&scenario.Type,
			&scenario.IsDefault,
			&scenario.ParentScenarioID,
			&scenario.CreatedAt,
			&scenario.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		scenarios = append(scenarios, scenario)
	}

	return scenarios, nil
}

func (d *DatabaseAdapter) GetScenario(userID int64, scenarioID int64) (*models.Scenario, error) {
	query := `
		SELECT
			s.id,
			s.name,
			s.type,
			s.is_default,
			s.parent_scenario_id,
			s.organisation_id,
			s.created_at,
			s.updated_at
		FROM scenarios s
		WHERE s.id = ?
		  AND s.organisation_id = get_current_user_organisation_id(?)
		LIMIT 1
	`

	var scenario models.Scenario
	err := d.db.QueryRow(query, scenarioID, userID).Scan(
		&scenario.ID,
		&scenario.Name,
		&scenario.Type,
		&scenario.IsDefault,
		&scenario.ParentScenarioID,
		&scenario.OrganisationID,
		&scenario.CreatedAt,
		&scenario.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &scenario, nil
}

func (d *DatabaseAdapter) GetDefaultScenario(userID int64) (*models.Scenario, error) {
	query := `
		SELECT
			s.id,
			s.name,
			s.type,
			s.is_default,
			s.parent_scenario_id,
			s.organisation_id,
			s.created_at,
			s.updated_at
		FROM scenarios s
		WHERE s.is_default = true
		  AND s.organisation_id = get_current_user_organisation_id(?)
		LIMIT 1
	`

	var scenario models.Scenario
	err := d.db.QueryRow(query, userID).Scan(
		&scenario.ID,
		&scenario.Name,
		&scenario.Type,
		&scenario.IsDefault,
		&scenario.ParentScenarioID,
		&scenario.OrganisationID,
		&scenario.CreatedAt,
		&scenario.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &scenario, nil
}

func (d *DatabaseAdapter) CreateScenario(payload models.CreateScenario, userID int64) (int64, error) {
	query := `
		INSERT INTO scenarios (name, type, parent_scenario_id, organisation_id)
		VALUES (?, ?, ?, get_current_user_organisation_id(?))
	`

	stmt, err := d.db.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		payload.Name,
		payload.Type,
		payload.ParentScenarioID,
		userID,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (d *DatabaseAdapter) UpdateScenario(payload models.UpdateScenario, userID int64, scenarioID int64) error {
	query := `
		UPDATE scenarios
		SET name = ?
		WHERE id = ?
		  AND organisation_id = get_current_user_organisation_id(?)
	`

	stmt, err := d.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(payload.Name, scenarioID, userID)
	return err
}

func (d *DatabaseAdapter) IsScenarioInUse(scenarioID int64) (bool, error) {
	query := `
		SELECT COUNT(*) > 0
		FROM users
		WHERE current_scenario_id = ?
	`

	var inUse bool
	err := d.db.QueryRow(query, scenarioID).Scan(&inUse)
	if err != nil {
		return false, err
	}

	return inUse, nil
}

func (d *DatabaseAdapter) GetUsersUsingScenario(scenarioID int64) ([]int64, error) {
	query := `
		SELECT id
		FROM users
		WHERE current_scenario_id = ?
	`

	rows, err := d.db.Query(query, scenarioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []int64
	for rows.Next() {
		var userID int64
		err := rows.Scan(&userID)
		if err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

func (d *DatabaseAdapter) DeleteScenario(userID int64, scenarioID int64) error {
	query := `
		DELETE FROM scenarios
		WHERE id = ?
		  AND organisation_id = get_current_user_organisation_id(?)
		  AND is_default = false
	`

	stmt, err := d.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(scenarioID, userID)
	return err
}

func (d *DatabaseAdapter) CopyScenarioData(sourceScenarioID int64, targetScenarioID int64, userID int64) error {
	tx, err := d.db.Begin()
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

	// Copy employees
	_, err = tx.Exec(`
		INSERT INTO employees (uuid, scenario_id, name, deleted, organisation_id, created_at)
		SELECT UUID(), ?, name, deleted, organisation_id, NOW()
		FROM employees
		WHERE scenario_id = ?
		  AND organisation_id = get_current_user_organisation_id(?)
	`, targetScenarioID, sourceScenarioID, userID)
	if err != nil {
		return err
	}

	// Copy salaries (matching employee UUIDs from source to target)
	_, err = tx.Exec(`
		INSERT INTO salaries (
			uuid, scenario_id, employee_id, cycle, hours_per_month, amount,
			vacation_days_per_year, from_date, to_date, is_termination,
			is_disabled, deleted, currency_id, created_at
		)
		SELECT
			UUID(),
			?,
			target_emp.id,
			src_sal.cycle,
			src_sal.hours_per_month,
			src_sal.amount,
			src_sal.vacation_days_per_year,
			src_sal.from_date,
			src_sal.to_date,
			src_sal.is_termination,
			src_sal.is_disabled,
			src_sal.deleted,
			src_sal.currency_id,
			NOW()
		FROM salaries src_sal
		INNER JOIN employees src_emp ON src_sal.employee_id = src_emp.id
		INNER JOIN employees target_emp ON target_emp.uuid = src_emp.uuid
		WHERE src_sal.scenario_id = ?
		  AND target_emp.scenario_id = ?
		  AND src_emp.organisation_id = get_current_user_organisation_id(?)
	`, targetScenarioID, sourceScenarioID, targetScenarioID, userID)
	if err != nil {
		return err
	}

	// Copy salary costs
	_, err = tx.Exec(`
		INSERT INTO salary_costs (
			uuid, scenario_id, salary_id, cycle, amount_type, amount,
			distribution_type, relative_offset, target_date, label_id
		)
		SELECT
			UUID(),
			?,
			target_sal.id,
			src_sc.cycle,
			src_sc.amount_type,
			src_sc.amount,
			src_sc.distribution_type,
			src_sc.relative_offset,
			src_sc.target_date,
			src_sc.label_id
		FROM salary_costs src_sc
		INNER JOIN salaries src_sal ON src_sc.salary_id = src_sal.id
		INNER JOIN salaries target_sal ON target_sal.uuid = src_sal.uuid
		WHERE src_sc.scenario_id = ?
		  AND target_sal.scenario_id = ?
	`, targetScenarioID, sourceScenarioID, targetScenarioID)
	if err != nil {
		return err
	}

	// Copy transactions
	_, err = tx.Exec(`
		INSERT INTO transactions (
			uuid, scenario_id, name, amount, vat_included, is_disabled, cycle,
			type, start_date, end_date, deleted, vat_id, category_id,
			employee_id, currency_id, organisation_id, created_at
		)
		SELECT
			UUID(),
			?,
			name,
			amount,
			vat_included,
			is_disabled,
			cycle,
			type,
			start_date,
			end_date,
			deleted,
			vat_id,
			category_id,
			employee_id,
			currency_id,
			organisation_id,
			NOW()
		FROM transactions
		WHERE scenario_id = ?
		  AND organisation_id = get_current_user_organisation_id(?)
	`, targetScenarioID, sourceScenarioID, userID)
	if err != nil {
		return err
	}

	return nil
}
