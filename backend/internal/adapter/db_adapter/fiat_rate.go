package db_adapter

import "liquiswiss/pkg/models"

func (d *DatabaseAdapter) ListFiatRates(base string) ([]models.FiatRate, error) {
	fiatRates := []models.FiatRate{}

	query, err := sqlQueries.ReadFile("queries/list_fiat_rates.sql")
	if err != nil {
		return nil, err
	}

	rows, err := d.db.Query(string(query), base)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var fiatRate models.FiatRate

		err := rows.Scan(&fiatRate.ID, &fiatRate.Base, &fiatRate.Target, &fiatRate.Rate, &fiatRate.UpdatedAt)
		if err != nil {
			return nil, err
		}

		fiatRates = append(fiatRates, fiatRate)
	}

	return fiatRates, nil
}

func (d *DatabaseAdapter) GetFiatRate(base, target string) (*models.FiatRate, error) {
	var fiatRate models.FiatRate

	query, err := sqlQueries.ReadFile("queries/get_fiat_rate.sql")
	if err != nil {
		return nil, err
	}

	err = d.db.QueryRow(string(query), base, target).Scan(
		&fiatRate.ID, &fiatRate.Base, &fiatRate.Target, &fiatRate.Rate, &fiatRate.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &fiatRate, nil
}

func (d *DatabaseAdapter) UpsertFiatRate(payload models.CreateFiatRate) error {
	query, err := sqlQueries.ReadFile("queries/upsert_fiat_rate.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		payload.Base, payload.Target, payload.Rate,
	)
	if err != nil {
		return err
	}

	return nil
}

func (d *DatabaseAdapter) CountUniqueCurrenciesInFiatRates() (int64, error) {
	query, err := sqlQueries.ReadFile("queries/count_unique_currencies_in_fiat_rates.sql")
	if err != nil {
		return 0, err
	}

	var totalCount int64
	err = d.db.QueryRow(string(query)).Scan(&totalCount)
	if err != nil {
		return 0, err
	}

	return totalCount, nil
}
