package db_service

import (
	"errors"
	"liquiswiss/pkg/models"
	"strings"
	"time"
)

func (s *DatabaseService) GetProfile(userID int64) (*models.User, error) {
	var user models.User

	query, err := sqlQueries.ReadFile("queries/get_profile.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), userID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CurrentOrganisationID,
		&user.Currency.ID,
		&user.Currency.Code,
		&user.Currency.Description,
		&user.Currency.LocaleCode,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *DatabaseService) GetUserPasswordByEMail(email string) (*models.Login, error) {
	var loginUser models.Login

	query, err := sqlQueries.ReadFile("queries/get_user_password_by_email.sql")
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(string(query), email).Scan(
		&loginUser.ID, &loginUser.Password,
	)
	if err != nil {
		return nil, err
	}

	return &loginUser, nil
}

func (s *DatabaseService) CreateUser(email string, password string) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_user.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(email, password)
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

func (s *DatabaseService) UpdateProfile(payload models.UpdateUser, userID int64) error {
	// Base query
	query := "UPDATE users SET "
	queryBuild := []string{}
	args := []interface{}{}

	// Dynamically add fields that are not nil
	if payload.Name != nil {
		queryBuild = append(queryBuild, "name = ?")
		args = append(args, *payload.Name)
	}

	// TODO: Disabled until email change logic is implemented
	//if payload.Email != nil {
	//	queryBuild = append(queryBuild, "email = ?")
	//	args = append(args, *payload.Email)
	//}

	// Add WHERE clause
	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ?"
	args = append(args, userID)

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

func (s *DatabaseService) UpdatePassword(password string, userID int64) error {
	// Base query
	query := "UPDATE users SET "
	queryBuild := []string{}
	args := []interface{}{}

	// Dynamically add fields that are not nil
	queryBuild = append(queryBuild, "password = ?")
	args = append(args, password)

	// Add WHERE clause
	query += strings.Join(queryBuild, ", ")
	query += " WHERE id = ?"
	args = append(args, userID)

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

// ResetPassword is only called internally and only after the reset password has been confirmed
func (s *DatabaseService) ResetPassword(password string, email string) error {
	// Base query
	query := "UPDATE users SET "
	queryBuild := []string{}
	args := []interface{}{}

	// Dynamically add fields that are not nil
	queryBuild = append(queryBuild, "password = ?")
	args = append(args, password)

	// Add WHERE clause
	query += strings.Join(queryBuild, ", ")
	query += " WHERE email = ?"
	args = append(args, email)

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

func (s *DatabaseService) CheckUserExistence(id int64) (bool, error) {
	query, err := sqlQueries.ReadFile("queries/check_user_existence.sql")
	if err != nil {
		return false, err
	}

	var exists bool
	err = s.db.QueryRow(string(query), id).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *DatabaseService) SetUserCurrentOrganisation(userID int64, organisationID int64) error {
	query, err := sqlQueries.ReadFile("queries/set_user_current_organisation.sql")
	if err != nil {
		return err
	}

	res, err := s.db.Exec(string(query), organisationID, userID, organisationID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) CreateResetPassword(email, code string, delay time.Duration) (bool, error) {
	query, err := sqlQueries.ReadFile("queries/create_reset_password.sql")
	if err != nil {
		return false, err
	}

	res, err := s.db.Exec(string(query), email, code, email, email, delay.Minutes())
	if err != nil {
		return false, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

func (s *DatabaseService) ValidateResetPassword(email, code string, hours time.Duration) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/validate_reset_password.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var id int64
	err = stmt.QueryRow(email, code, hours.Hours()).Scan(&id)
	if err != nil {
		return 0, err
	}
	if id <= 0 {
		return 0, errors.New("ungültiges Resultat für Passwort Zurücksetzen")
	}

	return id, nil
}

func (s *DatabaseService) DeleteResetPassword(email string) error {
	query, err := sqlQueries.ReadFile("queries/delete_reset_password.sql")
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(email)
	if err != nil {
		return err
	}

	return nil
}
