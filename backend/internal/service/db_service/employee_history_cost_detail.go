package db_service

import (
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"time"
)

func (s *DatabaseService) ListEmployeeHistoryCostDetails(historyCostID int64) ([]models.EmployeeHistoryCostDetail, error) {
	historyCostDetails := make([]models.EmployeeHistoryCostDetail, 0)

	query, err := sqlQueries.ReadFile("queries/list_employee_history_cost_details.sql")
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(string(query), historyCostID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var historyCostDetail models.EmployeeHistoryCostDetail

		err := rows.Scan(
			&historyCostDetail.ID,
			&historyCostDetail.Month,
			&historyCostDetail.Amount,
			&historyCostDetail.Divider,
			&historyCostDetail.CostID,
		)
		if err != nil {
			return nil, err
		}

		historyCostDetails = append(historyCostDetails, historyCostDetail)
	}

	return historyCostDetails, nil
}

func (s *DatabaseService) CalculateEmployeeHistoryCostDetails(historyCostID int64, userID int64) error {
	cost, err := s.GetEmployeeHistoryCost(userID, historyCostID)
	if err != nil {
		return err
	}
	history, err := s.GetEmployeeHistory(userID, cost.EmployeeHistoryID)
	if err != nil {
		return err
	}

	// Cleanup First
	err = s.ClearEmployeeHistoryCostDetails(cost.ID)
	if err != nil {
		return err
	}

	// Next Payment date
	currCostExecutionPtr := s.CalculateCostExecutionDate(
		history.FromDate,
		history.ToDate,
		history.Cycle,
		cost.TargetDate,
		cost.Cycle,
		cost.RelativeOffset,
		history.DBDate,
		true,
	)
	if currCostExecutionPtr == nil {
		return nil
	}
	nextCostExecution := time.Time(*currCostExecutionPtr)

	today := utils.GetTodayAsUTC()
	maxEndDate := today.AddDate(3, 0, 0)
	// We include the whole final month, otherwise the results might be confusing
	lastDayOfMaxEndDate := time.Date(maxEndDate.Year(), maxEndDate.Month()+1, 0, 23, 59, 59, 999999999, maxEndDate.Location())

	//currDate := time.Time(history.DBDate)
	fromDate := time.Time(history.FromDate)
	toDate := lastDayOfMaxEndDate
	if history.ToDate != nil {
		toDate = time.Time(*history.ToDate)
		toDate = time.Date(toDate.Year(), toDate.Month()+1, 0, 23, 59, 59, 999999999, toDate.Location())
	}

	for {
		// 1. Accumulate backwards
		validMonths := []time.Time{}
		hasEnded := false
		for i := cost.RelativeOffset; i > 0; i-- {
			month := addCycle(nextCostExecution, cost.Cycle, -i)
			if month.Before(fromDate) {
				continue
			}
			if month.After(toDate) {
				hasEnded = true
				continue
			}
			validMonths = append(validMonths, month)
		}

		amountPerMonth := s.CalculateCostAmount(*cost, *history)
		totalAmount := amountPerMonth * uint64(len(validMonths))
		if totalAmount > 0 {
			monthStr := nextCostExecution.Format("2006-01")
			payload := models.CreateEmployeeHistoryCostDetail{
				Month:   monthStr,
				Amount:  totalAmount,
				Divider: uint(len(validMonths)),
				CostID:  cost.ID,
			}
			_, err := s.UpsertEmployeeHistoryCostDetails(payload)
			if err != nil {
				return err
			}
		}

		if hasEnded {
			break
		}

		// 3. Calculate next payment date
		nextCostExecution = addCycle(nextCostExecution, cost.Cycle, cost.RelativeOffset)
	}

	return nil
}

func (s *DatabaseService) UpsertEmployeeHistoryCostDetails(payload models.CreateEmployeeHistoryCostDetail) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/upsert_employee_history_cost_detail.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		payload.Month,
		payload.Amount,
		payload.Divider,
		payload.CostID,
	)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted employee
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *DatabaseService) ClearEmployeeHistoryCostDetails(employeeHistoryCostDetailID int64) error {
	query, err := sqlQueries.ReadFile("queries/clear_employee_history_cost_detail.sql")
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(employeeHistoryCostDetailID)
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) RefreshCostDetails(userID int64, historyID int64) error {
	costs, _, err := s.ListEmployeeHistoryCosts(userID, historyID, 1, 10000)
	if err != nil {
		return err
	}
	for _, cost := range costs {
		err = s.CalculateEmployeeHistoryCostDetails(cost.ID, userID)
		if err != nil {
			return err
		}
	}
	return nil
}
