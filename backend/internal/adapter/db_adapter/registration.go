package db_adapter

import (
	"errors"
	"time"
)

func (d *DatabaseAdapter) CreateRegistration(email, code string) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_registration.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(email, code, email)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted registration
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (d *DatabaseAdapter) DeleteRegistration(registrationID int64, email string) error {
	query, err := sqlQueries.ReadFile("queries/delete_registration.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(registrationID, email)
	if err != nil {
		return err
	}

	return nil
}

func (d *DatabaseAdapter) ValidateRegistration(email, code string, validity time.Duration) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/validate_registration.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var id int64
	err = stmt.QueryRow(email, code, validity.Hours()).Scan(&id)
	if err != nil {
		return 0, err
	}
	if id <= 0 {
		return 0, errors.New("ungültiges Resultat für Registrierung")
	}

	return id, nil
}
