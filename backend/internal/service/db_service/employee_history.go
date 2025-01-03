package db_service

import (
	"database/sql"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/types"
	"liquiswiss/pkg/utils"
	"strings"
	"time"
)

func (s *DatabaseService) ListEmployeeHistory(userID int64, employeeID int64, page int64, limit int64) ([]models.EmployeeHistory, int64, error) {
	employeeHistories := make([]models.EmployeeHistory, 0)
	var totalCount int64

	query, err := sqlQueries.ReadFile("queries/list_employee_histories.sql")
	if err != nil {
		return nil, 0, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, 0, err
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

	rows, err := tx.Query(string(query), employeeID, userID, (page)*limit, 0)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var employeeHistory models.EmployeeHistory
		var fromDate sql.NullTime
		var toDate sql.NullTime

		employeeHistory.Currency = models.Currency{}

		err := rows.Scan(
			&employeeHistory.ID,
			&employeeHistory.EmployeeID,
			&employeeHistory.HoursPerMonth,
			&employeeHistory.Salary,
			&employeeHistory.Cycle,
			&employeeHistory.Currency.ID,
			&employeeHistory.Currency.LocaleCode,
			&employeeHistory.Currency.Description,
			&employeeHistory.Currency.Code,
			&employeeHistory.VacationDaysPerYear,
			&fromDate,
			&toDate,
			&employeeHistory.NextExecutionDate,
			&employeeHistory.EmployeeDeductions,
			&employeeHistory.EmployerCosts,
			&employeeHistory.WithSeparateCosts,
			&totalCount,
		)
		if err != nil {
			return nil, 0, err
		}

		if fromDate.Valid {
			convertedDate := types.AsDate(fromDate.Time)
			employeeHistory.FromDate = convertedDate
		}
		if toDate.Valid {
			convertedDate := types.AsDate(toDate.Time)
			employeeHistory.ToDate = &convertedDate
		}

		employeeHistories = append(employeeHistories, employeeHistory)
	}

	return employeeHistories, totalCount, nil
}

func (s *DatabaseService) GetEmployeeHistory(userID int64, historyID int64) (*models.EmployeeHistory, error) {
	var employeeHistory models.EmployeeHistory
	var fromDate time.Time
	var toDate sql.NullTime

	employeeHistory.Currency = models.Currency{}

	query, err := sqlQueries.ReadFile("queries/get_employee_history.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), historyID, userID).Scan(
		&employeeHistory.ID,
		&employeeHistory.EmployeeID,
		&employeeHistory.HoursPerMonth,
		&employeeHistory.Salary,
		&employeeHistory.Cycle,
		&employeeHistory.Currency.ID,
		&employeeHistory.Currency.LocaleCode,
		&employeeHistory.Currency.Description,
		&employeeHistory.Currency.Code,
		&employeeHistory.VacationDaysPerYear,
		&fromDate,
		&toDate,
		&employeeHistory.NextExecutionDate,
		&employeeHistory.EmployeeDeductions,
		&employeeHistory.EmployerCosts,
		&employeeHistory.WithSeparateCosts,
	)
	if err != nil {
		return nil, err
	}

	employeeHistory.FromDate = types.AsDate(fromDate)

	if toDate.Valid {
		convertedDate := types.AsDate(toDate.Time)
		employeeHistory.ToDate = &convertedDate
	}

	return &employeeHistory, nil
}

