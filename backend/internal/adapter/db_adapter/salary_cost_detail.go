package db_adapter

import (
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"time"
)

func (d *DatabaseAdapter) ListSalaryCostDetails(salaryCostID int64) ([]models.SalaryCostDetail, error) {
	salaryCostDetails := make([]models.SalaryCostDetail, 0)

	query, err := sqlQueries.ReadFile("queries/list_salary_cost_details.sql")
	if err != nil {
		return nil, err
	}

	rows, err := d.db.Query(string(query), salaryCostID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var salaryCostDetail models.SalaryCostDetail

		err := rows.Scan(
			&salaryCostDetail.ID,
			&salaryCostDetail.Month,
			&salaryCostDetail.Amount,
			&salaryCostDetail.Divider,
			&salaryCostDetail.CostID,
		)
		if err != nil {
			return nil, err
		}

		salaryCostDetails = append(salaryCostDetails, salaryCostDetail)
	}

	return salaryCostDetails, nil
}

func (d *DatabaseAdapter) CalculateSalaryCostDetails(userID int64, salaryCostID int64) error {
	cost, err := d.GetSalaryCost(userID, salaryCostID)
	if err != nil {
		return err
	}
	salary, err := d.GetSalary(userID, cost.SalaryID)
	if err != nil {
		return err
	}

	// Cleanup First
	err = d.ClearSalaryCostDetails(cost.ID)
	if err != nil {
		return err
	}

	// Next Payment date
	currCostExecutionPtr := d.CalculateCostExecutionDate(
		salary.FromDate,
		salary.ToDate,
		salary.Cycle,
		cost.TargetDate,
		cost.Cycle,
		cost.RelativeOffset,
		salary.DBDate,
		true,
	)
	if currCostExecutionPtr == nil {
		return nil
	}
	nextCostExecution := time.Time(*currCostExecutionPtr)

	today := utils.GetTodayAsUTC()
	maxEndDate := today.AddDate(utils.MaxForecastYears, 0, 0)
	// We include the whole final month, otherwise the results might be confusing
	lastDayOfMaxEndDate := time.Date(maxEndDate.Year(), maxEndDate.Month()+1, 0, 23, 59, 59, 999999999, maxEndDate.Location())

	//currDate := time.Time(salary.DBDate)
	fromDate := time.Time(salary.FromDate)
	toDate := lastDayOfMaxEndDate
	if salary.ToDate != nil {
		toDate = time.Time(*salary.ToDate)
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

		amountPerMonth := d.CalculateCostAmount(*cost, *salary)
		totalAmount := amountPerMonth * uint64(len(validMonths))
		if totalAmount > 0 {
			monthStr := nextCostExecution.Format("2006-01")
			payload := models.CreateSalaryCostDetail{
				Month:   monthStr,
				Amount:  totalAmount,
				Divider: uint(len(validMonths)),
				CostID:  cost.ID,
			}
			_, err := d.UpsertSalaryCostDetails(payload)
			if err != nil {
				return err
			}
		}

		if hasEnded || cost.Cycle == utils.CycleOnce {
			break
		}

		// 3. Calculate next payment date
		nextCostExecution = addCycle(nextCostExecution, cost.Cycle, cost.RelativeOffset)
	}

	return nil
}

func (d *DatabaseAdapter) UpsertSalaryCostDetails(payload models.CreateSalaryCostDetail) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/upsert_salary_cost_detail.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := d.db.Prepare(string(query))
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

func (d *DatabaseAdapter) ClearSalaryCostDetails(salaryCostDetailID int64) error {
	query, err := sqlQueries.ReadFile("queries/clear_salary_cost_detail.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(salaryCostDetailID)
	if err != nil {
		return err
	}

	return nil
}

func (d *DatabaseAdapter) RefreshSalaryCostDetails(userID int64, salaryID int64) error {
	costs, _, err := d.ListSalaryCosts(userID, salaryID, 1, 10000)
	if err != nil {
		return err
	}
	for _, cost := range costs {
		err = d.CalculateSalaryCostDetails(cost.ID, userID)
		if err != nil {
			return err
		}
	}
	return nil
}
