package db_adapter

import (
	"database/sql"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/types"
	"liquiswiss/pkg/utils"
	"strings"
	"time"
)

func (d *DatabaseAdapter) ListSalaries(userID int64, employeeID int64, page int64, limit int64) ([]models.Salary, int64, error) {
	salaries := make([]models.Salary, 0)
	var totalCount int64

	query, err := sqlQueries.ReadFile("queries/list_salaries.sql")
	if err != nil {
		return nil, 0, err
	}

	rows, err := d.db.Query(string(query), employeeID, userID, (page)*limit, 0)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	salaryIDs := make([]int64, 0)
	for rows.Next() {
		var salaryID int64

		err := rows.Scan(
			&salaryID,
			&totalCount,
		)
		if err != nil {
			return nil, 0, err
		}

		salaryIDs = append(salaryIDs, salaryID)

	}

	for _, salaryID := range salaryIDs {
		salary, err := d.GetSalary(userID, salaryID)
		if err != nil {
			return nil, 0, err
		}
		salaries = append(salaries, *salary)
	}

	return salaries, totalCount, nil
}

func (d *DatabaseAdapter) GetSalary(userID int64, salaryID int64) (*models.Salary, error) {
	var salary models.Salary
	var toDate sql.NullTime

	salary.Currency = models.Currency{}

	query, err := sqlQueries.ReadFile("queries/get_salary.sql")
	if err != nil {
		return nil, err
	}

	err = d.db.QueryRow(string(query), salaryID, userID).Scan(
		&salary.ID,
		&salary.EmployeeID,
		&salary.HoursPerMonth,
		&salary.Amount,
		&salary.Cycle,
		&salary.Currency.ID,
		&salary.Currency.LocaleCode,
		&salary.Currency.Description,
		&salary.Currency.Code,
		&salary.VacationDaysPerYear,
		&salary.FromDate,
		&toDate,
		&salary.WithSeparateCosts,
		&salary.IsTermination,
		&salary.IsDisabled,
		&salary.DBDate,
	)
	if err != nil {
		return nil, err
	}

	if toDate.Valid {
		convertedDate := types.AsDate(toDate.Time)
		salary.ToDate = &convertedDate
	}

	return &salary, nil
}

