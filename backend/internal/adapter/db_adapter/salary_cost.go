package db_adapter

import (
	"database/sql"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"time"
)

func (d *DatabaseAdapter) ListSalaryCosts(userID int64, salaryID int64, page int64, limit int64) ([]models.SalaryCost, int64, error) {
	salaryCosts := make([]models.SalaryCost, 0)
	var totalCount int64

	query, err := sqlQueries.ReadFile("queries/list_salary_costs.sql")
	if err != nil {
		return nil, 0, err
	}

	rows, err := d.db.Query(string(query), salaryID, userID, (page)*limit, 0)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	salaryCostIDs := make([]int64, 0)
	for rows.Next() {
		var salaryCostID int64

		err := rows.Scan(
			&salaryCostID,
			&totalCount,
		)
		if err != nil {
			return nil, 0, err
		}

		salaryCostIDs = append(salaryCostIDs, salaryCostID)
	}

	for _, salaryCostID := range salaryCostIDs {
		salaryCost, err := d.GetSalaryCost(userID, salaryCostID)
		if err != nil {
			return nil, 0, err
		}
		salaryCosts = append(salaryCosts, *salaryCost)
	}

	return salaryCosts, totalCount, nil
}

func (d *DatabaseAdapter) GetSalaryCost(userID int64, salaryCostID int64) (*models.SalaryCost, error) {
	var salaryCost models.SalaryCost
	var labelID sql.NullInt64
	var labelName sql.NullString

	query, err := sqlQueries.ReadFile("queries/get_salary_cost.sql")
	if err != nil {
		return nil, err
	}

	err = d.db.QueryRow(string(query), salaryCostID, userID).Scan(
		&salaryCost.ID,
		&labelID,
		&labelName,
		&salaryCost.Cycle,
		&salaryCost.AmountType,
		&salaryCost.Amount,
		&salaryCost.DistributionType,
		&salaryCost.RelativeOffset,
		&salaryCost.TargetDate,
		&salaryCost.SalaryID,
		&salaryCost.SalaryCycle,
		&salaryCost.SalaryAmount,
		&salaryCost.SalaryFromDate,
		&salaryCost.SalaryToDate,
		&salaryCost.DBDate,
	)
	if err != nil {
		return nil, err
	}

	if labelID.Valid && labelName.Valid {
		salaryCost.Label = &models.SalaryCostLabel{
			ID:   labelID.Int64,
			Name: labelName.String,
		}
	}

	return &salaryCost, nil
}

func (d *DatabaseAdapter) CreateSalaryCost(payload models.CreateSalaryCost, userID int64, salaryID int64) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_salary_cost.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := d.db.Prepare(string(query))
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
		salaryID,
		salaryID,
		userID,
	)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted salary
	salaryCostID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if salaryCostID == 0 {
		return 0, sql.ErrNoRows
	}

	return salaryCostID, nil
}

func (d *DatabaseAdapter) UpdateSalaryCost(payload models.CreateSalaryCost, userID int64, salaryCostID int64) error {
	query, err := sqlQueries.ReadFile("queries/update_salary_cost.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
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
		salaryCostID,
		userID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (d *DatabaseAdapter) DeleteSalaryCost(salaryCostID int64, userID int64) error {
	query, err := sqlQueries.ReadFile("queries/delete_salary_cost.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(salaryCostID, userID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (d *DatabaseAdapter) CopySalaryCosts(payload models.CopySalaryCosts, userID int64, salaryID int64) error {
	query, err := sqlQueries.ReadFile("queries/copy_salary_cost.sql")
	if err != nil {
		return err
	}

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

	stmt, err := tx.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, id := range payload.IDs {
		exec, err := stmt.Exec(
			salaryID,
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
