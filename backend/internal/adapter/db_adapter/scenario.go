package db_adapter

import (
	"database/sql"
	"strings"

	"liquiswiss/pkg/models"
)

func (d *DatabaseAdapter) ListScenarios(userID, page, limit int64) ([]models.Scenario, int64, error) {
	scenarios := []models.Scenario{}
	var totalCount int64

	query, err := sqlQueries.ReadFile("queries/list_scenarios.sql")
	if err != nil {
		return nil, 0, err
	}

	rows, err := d.db.Query(string(query), userID, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var scenario models.Scenario
		var parentID sql.NullInt64

		err := rows.Scan(
			&scenario.ID,
			&scenario.Name,
			&scenario.IsDefault,
			&scenario.CreatedAt,
			&parentID,
			&scenario.OrganisationID,
			&totalCount,
		)
		if err != nil {
			return nil, 0, err
		}

		if parentID.Valid {
			scenario.ParentScenarioID = &parentID.Int64
		} else {
			scenario.ParentScenarioID = nil
		}

		scenarios = append(scenarios, scenario)
	}

	return scenarios, totalCount, nil
}

func (d *DatabaseAdapter) GetScenario(userID, scenarioID int64) (*models.Scenario, error) {
	query, err := sqlQueries.ReadFile("queries/get_scenario.sql")
	if err != nil {
		return nil, err
	}

	var scenario models.Scenario
	var parentID sql.NullInt64

	err = d.db.QueryRow(string(query), scenarioID, userID).Scan(
		&scenario.ID,
		&scenario.Name,
		&scenario.IsDefault,
		&scenario.CreatedAt,
		&parentID,
		&scenario.OrganisationID,
	)
	if err != nil {
		return nil, err
	}

	if parentID.Valid {
		scenario.ParentScenarioID = &parentID.Int64
	}

	return &scenario, nil
}

func (d *DatabaseAdapter) CreateScenario(payload models.CreateScenario, userID int64) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_scenario.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(payload.Name, payload.IsDefault, payload.ParentScenarioID, userID)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (d *DatabaseAdapter) CreateScenarioForOrganisation(payload models.CreateScenario, organisationID int64) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_scenario_for_organisation.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(payload.Name, payload.IsDefault, payload.ParentScenarioID, organisationID)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (d *DatabaseAdapter) UpdateScenario(payload models.UpdateScenario, userID, scenarioID int64) error {
	setParts := []string{}
	args := []interface{}{}

	if payload.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *payload.Name)
	}

	if payload.IsDefault != nil {
		setParts = append(setParts, "is_default = ?")
		args = append(args, *payload.IsDefault)
	}

	if payload.ParentScenarioID != nil {
		setParts = append(setParts, "parent_scenario_id = ?")
		args = append(args, payload.ParentScenarioID)
	}

	if len(setParts) == 0 {
		return nil
	}

	query := "UPDATE scenarios SET " + strings.Join(setParts, ", ") + " WHERE id = ? AND organisation_id = get_current_user_organisation_id(?)"
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

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (d *DatabaseAdapter) DeleteScenario(userID, scenarioID int64) error {
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

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (d *DatabaseAdapter) GetDefaultScenarioID(organisationID int64) (int64, error) {
	query := "SELECT id FROM scenarios WHERE organisation_id = ? AND is_default = TRUE LIMIT 1"

	var scenarioID int64
	err := d.db.QueryRow(query, organisationID).Scan(&scenarioID)
	if err != nil {
		return 0, err
	}

	return scenarioID, nil
}
