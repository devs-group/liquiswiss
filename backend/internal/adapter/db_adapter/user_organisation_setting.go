package db_adapter

import (
	"database/sql"
	"liquiswiss/pkg/models"
	"strings"
)

func (d *DatabaseAdapter) GetUserOrganisationSetting(userID int64) (*models.UserOrganisationSetting, error) {
	var setting models.UserOrganisationSetting

	query, err := sqlQueries.ReadFile("queries/get_user_organisation_setting.sql")
	if err != nil {
		return nil, err
	}

	err = d.db.QueryRow(string(query), userID, userID).Scan(
		&setting.ID,
		&setting.UserID,
		&setting.OrganisationID,
		&setting.ForecastMonths,
		&setting.ForecastPerformance,
		&setting.ForecastRevenueDetails,
		&setting.ForecastExpenseDetails,
		&setting.ForecastChildDetails,
		&setting.EmployeeDisplay,
		&setting.EmployeeSortBy,
		&setting.EmployeeSortOrder,
		&setting.EmployeeHideTerminated,
		&setting.TransactionDisplay,
		&setting.TransactionSortBy,
		&setting.TransactionSortOrder,
		&setting.TransactionHideDisabled,
		&setting.BankAccountDisplay,
		&setting.BankAccountSortBy,
		&setting.BankAccountSortOrder,
		&setting.CreatedAt,
		&setting.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &setting, nil
}

func (d *DatabaseAdapter) CreateUserOrganisationSetting(userID int64) (int64, error) {
	query := `INSERT INTO user_organisation_settings (user_id, organisation_id)
		VALUES (?, get_current_user_organisation_id(?))`

	stmt, err := d.db.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(userID, userID)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (d *DatabaseAdapter) UpdateUserOrganisationSetting(payload models.UpdateUserOrganisationSetting, userID int64) error {
	query := "UPDATE user_organisation_settings SET "
	queryBuild := []string{}
	args := []any{}

	if payload.ForecastMonths != nil {
		queryBuild = append(queryBuild, "forecast_months = ?")
		args = append(args, *payload.ForecastMonths)
	}

	if payload.ForecastPerformance != nil {
		queryBuild = append(queryBuild, "forecast_performance = ?")
		args = append(args, *payload.ForecastPerformance)
	}

	if payload.ForecastRevenueDetails != nil {
		queryBuild = append(queryBuild, "forecast_revenue_details = ?")
		args = append(args, *payload.ForecastRevenueDetails)
	}

	if payload.ForecastExpenseDetails != nil {
		queryBuild = append(queryBuild, "forecast_expense_details = ?")
		args = append(args, *payload.ForecastExpenseDetails)
	}

	if payload.ForecastChildDetails != nil {
		queryBuild = append(queryBuild, "forecast_child_details = ?")
		args = append(args, *payload.ForecastChildDetails)
	}

	if payload.EmployeeDisplay != nil {
		queryBuild = append(queryBuild, "employee_display = ?")
		args = append(args, *payload.EmployeeDisplay)
	}

	if payload.EmployeeSortBy != nil {
		queryBuild = append(queryBuild, "employee_sort_by = ?")
		args = append(args, *payload.EmployeeSortBy)
	}

	if payload.EmployeeSortOrder != nil {
		queryBuild = append(queryBuild, "employee_sort_order = ?")
		args = append(args, *payload.EmployeeSortOrder)
	}

	if payload.EmployeeHideTerminated != nil {
		queryBuild = append(queryBuild, "employee_hide_terminated = ?")
		args = append(args, *payload.EmployeeHideTerminated)
	}

	if payload.TransactionDisplay != nil {
		queryBuild = append(queryBuild, "transaction_display = ?")
		args = append(args, *payload.TransactionDisplay)
	}

	if payload.TransactionSortBy != nil {
		queryBuild = append(queryBuild, "transaction_sort_by = ?")
		args = append(args, *payload.TransactionSortBy)
	}

	if payload.TransactionSortOrder != nil {
		queryBuild = append(queryBuild, "transaction_sort_order = ?")
		args = append(args, *payload.TransactionSortOrder)
	}

	if payload.TransactionHideDisabled != nil {
		queryBuild = append(queryBuild, "transaction_hide_disabled = ?")
		args = append(args, *payload.TransactionHideDisabled)
	}

	if payload.BankAccountDisplay != nil {
		queryBuild = append(queryBuild, "bank_account_display = ?")
		args = append(args, *payload.BankAccountDisplay)
	}

	if payload.BankAccountSortBy != nil {
		queryBuild = append(queryBuild, "bank_account_sort_by = ?")
		args = append(args, *payload.BankAccountSortBy)
	}

	if payload.BankAccountSortOrder != nil {
		queryBuild = append(queryBuild, "bank_account_sort_order = ?")
		args = append(args, *payload.BankAccountSortOrder)
	}

	if len(queryBuild) == 0 {
		return nil
	}

	query += strings.Join(queryBuild, ", ")
	query += " WHERE user_id = ? AND organisation_id = get_current_user_organisation_id(?)"
	args = append(args, userID, userID)

	stmt, err := d.db.Prepare(query)
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
