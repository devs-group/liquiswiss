package db_adapter

import (
	"database/sql"
	"errors"
	"fmt"
	"liquiswiss/pkg/models"
	"strings"
)

// ErrNoScenario means the insert found no active scenario for the user (no organisation or no longer a member)
var ErrNoScenario = errors.New("no active scenario")

func scenarioInsertID(res sql.Result) (int64, error) {
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if affected == 0 {
		return 0, ErrNoScenario
	}
	return res.LastInsertId()
}

func scanScenario(scanner interface{ Scan(...any) error }) (models.Scenario, error) {
	var scenario models.Scenario
	var parentID sql.NullInt64
	err := scanner.Scan(&scenario.ID, &scenario.Name, &scenario.IsBase, &parentID, &scenario.CreatedAt, &scenario.IsCurrent)
	if err != nil {
		return scenario, err
	}
	if parentID.Valid {
		scenario.ParentID = &parentID.Int64
	}
	return scenario, nil
}

func (d *DatabaseAdapter) GetCurrentScenarioID(userID int64) (int64, error) {
	var scenarioID sql.NullInt64
	if err := d.db.QueryRow("SELECT get_current_user_scenario_id(?)", userID).Scan(&scenarioID); err != nil {
		return 0, err
	}
	if !scenarioID.Valid {
		return 0, sql.ErrNoRows
	}
	return scenarioID.Int64, nil
}

func (d *DatabaseAdapter) ListScenarios(userID int64) ([]models.Scenario, error) {
	scenarios := make([]models.Scenario, 0)

	query, err := sqlQueries.ReadFile("queries/list_scenarios.sql")
	if err != nil {
		return nil, err
	}

	rows, err := d.db.Query(string(query), userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		scenario, err := scanScenario(rows)
		if err != nil {
			return nil, err
		}
		scenarios = append(scenarios, scenario)
	}

	return scenarios, rows.Err()
}

func (d *DatabaseAdapter) GetScenario(userID int64, scenarioID int64) (*models.Scenario, error) {
	query, err := sqlQueries.ReadFile("queries/get_scenario.sql")
	if err != nil {
		return nil, err
	}

	scenario, err := scanScenario(d.db.QueryRow(string(query), userID, scenarioID, userID))
	if err != nil {
		return nil, err
	}

	return &scenario, nil
}

func (d *DatabaseAdapter) UpdateScenario(payload models.UpdateScenario, userID int64, scenarioID int64) error {
	if payload.Name == nil {
		return nil
	}

	query, err := sqlQueries.ReadFile("queries/update_scenario.sql")
	if err != nil {
		return err
	}

	_, err = d.db.Exec(string(query), *payload.Name, scenarioID, userID)
	return err
}

