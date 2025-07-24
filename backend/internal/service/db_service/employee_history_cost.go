package db_service

import (
	"database/sql"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/types"
	"liquiswiss/pkg/utils"
	"time"
)

func (s *DatabaseService) ListEmployeeHistoryCosts(userID int64, historyID int64, page int64, limit int64) ([]models.EmployeeHistoryCost, int64, error) {
	historyCosts := make([]models.EmployeeHistoryCost, 0)
	var totalCount int64

	query, err := sqlQueries.ReadFile("queries/list_employee_history_costs.sql")
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(string(query), historyID, userID, (page)*limit, 0)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var historyCostID int64

		err := rows.Scan(
			&historyCostID,
			&totalCount,
		)
		if err != nil {
			return nil, 0, err
		}

		historyCost, err := s.GetEmployeeHistoryCost(userID, historyCostID)
		if err != nil {
			return nil, 0, err
		}

		historyCosts = append(historyCosts, *historyCost)
	}

	return historyCosts, totalCount, nil
}

func (s *DatabaseService) GetEmployeeHistoryCost(userID int64, historyCostID int64) (*models.EmployeeHistoryCost, error) {
	var employeeHistoryCost models.EmployeeHistoryCost
	var labelID sql.NullInt64
	var labelName sql.NullString

	query, err := sqlQueries.ReadFile("queries/get_employee_history_cost.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), historyCostID, userID).Scan(
		&employeeHistoryCost.ID,
		&labelID,
		&labelName,
		&employeeHistoryCost.Cycle,
		&employeeHistoryCost.AmountType,
		&employeeHistoryCost.Amount,
		&employeeHistoryCost.DistributionType,
		&employeeHistoryCost.RelativeOffset,
		&employeeHistoryCost.TargetDate,
		&employeeHistoryCost.EmployeeHistoryID,
		&employeeHistoryCost.HistoryCycle,
		&employeeHistoryCost.HistorySalary,
		&employeeHistoryCost.HistoryFromDate,
		&employeeHistoryCost.HistoryToDate,
		&employeeHistoryCost.DBDate,
	)
	if err != nil {
		return nil, err
	}

	if labelID.Valid && labelName.Valid {
		employeeHistoryCost.Label = &models.EmployeeHistoryCostLabel{
			ID:   labelID.Int64,
			Name: labelName.String,
		}
	}

	costDetails, err := s.ListEmployeeHistoryCostDetails(employeeHistoryCost.ID)
	if err != nil {
		return nil, err
	}

	employeeHistoryCost.CalculatedCostDetails = costDetails

	currDate := time.Time(employeeHistoryCost.DBDate)

	var next *models.EmployeeHistoryCostDetail
	for i := range costDetails {
		costDetail := costDetails[i]
		dt, err := time.Parse("2006-01", costDetail.Month)
		if err != nil {
			continue
		}
		if currDate.Format("2006-01") == costDetail.Month || dt.After(currDate) {
			next = &costDetail
			break
		}
	}

	if next != nil {
		dt, err := time.Parse("2006-01", next.Month)
		if err != nil {
			return nil, err
		}
		dtAsDate := types.AsDate(dt)
		//employeeHistoryCost.CalculatedPreviousExecutionDate = ??
		employeeHistoryCost.CalculatedNextExecutionDate = &dtAsDate
		employeeHistoryCost.CalculatedNextCost = next.Amount
		employeeHistoryCost.CalculatedAmount = next.Amount / uint64(next.Divider)
	}

	return &employeeHistoryCost, nil
}

func (s *DatabaseService) CreateEmployeeHistoryCost(payload models.CreateEmployeeHistoryCost, userID int64, historyID int64) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_employee_history_cost.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	// Prepare target date
	var targetDate sql.NullTime
	if payload.TargetDate != nil {
		parsedTargetDate, err := time.Parse(utils.InternalDateFormat, *payload.TargetDate)
		if err != nil {
			return 0, err
		}
		targetDate = sql.NullTime{Time: parsedTargetDate, Valid: true}
	} else {
		targetDate = sql.NullTime{Valid: false}
	}

	res, err := stmt.Exec(
		payload.Cycle,
		payload.AmountType,
		payload.Amount,
		payload.DistributionType,
		payload.RelativeOffset,
		targetDate,
		payload.LabelID,
		historyID,
		historyID,
		userID,
	)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted history
	historyCostID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if historyCostID == 0 {
		return 0, sql.ErrNoRows
	}

	return historyCostID, nil
}

func (s *DatabaseService) UpdateEmployeeHistoryCost(payload models.CreateEmployeeHistoryCost, userID int64, historyCostID int64) error {
	query, err := sqlQueries.ReadFile("queries/update_employee_history_cost.sql")
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Prepare target date
	var targetDate sql.NullTime
	if payload.TargetDate != nil {
		parsedTargetDate, err := time.Parse(utils.InternalDateFormat, *payload.TargetDate)
		if err != nil {
			return err
		}
		targetDate = sql.NullTime{Time: parsedTargetDate, Valid: true}
	} else {
		targetDate = sql.NullTime{Valid: false}
	}

	_, err = stmt.Exec(
		payload.Cycle,
		payload.AmountType,
		payload.Amount,
		payload.DistributionType,
		payload.RelativeOffset,
		targetDate,
		payload.LabelID,
		historyCostID,
		userID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) DeleteEmployeeHistoryCost(historyCostID int64, userID int64) error {
	query, err := sqlQueries.ReadFile("queries/delete_employee_history_cost.sql")
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(historyCostID, userID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) CopyEmployeeHistoryCosts(payload models.CopyEmployeeHistoryCosts, userID int64, historyID int64) error {
	query, err := sqlQueries.ReadFile("queries/copy_employee_history_cost.sql")
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
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

	stmt, err := tx.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, id := range payload.IDs {
		exec, err := stmt.Exec(
			historyID,
			id,
			userID,
		)
		if err != nil {
			return err
		}
		affected, err := exec.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return sql.ErrNoRows
		}
	}

	return nil
}
