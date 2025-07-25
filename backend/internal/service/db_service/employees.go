package db_service

import (
	"database/sql"
	"fmt"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/types"
	"strings"
)

func (s *DatabaseService) ListEmployees(userID int64, page int64, limit int64, sortBy string, sortOrder string) ([]models.Employee, int64, error) {
	employees := make([]models.Employee, 0)
	var totalCount int64
	sortByMap := map[string]string{
		"name": "e.name", "hoursPerMonth": "s.hours_per_month", "salary": "s.amount", "vacationDaysPerYear": "s.vacation_days_per_year",
		"fromDate": "s.from_date", "toDate": "s.to_date",
	}

	// Validate inputs
	sortBy = sortByMap[sortBy]
	if sortBy == "" || !allowedSortOrders[sortOrder] {
		return nil, 0, fmt.Errorf("invalid sort by or sort order")
	}

	query, err := sqlQueries.ReadFile("queries/list_employees.sql")
	if err != nil {
		return nil, 0, err
	}

	queryString := fmt.Sprintf(string(query), sortBy, sortBy, sortOrder)

	rows, err := s.db.Query(queryString, userID, (page)*limit, 0)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var employee models.Employee
		var fromDate sql.NullTime
		var toDate sql.NullTime

		employee.Currency = &models.Currency{}

		err := rows.Scan(
			&employee.ID,
			&employee.Name,
			&employee.HoursPerMonth,
			&employee.SalaryAmount,
			&employee.Cycle,
			&employee.Currency.ID,
			&employee.Currency.LocaleCode,
			&employee.Currency.Description,
			&employee.Currency.Code,
			&employee.VacationDaysPerYear,
			&fromDate,
			&toDate,
			&employee.IsInFuture,
			&employee.WithSeparateCosts,
			&employee.SalaryID,
			&totalCount,
		)
		if err != nil {
			return nil, 0, err
		}

		if fromDate.Valid {
			convertedDate := types.AsDate(fromDate.Time)
			employee.FromDate = &convertedDate
		}
		if toDate.Valid {
			convertedDate := types.AsDate(toDate.Time)
			employee.ToDate = &convertedDate
		}

		employees = append(employees, employee)
	}

	return employees, totalCount, nil
}

func (s *DatabaseService) GetEmployee(userID int64, employeeID int64) (*models.Employee, error) {
	var employee models.Employee
	var fromDate sql.NullTime
	var toDate sql.NullTime

	employee.Currency = &models.Currency{}

	query, err := sqlQueries.ReadFile("queries/get_employee.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), employeeID, userID).Scan(
		&employee.ID,
		&employee.Name,
		&employee.HoursPerMonth,
		&employee.SalaryAmount,
		&employee.Cycle,
		&employee.Currency.ID,
		&employee.Currency.LocaleCode,
		&employee.Currency.Description,
		&employee.Currency.Code,
		&employee.VacationDaysPerYear,
		&fromDate,
		&toDate,
		&employee.IsInFuture,
		&employee.WithSeparateCosts,
		&employee.SalaryID,
	)
	if err != nil {
		return nil, err
	}

	if fromDate.Valid {
		convertedDate := types.AsDate(fromDate.Time)
		employee.FromDate = &convertedDate
	}
	if toDate.Valid {
		convertedDate := types.AsDate(toDate.Time)
		employee.ToDate = &convertedDate
	}

	return &employee, nil
}

func (s *DatabaseService) CreateEmployee(payload models.CreateEmployee, userID int64) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_employee.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		payload.Name, userID,
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

func (s *DatabaseService) UpdateEmployee(payload models.UpdateEmployee, userID int64, employeeID int64) error {
	// Base query
	query := "UPDATE employees SET "
	queryBuild := []string{}
	args := []interface{}{}

	// Dynamically add fields that are not nil
	if payload.Name != nil {
		queryBuild = append(queryBuild, "name = ?")
		args = append(args, *payload.Name)
	}

	// Add WHERE clause
	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ? AND organisation_id = get_current_user_organisation_id(?)"
	args = append(args, employeeID, userID)

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) DeleteEmployee(employeeID int64, userID int64) error {
	query, err := sqlQueries.ReadFile("queries/delete_employee.sql")
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(employeeID, userID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) CountEmployees(userID int64, page int64, limit int64) (int64, error) {
	var totalCount int64

	query, err := sqlQueries.ReadFile("queries/count_employees.sql")
	if err != nil {
		return 0, err
	}

	rows, err := s.db.Query(string(query), userID, limit, (page-1)*limit)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&totalCount)
		if err != nil {
			return 0, err
		}
	}

	return totalCount, nil
}