func (d *DatabaseAdapter) DeleteScenario(userID int64, scenarioID int64) error {
	query, err := sqlQueries.ReadFile("queries/delete_scenario.sql")
	if err != nil {
		return err
	}

	res, err := d.db.Exec(string(query), scenarioID, userID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (d *DatabaseAdapter) SetUserCurrentScenario(userID int64, scenarioID int64) error {
	query, err := sqlQueries.ReadFile("queries/set_user_current_scenario.sql")
	if err != nil {
		return err
	}

	res, err := d.db.Exec(string(query), userID, scenarioID, userID)
	if err != nil {
		return err
	}

	// ON DUPLICATE KEY reports 0 when the value did not change, so check membership separately
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		if _, err := d.GetScenario(userID, scenarioID); err != nil {
			return err
		}
	}

	return nil
}

// CreateScenario copies every scenario-bound row of the source scenario into a new scenario in one transaction
func (d *DatabaseAdapter) CreateScenario(userID int64, name string, sourceScenarioID int64) (int64, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var organisationID int64
	err = tx.QueryRow(
		"SELECT organisation_id FROM scenarios WHERE id = ? AND organisation_id = get_current_user_organisation_id(?)",
		sourceScenarioID, userID,
	).Scan(&organisationID)
	if err != nil {
		return 0, err
	}

	res, err := tx.Exec(
		"INSERT INTO scenarios (name, is_base, parent_id, organisation_id) VALUES (?, false, ?, ?)",
		name, sourceScenarioID, organisationID,
	)
	if err != nil {
		return 0, err
	}
	newScenarioID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	c := scenarioCopier{tx: tx, columns: map[string][]string{}}

	employeeMap, err := c.copyEmployees(sourceScenarioID, newScenarioID)
	if err != nil {
		return 0, err
	}
	if err := c.copyTransactions(sourceScenarioID, newScenarioID, employeeMap); err != nil {
		return 0, err
	}
	if err := c.copyRows("bank_accounts", "scenario_id = ?", []any{sourceScenarioID}, map[string]any{"scenario_id": newScenarioID}); err != nil {
		return 0, err
	}
	if err := c.copyRows("salary_cost_exclusions", "scenario_id = ?", []any{sourceScenarioID}, map[string]any{"scenario_id": newScenarioID}); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return newScenarioID, nil
}

type scenarioCopier struct {
	tx      *sql.Tx
	columns map[string][]string
}

// tableColumns returns every writable column except id, so new columns are copied without touching this code
func (c *scenarioCopier) tableColumns(table string) ([]string, error) {
	if cols, ok := c.columns[table]; ok {
		return cols, nil
	}

	rows, err := c.tx.Query(
		`SELECT COLUMN_NAME FROM information_schema.COLUMNS
		 WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME <> 'id' AND EXTRA NOT LIKE '%GENERATED%'
		 ORDER BY ORDINAL_POSITION`,
		table,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols := make([]string, 0)
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return nil, err
		}
		cols = append(cols, col)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("no columns found for table %s", table)
	}

	c.columns[table] = cols
	return cols, nil
}

func (c *scenarioCopier) buildCopy(table string, where string, overrides map[string]any) (string, []any, error) {
	cols, err := c.tableColumns(table)
	if err != nil {
		return "", nil, err
	}

	quoted := make([]string, len(cols))
	selects := make([]string, len(cols))
	args := make([]any, 0, len(overrides))
	for i, col := range cols {
		quoted[i] = "`" + col + "`"
		if value, ok := overrides[col]; ok {
			selects[i] = "?"
			args = append(args, value)
		} else {
			selects[i] = "`" + col + "`"
		}
	}

	query := fmt.Sprintf(
		"INSERT INTO `%s` (%s) SELECT %s FROM `%s` WHERE %s",
		table, strings.Join(quoted, ", "), strings.Join(selects, ", "), table, where,
	)
	return query, args, nil
}

func (c *scenarioCopier) copyRows(table string, where string, whereArgs []any, overrides map[string]any) error {
	query, args, err := c.buildCopy(table, where, overrides)
	if err != nil {
		return err
	}
	_, err = c.tx.Exec(query, append(args, whereArgs...)...)
	return err
}