func (d *DatabaseAdapter) CreateSalary(payload models.CreateSalary, userID int64, employeeID int64) (int64, int64, int64, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return 0, 0, 0, err
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

	query, err := sqlQueries.ReadFile("queries/create_salary.sql")
	if err != nil {
		return 0, 0, 0, err
	}

	stmt, err := tx.Prepare(string(query))
	if err != nil {
		return 0, 0, 0, err
	}
	defer stmt.Close()

	// Prepare entry and exit date
	fromDate, err := time.Parse(utils.InternalDateFormat, payload.FromDate)
	if err != nil {
		return 0, 0, 0, err
	}

	if payload.IsTermination {
		payload.HoursPerMonth = 0
		payload.Amount = 0
		payload.VacationDaysPerYear = 0
		payload.WithSeparateCosts = false
		payload.ToDate = nil
	}

	var toDate sql.NullTime
	if payload.ToDate != nil {
		parsedToDate, err := time.Parse(utils.InternalDateFormat, *payload.ToDate)
		if err != nil {
			return 0, 0, 0, err
		}
		toDate = sql.NullTime{Time: parsedToDate, Valid: true}
	} else {
		toDate = sql.NullTime{Valid: false}
	}

	// Activate separate costs by default
	if !payload.IsTermination {
		payload.WithSeparateCosts = true
	}

	res, err := stmt.Exec(
		employeeID,
		payload.HoursPerMonth,
		payload.Amount,
		payload.Cycle,
		payload.CurrencyID,
		payload.VacationDaysPerYear,
		fromDate,
		toDate,
		payload.WithSeparateCosts,
		payload.IsTermination,
		employeeID,
		userID,
	)
	if err != nil {
		return 0, 0, 0, err
	}

	// Get the ID of the newly inserted salary
	salaryID, err := res.LastInsertId()
	if err != nil {
		return 0, 0, 0, err
	}
	if salaryID == 0 {
		return 0, 0, 0, sql.ErrNoRows
	}

	// Fetch and adjust previous entry if required
	var previous models.Salary
	if err := tx.QueryRow(`
        SELECT id, from_date, to_date, cycle FROM salaries 
        WHERE employee_id = ? AND from_date < ? AND id != ? AND is_disabled = 0
        ORDER BY from_date DESC LIMIT 1
    `, employeeID, payload.FromDate, salaryID).
		Scan(&previous.ID, &previous.FromDate, &previous.ToDate, &previous.Cycle); err != nil && err != sql.ErrNoRows {
		return 0, 0, 0, err
	}

	if previous.ID != 0 {
		logger.Logger.Debugf("Checking previous salary entry")
		currentFromDate, err := time.Parse(utils.InternalDateFormat, payload.FromDate)
		if err != nil {
			return 0, 0, 0, err
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
			logger.Logger.Debugf("Adjusting previous salary entry to: %d", newToDate)
			_, err := tx.Exec(`
            UPDATE salaries SET to_date = ? WHERE id = ?
        `, newToDate, previous.ID)
			if err != nil {
				return 0, 0, 0, err
			}
		}
	}

	// Fetch and adjust current entry if required by next
	var next models.Salary
	if err := tx.QueryRow(`
	   SELECT id, from_date, to_date FROM salaries
       WHERE employee_id = ? AND from_date > ? AND id != ? AND is_disabled = 0
	   ORDER BY from_date ASC LIMIT 1
	`, employeeID, payload.FromDate, salaryID).
		Scan(&next.ID, &next.FromDate, &next.ToDate); err != nil && err != sql.ErrNoRows {
		return 0, 0, 0, err
	}

	if next.ID != 0 {
		logger.Logger.Debugf("Checking next salary entry")
		needsAdjustment := false

		nextFromDate := time.Time(next.FromDate)
		if payload.ToDate == nil {
			needsAdjustment = true
		} else {
			currentToDate, err := time.Parse(utils.InternalDateFormat, *payload.ToDate)
			if err != nil {
				return 0, 0, 0, err
			}
			if !currentToDate.Before(nextFromDate) {
				needsAdjustment = true
			}
		}

		if needsAdjustment {
			currentFromDate, err := time.Parse(utils.InternalDateFormat, payload.FromDate)
			if err != nil {
				return 0, 0, 0, err
			}
			newToDate := utils.GetNextAvailableDate(currentFromDate, nextFromDate, payload.Cycle)
			logger.Logger.Debugf("Adjusting current salary entry to: %d", newToDate)
			_, err = tx.Exec(`
            UPDATE salaries SET to_date = ? WHERE id = ?
        `, newToDate, salaryID)
			if err != nil {
				return 0, 0, 0, err
			}
		}
	}

	return salaryID, previous.ID, next.ID, nil
}