func (s *DatabaseService) CreateEmployeeHistory(payload models.CreateEmployeeHistory, userID int64, employeeID int64) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
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

	query, err := sqlQueries.ReadFile("queries/create_employee_history.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := tx.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	// Prepare entry and exit date
	fromDate, err := time.Parse(utils.InternalDateFormat, payload.FromDate)
	if err != nil {
		return 0, err
	}

	var toDate sql.NullTime
	if payload.ToDate != nil {
		parsedToDate, err := time.Parse(utils.InternalDateFormat, *payload.ToDate)
		if err != nil {
			return 0, err
		}
		toDate = sql.NullTime{Time: parsedToDate, Valid: true}
	} else {
		toDate = sql.NullTime{Valid: false}
	}

	// Activate separate costs by default
	payload.WithSeparateCosts = true

	res, err := stmt.Exec(
		employeeID,
		payload.HoursPerMonth,
		payload.Salary,
		payload.Cycle,
		payload.CurrencyID,
		payload.VacationDaysPerYear,
		fromDate,
		toDate,
		payload.WithSeparateCosts,
		employeeID,
		userID,
	)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted history
	historyID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if historyID == 0 {
		return 0, sql.ErrNoRows
	}

	// Fetch and adjust previous entry if required
	var previous models.EmployeeHistory
	if err := tx.QueryRow(`
        SELECT id, from_date, to_date, cycle FROM employee_histories 
        WHERE employee_id = ? AND from_date < ? AND id != ?
        ORDER BY from_date DESC LIMIT 1
    `, employeeID, payload.FromDate, historyID).
		Scan(&previous.ID, &previous.FromDate, &previous.ToDate, &previous.Cycle); err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if previous.ID != 0 {
		logger.Logger.Debugf("Checking previous history entry")
		currentFromDate, err := time.Parse(utils.InternalDateFormat, payload.FromDate)
		if err != nil {
			return 0, err
		}
		needsAdjustment := false
		if previous.ToDate == nil {
			needsAdjustment = true
		} else {
			previousToDate := time.Time(*previous.ToDate)
			if !previousToDate.Before(currentFromDate) {
				needsAdjustment = true
			}
		}

		if needsAdjustment {
			previousFromDate := time.Time(previous.FromDate)
			newToDate := utils.GetNextAvailableDate(previousFromDate, currentFromDate, previous.Cycle)
			logger.Logger.Debugf("Adjusting previous history entry to: %s", newToDate)
			_, err := tx.Exec(`
            UPDATE employee_histories SET to_date = ? WHERE id = ?
        `, newToDate, previous.ID)
			if err != nil {
				return 0, err
			}
		}
	}

	// Fetch and adjust current entry if required by next
	var next models.EmployeeHistory
	if err := tx.QueryRow(`
	   SELECT id, from_date, to_date FROM employee_histories
       WHERE employee_id = ? AND from_date > ? AND id != ?
	   ORDER BY from_date ASC LIMIT 1
	`, employeeID, payload.FromDate, historyID).
		Scan(&next.ID, &next.FromDate, &next.ToDate); err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if next.ID != 0 {
		logger.Logger.Debugf("Checking next history entry")
		needsAdjustment := false

		nextFromDate := time.Time(next.FromDate)
		if payload.ToDate == nil {
			needsAdjustment = true
		} else {
			currentToDate, err := time.Parse(utils.InternalDateFormat, *payload.ToDate)
			if err != nil {
				return 0, err
			}
			if !currentToDate.Before(nextFromDate) {
				needsAdjustment = true
			}
		}

		if needsAdjustment {
			currentFromDate, err := time.Parse(utils.InternalDateFormat, payload.FromDate)
			if err != nil {
				return 0, err
			}
			newToDate := utils.GetNextAvailableDate(currentFromDate, nextFromDate, payload.Cycle)
			logger.Logger.Debugf("Adjusting current history entry to: %s", newToDate)
			_, err = tx.Exec(`
            UPDATE employee_histories SET to_date = ? WHERE id = ?
        `, newToDate, historyID)
			if err != nil {
				return 0, err
			}
		}
	}

	return historyID, nil
}

