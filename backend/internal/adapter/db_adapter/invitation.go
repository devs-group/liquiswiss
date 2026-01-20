package db_adapter

import (
	"liquiswiss/pkg/models"
	"time"
)

func (d *DatabaseAdapter) CreateInvitation(organisationID int64, email string, role string, token string, invitedBy int64, expiresAt time.Time) (int64, error) {
	query, err := sqlQueries.ReadFile("queries/create_invitation.sql")
	if err != nil {
		return 0, err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(organisationID, email, role, token, invitedBy, expiresAt)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (d *DatabaseAdapter) ListInvitations(organisationID int64) ([]models.Invitation, error) {
	invitations := []models.Invitation{}

	query, err := sqlQueries.ReadFile("queries/list_invitations.sql")
	if err != nil {
		return nil, err
	}

	rows, err := d.db.Query(string(query), organisationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var invitation models.Invitation

		err := rows.Scan(
			&invitation.ID,
			&invitation.OrganisationID,
			&invitation.Email,
			&invitation.Role,
			&invitation.Token,
			&invitation.InvitedBy,
			&invitation.InvitedByName,
			&invitation.ExpiresAt,
			&invitation.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		invitations = append(invitations, invitation)
	}

	return invitations, nil
}

func (d *DatabaseAdapter) GetInvitationByID(organisationID int64, invitationID int64) (*models.Invitation, error) {
	var invitation models.Invitation

	query, err := sqlQueries.ReadFile("queries/get_invitation_by_id.sql")
	if err != nil {
		return nil, err
	}

	err = d.db.QueryRow(string(query), invitationID, organisationID).Scan(
		&invitation.ID,
		&invitation.OrganisationID,
		&invitation.Email,
		&invitation.Role,
		&invitation.Token,
		&invitation.InvitedBy,
		&invitation.InvitedByName,
		&invitation.ExpiresAt,
		&invitation.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &invitation, nil
}

func (d *DatabaseAdapter) GetInvitationByToken(token string) (*models.Invitation, error) {
	var invitation models.Invitation

	query, err := sqlQueries.ReadFile("queries/get_invitation_by_token.sql")
	if err != nil {
		return nil, err
	}

	err = d.db.QueryRow(string(query), token).Scan(
		&invitation.ID,
		&invitation.OrganisationID,
		&invitation.Email,
		&invitation.Role,
		&invitation.Token,
		&invitation.InvitedBy,
		&invitation.InvitedByName,
		&invitation.ExpiresAt,
		&invitation.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &invitation, nil
}

func (d *DatabaseAdapter) DeleteInvitation(organisationID int64, invitationID int64) error {
	query, err := sqlQueries.ReadFile("queries/delete_invitation.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(invitationID, organisationID)
	if err != nil {
		return err
	}

	return nil
}

func (d *DatabaseAdapter) DeleteInvitationByToken(token string) error {
	query, err := sqlQueries.ReadFile("queries/delete_invitation_by_token.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(token)
	if err != nil {
		return err
	}

	return nil
}

func (d *DatabaseAdapter) GetOrganisationName(organisationID int64) (string, error) {
	var name string

	query, err := sqlQueries.ReadFile("queries/get_organisation_name.sql")
	if err != nil {
		return "", err
	}

	err = d.db.QueryRow(string(query), organisationID).Scan(&name)
	if err != nil {
		return "", err
	}

	return name, nil
}

func (d *DatabaseAdapter) GetUserIDByEmail(email string) (int64, error) {
	var userID int64

	query, err := sqlQueries.ReadFile("queries/get_user_id_by_email.sql")
	if err != nil {
		return 0, err
	}

	err = d.db.QueryRow(string(query), email).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (d *DatabaseAdapter) CheckUserInOrganisation(userID int64, organisationID int64) (bool, error) {
	var exists bool

	query, err := sqlQueries.ReadFile("queries/check_user_in_organisation.sql")
	if err != nil {
		return false, err
	}

	err = d.db.QueryRow(string(query), userID, organisationID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
