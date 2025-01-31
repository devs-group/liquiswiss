package db_service

import (
	"encoding/json"
	"fmt"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"log"
)

func (s *DatabaseService) ListForecasts(userID int64, limit int64) ([]models.Forecast, error) {
	forecasts := make([]models.Forecast, 0)

	query, err := sqlQueries.ReadFile("queries/list_forecasts.sql")
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(string(query), limit, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var forecast models.Forecast

		err := rows.Scan(
			&forecast.Data.Month, &forecast.Data.Revenue, &forecast.Data.Expense, &forecast.Data.Cashflow,
			&forecast.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		forecasts = append(forecasts, forecast)
	}

	return forecasts, nil
}

func (s *DatabaseService) ListForecastDetails(userID int64, limit int64) ([]models.ForecastDatabaseDetails, error) {
	forecastDetails := make([]models.ForecastDatabaseDetails, 0)

	query, err := sqlQueries.ReadFile("queries/list_forecast_details.sql")
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(string(query), limit, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var forecastDetail models.ForecastDatabaseDetails
		var revenueJSON, expenseJSON []byte

		if err := rows.Scan(&forecastDetail.Month, &revenueJSON, &expenseJSON, &forecastDetail.ForecastID); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(revenueJSON, &forecastDetail.Revenue); err != nil {
			log.Fatal(err)
		}
		if err := json.Unmarshal(expenseJSON, &forecastDetail.Expense); err != nil {
			log.Fatal(err)
		}

		forecastDetails = append(forecastDetails, forecastDetail)
	}

	return forecastDetails, nil
}

func (s *DatabaseService) UpsertForecast(payload models.CreateForecast, userID int64) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/upsert_forecast.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		payload.Month, payload.Revenue, payload.Expense, payload.Cashflow, userID,
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

func (s *DatabaseService) UpsertForecastDetail(payload models.CreateForecastDetail, userID, forecastID int64) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/upsert_forecast_detail.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	revenueJson, err := json.Marshal(payload.Revenue)
	if err != nil {
		return 0, err
	}

	expenseJson, err := json.Marshal(payload.Expense)
	if err != nil {
		return 0, err
	}

	res, err := stmt.Exec(
		payload.Month, revenueJson, expenseJson, forecastID, userID,
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

func (s *DatabaseService) ListForecastExclusions(userID, relatedID int64, relatedTable string) (map[string]bool, error) {
	sqlFile := ""
	switch relatedTable {
	case utils.TransactionsTableName:
		sqlFile = "queries/list_transaction_exclusions.sql"
	case utils.EmployeeHistoriesTableName:
		sqlFile = "queries/list_employee_history_exclusions.sql"
	case utils.EmployeeHistoryCostsTableName:
		sqlFile = "queries/list_employee_history_cost_exclusions.sql"
	default:
		return nil, fmt.Errorf("invalid relatedTable")
	}

	query, err := sqlQueries.ReadFile(sqlFile)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(string(query), relatedID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	forecastExclusions := make(map[string]bool, 0)
	for rows.Next() {
		var forecastExclusion models.ForecastExclusion

		err := rows.Scan(
			&forecastExclusion.ID,
			&forecastExclusion.ExcludeMonth,
			&forecastExclusion.TransactionID,
		)
		if err != nil {
			return nil, err
		}

		forecastExclusions[forecastExclusion.ExcludeMonth] = true
	}

	return forecastExclusions, nil
}

func (s *DatabaseService) CreateForecastExclusion(payload models.CreateForecastExclusion, userID int64) (int64, error) {
	sqlFile := ""
	switch payload.RelatedTable {
	case utils.TransactionsTableName:
		sqlFile = "queries/create_transaction_exclusion.sql"
	case utils.EmployeeHistoriesTableName:
		sqlFile = "queries/create_employee_history_exclusion.sql"
	case utils.EmployeeHistoryCostsTableName:
		sqlFile = "queries/create_employee_history_cost_exclusion.sql"
	default:
		return 0, fmt.Errorf("invalid relatedTable")
	}

	query, err := sqlQueries.ReadFile(sqlFile)
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		payload.Month, payload.RelatedID, userID,
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

func (s *DatabaseService) DeleteForecastExclusion(payload models.CreateForecastExclusion, userID int64) (int64, error) {
	sqlFile := ""
	switch payload.RelatedTable {
	case utils.TransactionsTableName:
		sqlFile = "queries/delete_transaction_exclusion.sql"
	case utils.EmployeeHistoriesTableName:
		sqlFile = "queries/delete_employee_history_exclusion.sql"
	case utils.EmployeeHistoryCostsTableName:
		sqlFile = "queries/delete_employee_history_cost_exclusion.sql"
	default:
		return 0, fmt.Errorf("invalid relatedTable")
	}

	query, err := sqlQueries.ReadFile(sqlFile)
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		payload.Month, payload.RelatedID, userID,
	)
	if err != nil {
		return 0, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}

func (s *DatabaseService) ClearForecasts(userID int64) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/clear_forecasts.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(userID)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted employee
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}
