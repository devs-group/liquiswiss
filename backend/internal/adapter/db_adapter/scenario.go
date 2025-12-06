package db_adapter

import (
	"liquiswiss/pkg/models"
	"strings"
)

func (d *DatabaseAdapter) ListScenarios(userID int64) ([]models.Scenario, error) {
	scenarios := make([]models.Scenario, 0)

	query, err := sqlQueries.ReadFile("queries/list_scenarios.sql")
	if err != nil {
		return nil, err
	}

	rows, err := d.db.Query(string(query), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var scenario models.Scenario

		err := rows.Scan(
			&scenario.ID, &scenario.Name, &scenario.IsDefault, &scenario.CreatedAt, &scenario.ParentScenarioID,
		)
		if err != nil {
			return nil, err
		}

		scenarios = append(scenarios, scenario)
	}

	return scenarios, nil
}

func (d *DatabaseAdapter) GetScenario(userID int64, scenarioID int64) (*models.Scenario, error) {
	var scenario models.Scenario

	query, err := sqlQueries.ReadFile("queries/get_scenario.sql")
	if err != nil {
		return nil, err
	}

	err = d.db.QueryRow(string(query), scenarioID, userID).Scan(
		&scenario.ID, &scenario.Name, &scenario.IsDefault, &scenario.CreatedAt, &scenario.ParentScenarioID,
	)
	if err != nil {
		return nil, err
	}

	return &scenario, nil
}

func (d *DatabaseAdapter) CreateScenario(payload models.CreateScenario, userID int64, isDefault bool) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_scenario.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		payload.Name, payload.ParentScenarioID, isDefault, userID,
	)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted scenario
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (d *DatabaseAdapter) UpdateScenario(payload models.UpdateScenario, userID int64, scenarioID int64) error {
	// Base query
	query := "UPDATE scenarios SET "
	queryBuild := []string{}
	args := []interface{}{}

	// Dynamically add fields that are not nil
	if payload.Name != nil {
		queryBuild = append(queryBuild, "name = ?")
		args = append(args, *payload.Name)
	}
	if payload.ParentScenarioID != nil {
		queryBuild = append(queryBuild, "parent_scenario_id = ?")
		args = append(args, *payload.ParentScenarioID)
	} else {
		queryBuild = append(queryBuild, "parent_scenario_id = NULL")
	}

	// Add WHERE clause
	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ? AND organisation_id = get_current_user_organisation_id(?)"
	args = append(args, scenarioID, userID)

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

func (d *DatabaseAdapter) DeleteScenario(userID int64, scenarioID int64) error {
	query, err := sqlQueries.ReadFile("queries/delete_scenario.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(scenarioID, userID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (d *DatabaseAdapter) AssignUserToScenario(userID int64, organisationID int64, scenarioID int64) error {
	query, err := sqlQueries.ReadFile("queries/assign_user_to_scenario.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, organisationID, scenarioID)
	if err != nil {
		return err
	}

	return nil
}
