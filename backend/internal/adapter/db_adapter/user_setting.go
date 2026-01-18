package db_adapter

import (
	"database/sql"
	"liquiswiss/pkg/models"
	"strings"
)

func (d *DatabaseAdapter) GetUserSetting(userID int64) (*models.UserSetting, error) {
	var userSetting models.UserSetting

	query, err := sqlQueries.ReadFile("queries/get_user_setting.sql")
	if err != nil {
		return nil, err
	}

	err = d.db.QueryRow(string(query), userID).Scan(
		&userSetting.ID,
		&userSetting.UserID,
		&userSetting.SettingsTab,
		&userSetting.SkipOrganisationSwitchQuestion,
		&userSetting.CreatedAt,
		&userSetting.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &userSetting, nil
}

func (d *DatabaseAdapter) CreateUserSetting(userID int64) (int64, error) {
	query := `INSERT INTO user_settings (user_id) VALUES (?)`

	stmt, err := d.db.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(userID)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (d *DatabaseAdapter) UpdateUserSetting(payload models.UpdateUserSetting, userID int64) error {
	query := "UPDATE user_settings SET "
	queryBuild := []string{}
	args := []any{}

	if payload.SettingsTab != nil {
		queryBuild = append(queryBuild, "settings_tab = ?")
		args = append(args, *payload.SettingsTab)
	}

	if payload.SkipOrganisationSwitchQuestion != nil {
		queryBuild = append(queryBuild, "skip_organisation_switch_question = ?")
		args = append(args, *payload.SkipOrganisationSwitchQuestion)
	}

	if len(queryBuild) == 0 {
		return nil
	}

	query += strings.Join(queryBuild, ", ")
	query += " WHERE user_id = ?"
	args = append(args, userID)

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
