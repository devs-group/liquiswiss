package db_service

import (
	"liquiswiss/pkg/models"
	"strings"
)

func (s *DatabaseService) ListCategories(userID, page, limit int64) ([]models.Category, int64, error) {
	categories := []models.Category{}
	var totalCount int64

	query, err := sqlQueries.ReadFile("queries/list_categories.sql")
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(string(query), userID, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var category models.Category

		err := rows.Scan(&category.ID, &category.Name, &category.CanEdit, &totalCount)
		if err != nil {
			return nil, 0, err
		}

		categories = append(categories, category)
	}

	return categories, totalCount, nil
}

func (s *DatabaseService) GetCategory(userID int64, categoryID int64) (*models.Category, error) {
	var category models.Category

	query, err := sqlQueries.ReadFile("queries/get_category.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), categoryID, userID).Scan(&category.ID, &category.Name, &category.CanEdit)
	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (s *DatabaseService) CreateCategory(userID *int64, payload models.CreateCategory) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_category.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(payload.Name, userID)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *DatabaseService) UpdateCategory(userID int64, payload models.UpdateCategory, categoryID int64) error {
	query := "UPDATE categories SET "
	queryBuild := []string{}
	args := []interface{}{}

	if payload.Name != nil {
		queryBuild = append(queryBuild, "name = ?")
		args = append(args, *payload.Name)
	}

	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ? AND organisation_id = get_current_user_organisation_id(?)"
	args = append(args, categoryID, userID)

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