func (d *DatabaseAdapter) UpdateSalary(payload models.UpdateSalary, employeeID int64, salaryID int64) (int64, int64, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return 0, 0, err
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
	query := "UPDATE salaries SET "
	queryBuild := []string{}
	args := []interface{}{}

	// Dynamically add fields that are not nil
	if payload.HoursPerMonth != nil {
		queryBuild = append(queryBuild, "hours_per_month = ?")
		args = append(args, *payload.HoursPerMonth)
	}
	if payload.Amount != nil {
		queryBuild = append(queryBuild, "amount = ?")
		args = append(args, *payload.Amount)
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
			return 0, 0, err
		}
		args = append(args, entryDate)
	}
	if payload.WithSeparateCosts != nil {
		queryBuild = append(queryBuild, "with_separate_costs = ?")
		args = append(args, *payload.WithSeparateCosts)
	}
	if payload.IsDisabled != nil {
		queryBuild = append(queryBuild, "is_disabled = ?")
		args = append(args, *payload.IsDisabled)
	}
	// Always consider ToDate in case it is set back to null
	queryBuild = append(queryBuild, "to_date = ?")
	if payload.ToDate != nil {
		exitDate, err := time.Parse(utils.InternalDateFormat, *payload.ToDate)
		if err != nil {
			return 0, 0, err
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
	args = append(args, salaryID, employeeID)

	stmt, err := tx.Prepare(query)
	if err != nil {
		return 0, 0, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return 0, 0, err
	}

	// Fetch and adjust previous entry if required
	var previous models.Salary
	if err := tx.QueryRow(`
        SELECT id, from_date, to_date, cycle FROM salaries 
        WHERE employee_id = ? AND from_date < ? AND id != ? AND is_disabled = 0
        ORDER BY from_date DESC LIMIT 1
    `, employeeID, payload.FromDate, salaryID).
		Scan(&previous.ID, &previous.FromDate, &previous.ToDate, &previous.Cycle); err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	if previous.ID != 0 {
		logger.Logger.Debugf("Checking previous salary entry")
		currentFromDate, err := time.Parse(utils.InternalDateFormat, *payload.FromDate)
		if err != nil {
			return 0, 0, err
		}
		previousFromDate := time.Time(previous.FromDate)
		newToDate := utils.GetNextAvailableDate(previousFromDate, currentFromDate, previous.Cycle)
		logger.Logger.Debugf("Adjusting previous salary entry (ID: %d) toDate to: %d", previous.ID, newToDate)
		_, err = tx.Exec(`
				UPDATE salaries SET to_date = ? WHERE id = ?
			`, newToDate, previous.ID)
		if err != nil {
			logger.Logger.Error(err)
			return 0, 0, err
		}
	}

	// Fetch and adjust current entry if required by next
	var next models.Salary
	if err := tx.QueryRow(`
	   SELECT id, from_date, to_date FROM salaries
       WHERE employee_id = ? AND from_date > ? AND id != ? AND is_disabled = 0
	   ORDER BY from_date ASC LIMIT 1
	`, employeeID, payload.FromDate, salaryID).Scan(&next.ID, &next.FromDate, &next.ToDate); err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	if next.ID != 0 {
		logger.Logger.Debugf("Checking next salary entry")
		currentFromDate, err := time.Parse(utils.InternalDateFormat, *payload.FromDate)
		if err != nil {
			return 0, 0, err
		}
		nextFromDate := time.Time(next.FromDate)
		newToDate := utils.GetNextAvailableDate(currentFromDate, nextFromDate, *payload.Cycle)
		logger.Logger.Debugf("Adjusting current salary entry (ID: %d) toDate to: %d", salaryID, newToDate)
		_, err = tx.Exec(`
				UPDATE salaries SET to_date = ? WHERE id = ?
			`, newToDate, salaryID)
		if err != nil {
			return 0, 0, err
		}
	}

	return previous.ID, next.ID, nil
}

func (d *DatabaseAdapter) DeleteSalary(toDeleteSalary *models.Salary, userID int64) (int64, int64, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return 0, 0, err
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

	query, err := sqlQueries.ReadFile("queries/delete_salary.sql")
	if err != nil {
		return 0, 0, err
	}

	stmt, err := tx.Prepare(string(query))
	if err != nil {
		return 0, 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(toDeleteSalary.ID, userID)
	if err != nil {
		return 0, 0, err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return 0, 0, err
	}

	// In case there is a previous and next entry, adjust the previous toDate related to the next one
	currentFromDate := time.Time(toDeleteSalary.FromDate)
	currentFromDateFormatted := currentFromDate.Format(utils.InternalDateFormat)
	var previous models.Salary
	if err := tx.QueryRow(`
        SELECT id, from_date, to_date, cycle FROM salaries 
        WHERE employee_id = ? AND from_date < ? AND id != ? AND is_disabled = 0
        ORDER BY from_date DESC LIMIT 1
    `, toDeleteSalary.EmployeeID, currentFromDateFormatted, toDeleteSalary.ID,
	).Scan(&previous.ID, &previous.FromDate, &previous.ToDate, &previous.Cycle); err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	var next models.Salary
	if err := tx.QueryRow(`
	   SELECT id, from_date, to_date FROM salaries
       WHERE employee_id = ? AND from_date > ? AND id != ? AND is_disabled = 0
	   ORDER BY from_date ASC LIMIT 1
	`, toDeleteSalary.EmployeeID, currentFromDateFormatted, toDeleteSalary.ID,
	).Scan(&next.ID, &next.FromDate, &next.ToDate); err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}

	if previous.ID != 0 && next.ID != 0 {
		previousFromDate := time.Time(previous.FromDate)
		nextFromDate := time.Time(next.FromDate)
		newToDate := utils.GetNextAvailableDate(previousFromDate, nextFromDate, previous.Cycle)
		logger.Logger.Debugf("Adjusting previous salary entry (ID: %d) toDate to: %d", previous.ID, newToDate)
		_, err := tx.Exec(`
		   UPDATE salaries SET to_date = ? WHERE id = ?
		`, newToDate, previous.ID)
		if err != nil {
			return 0, 0, err
		}
	} else if previous.ID != 0 && next.ID == 0 {
		// If there is no next remove the toDate from the latest
		logger.Logger.Debugf("Adjusting previous salary entry (ID: %d) toDate to: NULL", previous.ID)
		_, err := tx.Exec(`
		   UPDATE salaries SET to_date = ? WHERE id = ?
		`, nil, previous.ID)
		if err != nil {
			return 0, 0, err
		}
	}

	return previous.ID, next.ID, nil
}
