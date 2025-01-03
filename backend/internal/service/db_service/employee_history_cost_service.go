package db_service

import (
	"database/sql"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/types"
)

func (s *DatabaseService) ListEmployeeHistoryCost(employeeHistoryID, userID, limit int64) ([]models.EmployeeHistoryCost, int64, error) {
	subQuery, err := sqlQueries.ReadFile("queries/list_employee_history_costs.sql")
	if err != nil {
		return nil, 0, err
	}

	subRows, err := s.db.Query(string(subQuery), employeeHistoryID, userID, limit, 0)
	if err != nil {
		return nil, 0, err
	}
	defer subRows.Close()

	var totalSubCount int64

	var employeeHistoryCosts []models.EmployeeHistoryCost
	for subRows.Next() {
		var employeeHistoryCost models.EmployeeHistoryCost
		var targetDate sql.NullTime

		err := subRows.Scan(
			&employeeHistoryCost.ID,
			&employeeHistoryCost.Label.ID,
			&employeeHistoryCost.Label.Name,
			&employeeHistoryCost.Cycle,
			&employeeHistoryCost.AmountType,
			&employeeHistoryCost.Amount,
			&employeeHistoryCost.RelativeOffset,
			&targetDate,
			&employeeHistoryCost.PreviousExecutionDate,
			&employeeHistoryCost.NextExecutionDate,
			&employeeHistoryCost.NextCost,
			// Forget about this for now (or ever :D)
			&totalSubCount,
		)
		if err != nil {
			return nil, 0, err
		}

		if targetDate.Valid {
			convertedDate := types.AsDate(targetDate.Time)
			employeeHistoryCost.TargetDate = &convertedDate
		}

		employeeHistoryCosts = append(employeeHistoryCosts, employeeHistoryCost)
	}

	return employeeHistoryCosts, totalSubCount, nil
}