func (c *scenarioCopier) copyRow(table string, id int64, overrides map[string]any) (int64, error) {
	query, args, err := c.buildCopy(table, "id = ?", overrides)
	if err != nil {
		return 0, err
	}
	res, err := c.tx.Exec(query, append(args, id)...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (c *scenarioCopier) ids(query string, args ...any) ([]int64, error) {
	rows, err := c.tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (c *scenarioCopier) copyEmployees(sourceScenarioID, newScenarioID int64) (map[int64]int64, error) {
	employeeMap := map[int64]int64{}
	costMap := map[int64]int64{}

	employeeIDs, err := c.ids("SELECT id FROM employees WHERE scenario_id = ?", sourceScenarioID)
	if err != nil {
		return nil, err
	}

	for _, employeeID := range employeeIDs {
		newEmployeeID, err := c.copyRow("employees", employeeID, map[string]any{"scenario_id": newScenarioID})
		if err != nil {
			return nil, err
		}
		employeeMap[employeeID] = newEmployeeID

		salaryIDs, err := c.ids("SELECT id FROM salaries WHERE employee_id = ?", employeeID)
		if err != nil {
			return nil, err
		}
		for _, salaryID := range salaryIDs {
			newSalaryID, err := c.copyRow("salaries", salaryID, map[string]any{"employee_id": newEmployeeID})
			if err != nil {
				return nil, err
			}
			if err := c.copyRows("salary_exclusions", "salary_id = ?", []any{salaryID}, map[string]any{"salary_id": newSalaryID}); err != nil {
				return nil, err
			}

			costIDs, err := c.ids("SELECT id FROM salary_costs WHERE salary_id = ?", salaryID)
			if err != nil {
				return nil, err
			}
			for _, costID := range costIDs {
				newCostID, err := c.copyRow("salary_costs", costID, map[string]any{"salary_id": newSalaryID})
				if err != nil {
					return nil, err
				}
				costMap[costID] = newCostID
				if err := c.copyRows("salary_cost_details", "cost_id = ?", []any{costID}, map[string]any{"cost_id": newCostID}); err != nil {
					return nil, err
				}
			}
		}
	}

	return employeeMap, c.copyBaseLinks(costMap)
}

// copyBaseLinks runs after all costs exist, because a link can point to any cost of the scenario
func (c *scenarioCopier) copyBaseLinks(costMap map[int64]int64) error {
	if len(costMap) == 0 {
		return nil
	}

	placeholders := make([]string, 0, len(costMap))
	args := make([]any, 0, len(costMap))
	for oldID := range costMap {
		placeholders = append(placeholders, "?")
		args = append(args, oldID)
	}

	rows, err := c.tx.Query(
		"SELECT cost_id, base_cost_id FROM salary_cost_base_links WHERE cost_id IN ("+strings.Join(placeholders, ", ")+")",
		args...,
	)
	if err != nil {
		return err
	}
	type link struct{ costID, baseCostID int64 }
	links := make([]link, 0)
	for rows.Next() {
		var l link
		if err := rows.Scan(&l.costID, &l.baseCostID); err != nil {
			rows.Close()
			return err
		}
		links = append(links, l)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, l := range links {
		newBaseID, ok := costMap[l.baseCostID]
		if !ok {
			continue
		}
		if _, err := c.tx.Exec(
			"INSERT INTO salary_cost_base_links (cost_id, base_cost_id) VALUES (?, ?)",
			costMap[l.costID], newBaseID,
		); err != nil {
			return err
		}
	}

	return nil
}

func (c *scenarioCopier) copyTransactions(sourceScenarioID, newScenarioID int64, employeeMap map[int64]int64) error {
	type transactionRef struct {
		id         int64
		employeeID sql.NullInt64
	}

	rows, err := c.tx.Query("SELECT id, employee_id FROM transactions WHERE scenario_id = ?", sourceScenarioID)
	if err != nil {
		return err
	}
	refs := make([]transactionRef, 0)
	for rows.Next() {
		var ref transactionRef
		if err := rows.Scan(&ref.id, &ref.employeeID); err != nil {
			rows.Close()
			return err
		}
		refs = append(refs, ref)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, ref := range refs {
		overrides := map[string]any{"scenario_id": newScenarioID}
		if ref.employeeID.Valid {
			if newEmployeeID, ok := employeeMap[ref.employeeID.Int64]; ok {
				overrides["employee_id"] = newEmployeeID
			} else {
				overrides["employee_id"] = nil
			}
		}

		newTransactionID, err := c.copyRow("transactions", ref.id, overrides)
		if err != nil {
			return err
		}
		if err := c.copyRows("transaction_exclusions", "transaction_id = ?", []any{ref.id}, map[string]any{"transaction_id": newTransactionID}); err != nil {
			return err
		}
	}

	return nil
}