func (s *DatabaseService) UpdateEmployeeHistory(payload models.UpdateEmployeeHistory, employeeID int64, historyID int64) error {
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

	// Base query
	query := "UPDATE employee_histories SET "
	queryBuild := []string{}
	args := []interface{}{}

	// Dynamically add fields that are not nil
	if payload.HoursPerMonth != nil {
		queryBuild = append(queryBuild, "hours_per_month = ?")
		args = append(args, *payload.HoursPerMonth)
	}
	if payload.Salary != nil {
		queryBuild = append(queryBuild, "salary = ?")
		args = append(args, *payload.Salary)
	}
	if payload.CurrencyID != nil {
		queryBuild = append(queryBuild, "currency_id = ?")
		args = append(args, *payload.CurrencyID)
	}
	if payload.VacationDaysPerYear != nil {
		queryBuild = append(queryBuild, "vacation_days_per_year = ?")
		args = append(args, *payload.VacationDaysPerYear)
	}
	if payload.FromDate != nil {
		queryBuild = append(queryBuild, "from_date = ?")
		entryDate, err := time.Parse(utils.InternalDateFormat, *payload.FromDate)
		if err != nil {
			return err
		}
		args = append(args, entryDate)
	}
	if payload.WithSeparateCosts != nil {
		queryBuild = append(queryBuild, "with_separate_costs = ?")
		args = append(args, *payload.WithSeparateCosts)
	}
	// Always consider ToDate in case it is set back to null
	queryBuild = append(queryBuild, "to_date = ?")
	if payload.ToDate != nil {
		exitDate, err := time.Parse(utils.InternalDateFormat, *payload.ToDate)
		if err != nil {
			return err
		}
		args = append(args, exitDate)
	} else {
		args = append(args, nil)
	}
	if payload.Cycle != nil {
		queryBuild = append(queryBuild, "cycle = ?")
		args = append(args, *payload.Cycle)
	}

	// Add WHERE clause
	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ? AND employee_id = ?"
	args = append(args, historyID, employeeID)

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}

	// Fetch and adjust previous entry if required
	var previous models.EmployeeHistory
	if err := tx.QueryRow(`
        SELECT id, from_date, to_date, cycle FROM employee_histories 
        WHERE employee_id = ? AND from_date < ? AND id != ?
        ORDER BY from_date DESC LIMIT 1
    `, employeeID, payload.FromDate, historyID).
		Scan(&previous.ID, &previous.FromDate, &previous.ToDate, &previous.Cycle); err != nil && err != sql.ErrNoRows {
		return err
	}
	if previous.ID != 0 {
		logger.Logger.Debugf("Checking previous history entry")
		currentFromDate, err := time.Parse(utils.InternalDateFormat, *payload.FromDate)
		if err != nil {
			return err
		}
		previousFromDate := time.Time(previous.FromDate)
		newToDate := utils.GetNextAvailableDate(previousFromDate, currentFromDate, previous.Cycle)
		logger.Logger.Debugf("Adjusting previous history entry (ID: %d) toDate to: %s", previous.ID, newToDate)
		_, err = tx.Exec(`
				UPDATE employee_histories SET to_date = ? WHERE id = ?
			`, newToDate, previous.ID)
		if err != nil {
			logger.Logger.Error(err)
			return err
		}
	}

	// Fetch and adjust current entry if required by next
	var next models.EmployeeHistory
	if err := tx.QueryRow(`
	   SELECT id, from_date, to_date FROM employee_histories
       WHERE employee_id = ? AND from_date > ? AND id != ?
	   ORDER BY from_date ASC LIMIT 1
	`, employeeID, payload.FromDate, historyID).Scan(&next.ID, &next.FromDate, &next.ToDate); err != nil && err != sql.ErrNoRows {
		return err
	}
	if next.ID != 0 {
		logger.Logger.Debugf("Checking next history entry")
		currentFromDate, err := time.Parse(utils.InternalDateFormat, *payload.FromDate)
		if err != nil {
			return err
		}
		nextFromDate := time.Time(next.FromDate)
		newToDate := utils.GetNextAvailableDate(currentFromDate, nextFromDate, *payload.Cycle)
		logger.Logger.Debugf("Adjusting current history entry (ID: %d) toDate to: %s", historyID, newToDate)
		_, err = tx.Exec(`
				UPDATE employee_histories SET to_date = ? WHERE id = ?
			`, newToDate, historyID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *DatabaseService) DeleteEmployeeHistory(toDeleteEmployeeHistory *models.EmployeeHistory, userID int64) error {
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

	query, err := sqlQueries.ReadFile("queries/delete_employee_history.sql")
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(toDeleteEmployeeHistory.ID, userID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	// In case there is a previous and next entry, adjust the previous toDate related to the next one
	currentFromDate := time.Time(toDeleteEmployeeHistory.FromDate)
	currentFromDateFormatted := currentFromDate.Format(utils.InternalDateFormat)
	var previous models.EmployeeHistory
	if err := tx.QueryRow(`
        SELECT id, from_date, to_date, cycle FROM employee_histories 
        WHERE employee_id = ? AND from_date < ? AND id != ?
        ORDER BY from_date DESC LIMIT 1
    `, toDeleteEmployeeHistory.EmployeeID, currentFromDateFormatted, toDeleteEmployeeHistory.ID,
	).Scan(&previous.ID, &previous.FromDate, &previous.ToDate, &previous.Cycle); err != nil && err != sql.ErrNoRows {
		return err
	}
	var next models.EmployeeHistory
	if err := tx.QueryRow(`
	   SELECT id, from_date, to_date FROM employee_histories
       WHERE employee_id = ? AND from_date > ? AND id != ?
	   ORDER BY from_date ASC LIMIT 1
	`, toDeleteEmployeeHistory.EmployeeID, currentFromDateFormatted, toDeleteEmployeeHistory.ID,
	).Scan(&next.ID, &next.FromDate, &next.ToDate); err != nil && err != sql.ErrNoRows {
		return err
	}

	if previous.ID != 0 && next.ID != 0 {
		previousFromDate := time.Time(previous.FromDate)
		nextFromDate := time.Time(next.FromDate)
		newToDate := utils.GetNextAvailableDate(previousFromDate, nextFromDate, previous.Cycle)
		logger.Logger.Debugf("Adjusting previous history entry (ID: %d) toDate to: %s", previous.ID, newToDate)
		_, err := tx.Exec(`
		   UPDATE employee_histories SET to_date = ? WHERE id = ?
		`, newToDate, previous.ID)
		if err != nil {
			return err
		}
	} else if previous.ID != 0 && next.ID == 0 {
		// If there is no next remove the toDate from the latest
		logger.Logger.Debugf("Adjusting previous history entry (ID: %d) toDate to: NULL", previous.ID)
		_, err := tx.Exec(`
		   UPDATE employee_histories SET to_date = ? WHERE id = ?
		`, nil, previous.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
