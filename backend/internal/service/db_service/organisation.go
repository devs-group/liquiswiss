package db_service

import (
	"liquiswiss/pkg/models"
	"strings"
)

func (s *DatabaseService) ListOrganisations(userID int64, page int64, limit int64) ([]models.Organisation, int64, error) {
	organisations := []models.Organisation{}
	var totalCount int64

	query, err := sqlQueries.ReadFile("queries/list_organisations.sql")
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(string(query), userID, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var organisation models.Organisation

		err := rows.Scan(
			&organisation.ID,
			&organisation.Name,
			&organisation.Currency.ID,
			&organisation.Currency.Code,
			&organisation.Currency.Description,
			&organisation.Currency.LocaleCode,
			&organisation.MemberCount,
			&organisation.Role,
			&organisation.IsDefault,
			&totalCount,
		)
		if err != nil {
			return nil, 0, err
		}

		organisations = append(organisations, organisation)
	}

	return organisations, totalCount, nil
}

func (s *DatabaseService) GetOrganisation(userID int64, organisationID int64) (*models.Organisation, error) {
	var organisation models.Organisation

	query, err := sqlQueries.ReadFile("queries/get_organisation.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), organisationID, userID).Scan(
		&organisation.ID,
		&organisation.Name,
		&organisation.Currency.ID,
		&organisation.Currency.Code,
		&organisation.Currency.Description,
		&organisation.Currency.LocaleCode,
		&organisation.MemberCount,
		&organisation.Role,
	)
	if err != nil {
		return nil, err
	}

	return &organisation, nil
}

func (s *DatabaseService) CreateOrganisation(name string) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_organisation.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(name)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted user
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *DatabaseService) UpdateOrganisation(payload models.UpdateOrganisation, userID int64, organisationID int64) error {
	// Base query
	query := "UPDATE organisations SET "
	queryBuild := []string{}
	args := []interface{}{}

	// Dynamically add fields that are not nil
	if payload.Name != nil {
		queryBuild = append(queryBuild, "name = ?")
		args = append(args, *payload.Name)
	}
	if payload.CurrencyID != nil {
		queryBuild = append(queryBuild, "main_currency_id = ?")
		args = append(args, *payload.CurrencyID)
	}

	// Add WHERE clause
	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ? AND id = get_current_user_organisation_id(?)"
	args = append(args, organisationID, userID)

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

func (s *DatabaseService) AssignUserToOrganisation(userID int64, organisationID int64, role string, isDefault bool) error {
	query, err := sqlQueries.ReadFile("queries/assign_user_to_organisation.sql")
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, organisationID, role, isDefault)
	if err != nil {
		return err
	}

	return nil
}
