package db_adapter

import (
	"liquiswiss/pkg/models"
)

func (d *DatabaseAdapter) ListMembers(organisationID int64) ([]models.OrganisationMember, error) {
	members := []models.OrganisationMember{}

	query, err := sqlQueries.ReadFile("queries/list_members.sql")
	if err != nil {
		return nil, err
	}

	rows, err := d.db.Query(string(query), organisationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var member models.OrganisationMember

		err := rows.Scan(
			&member.UserID,
			&member.Name,
			&member.Email,
			&member.Role,
			&member.IsDefault,
		)
		if err != nil {
			return nil, err
		}

		members = append(members, member)
	}

	return members, nil
}

func (d *DatabaseAdapter) GetMember(organisationID int64, userID int64) (*models.OrganisationMember, error) {
	var member models.OrganisationMember

	query, err := sqlQueries.ReadFile("queries/get_member.sql")
	if err != nil {
		return nil, err
	}

	err = d.db.QueryRow(string(query), organisationID, userID).Scan(
		&member.UserID,
		&member.Name,
		&member.Email,
		&member.Role,
		&member.IsDefault,
	)
	if err != nil {
		return nil, err
	}

	return &member, nil
}

func (d *DatabaseAdapter) UpdateMemberRole(organisationID int64, userID int64, role string) error {
	query, err := sqlQueries.ReadFile("queries/update_member_role.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(role, organisationID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (d *DatabaseAdapter) DeleteMember(organisationID int64, userID int64) error {
	query, err := sqlQueries.ReadFile("queries/delete_member.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(organisationID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (d *DatabaseAdapter) CountOwners(organisationID int64) (int64, error) {
	var count int64

	query, err := sqlQueries.ReadFile("queries/count_owners.sql")
	if err != nil {
		return 0, err
	}

	err = d.db.QueryRow(string(query), organisationID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (d *DatabaseAdapter) GetMemberPermission(userID int64, organisationID int64) (*models.MemberPermission, error) {
	var permission models.MemberPermission

	query, err := sqlQueries.ReadFile("queries/get_member_permission.sql")
	if err != nil {
		return nil, err
	}

	err = d.db.QueryRow(string(query), userID, organisationID).Scan(
		&permission.ID,
		&permission.UserID,
		&permission.OrganisationID,
		&permission.EntityType,
		&permission.CanView,
		&permission.CanEdit,
		&permission.CanDelete,
		&permission.CreatedAt,
		&permission.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &permission, nil
}

func (d *DatabaseAdapter) UpsertMemberPermission(userID int64, organisationID int64, canView bool, canEdit bool, canDelete bool) error {
	query, err := sqlQueries.ReadFile("queries/upsert_member_permission.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, organisationID, canView, canEdit, canDelete)
	if err != nil {
		return err
	}

	return nil
}

func (d *DatabaseAdapter) DeleteMemberPermissions(userID int64, organisationID int64) error {
	query, err := sqlQueries.ReadFile("queries/delete_member_permissions.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, organisationID)
	if err != nil {
		return err
	}

	return nil
}
