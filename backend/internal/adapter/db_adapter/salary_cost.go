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

	baseIDs, err := d.ListSalaryCostBaseIDs(salaryCost.ID)
	if err != nil {
		return nil, err
	}
	salaryCost.BaseSalaryCostIDs = baseIDs

	return &salaryCost, nil
}

func (d *DatabaseAdapter) ListSalaryCostBaseIDs(costID int64) ([]int64, error) {
	query, err := sqlQueries.ReadFile("queries/list_salary_cost_base_ids.sql")
	if err != nil {
		return nil, err
	}

	rows, err := d.db.Query(string(query), costID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	baseIDs := make([]int64, 0)
	for rows.Next() {
		var baseID int64
		if err := rows.Scan(&baseID); err != nil {
			return nil, err
		}
		baseIDs = append(baseIDs, baseID)
	}

	return baseIDs, nil
}

func (d *DatabaseAdapter) SetSalaryCostBaseLinks(costID int64, baseIDs []int64) (err error) {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
		if err != nil {
			tx.Rollback()
		}
	}()

	deleteQuery, err := sqlQueries.ReadFile("queries/delete_salary_cost_base_links.sql")
	if err != nil {
		return err
	}
	deleteStmt, err := tx.Prepare(string(deleteQuery))
	if err != nil {
		return err
	}
	defer deleteStmt.Close()

	if _, err = deleteStmt.Exec(costID); err != nil {
		return err
	}

	if len(baseIDs) == 0 {
		err = tx.Commit()
		return err
	}

	insertQuery, err := sqlQueries.ReadFile("queries/insert_salary_cost_base_link.sql")
	if err != nil {
		return err
	}
	insertStmt, err := tx.Prepare(string(insertQuery))
	if err != nil {
		return err
	}
	defer insertStmt.Close()

	unique := make(map[int64]struct{}, len(baseIDs))
	for _, baseID := range baseIDs {
		if _, exists := unique[baseID]; exists {
			continue
		}
		unique[baseID] = struct{}{}

		if _, err = insertStmt.Exec(costID, baseID); err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
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

func (d *DatabaseAdapter) DeleteSalaryCost(userID int64, salaryCostID int64) error {
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

	oldToNewIDs := make(map[int64]int64, len(payload.IDs))
	baseLinks := make(map[int64][]int64, len(payload.IDs))

	for _, id := range payload.IDs {
		links, err := d.ListSalaryCostBaseIDs(id)
		if err != nil {
			return err
		}
		baseLinks[id] = links

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

		insertedID, err := exec.LastInsertId()
		if err != nil {
			return err
		}
		oldToNewIDs[id] = insertedID
	}

	deleteQuery, err := sqlQueries.ReadFile("queries/delete_salary_cost_base_links.sql")
	if err != nil {
		return err
	}
	deleteStmt, err := tx.Prepare(string(deleteQuery))
	if err != nil {
		return err
	}
	defer deleteStmt.Close()

	insertQuery, err := sqlQueries.ReadFile("queries/insert_salary_cost_base_link.sql")
	if err != nil {
		return err
	}
	insertStmt, err := tx.Prepare(string(insertQuery))
	if err != nil {
		return err
	}
	defer insertStmt.Close()

	for srcID, newID := range oldToNewIDs {
		if _, err = deleteStmt.Exec(newID); err != nil {
			return err
		}

		links := baseLinks[srcID]
		if len(links) == 0 {
			continue
		}

		unique := make(map[int64]struct{})
		for _, baseID := range links {
			mappedID, ok := oldToNewIDs[baseID]
			if !ok {
				continue
			}
			if mappedID == newID {
				continue
			}
			if _, exists := unique[mappedID]; exists {
				continue
			}
			unique[mappedID] = struct{}{}

			if _, err = insertStmt.Exec(newID, mappedID); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *DatabaseAdapter) DeleteSalaryCostsBySalaryID(salaryID int64) error {
	query, err := sqlQueries.ReadFile("queries/delete_salary_costs_by_salary.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(salaryID)
	if err != nil {
		return err
	}

	return nil
}
