package db_service

import (
	"liquiswiss/pkg/models"
	"strings"
)

func (s *DatabaseService) ListCurrencies(page int64, limit int64) ([]models.Currency, int64, error) {
	currencies := []models.Currency{}
	var totalCount int64

	query, err := sqlQueries.ReadFile("queries/list_currencies.sql")
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(string(query), limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var currency models.Currency

		err := rows.Scan(&currency.ID, &currency.Code, &currency.Description, &currency.LocaleCode, &totalCount)
		if err != nil {
			return nil, 0, err
		}

		currencies = append(currencies, currency)
	}

	return currencies, totalCount, nil
}

func (s *DatabaseService) GetCurrency(id string) (*models.Currency, error) {
	var currency models.Currency

	query, err := sqlQueries.ReadFile("queries/get_currency.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), id).Scan(
		&currency.ID, &currency.Code, &currency.Description, &currency.LocaleCode,
	)
	if err != nil {
		return nil, err
	}

	return &currency, nil
}

func (s *DatabaseService) CreateCurrency(payload models.CreateCurrency) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_currency.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(payload.Code, payload.Description, payload.LocaleCode)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *DatabaseService) UpdateCurrency(payload models.UpdateCurrency, currencyID string) error {
	query := "UPDATE currencies SET "
	queryBuild := []string{}
	args := []interface{}{}

	if payload.Code != nil {
		queryBuild = append(queryBuild, "code = ?")
		args = append(args, *payload.Code)
	}
	if payload.Description != nil {
		queryBuild = append(queryBuild, "description = ?")
		args = append(args, *payload.Description)
	}
	if payload.LocaleCode != nil {
		queryBuild = append(queryBuild, "locale_code = ?")
		args = append(args, *payload.LocaleCode)
	}

	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ?"
	args = append(args, currencyID)

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