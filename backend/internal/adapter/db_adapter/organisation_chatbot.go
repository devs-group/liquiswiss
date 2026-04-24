package db_adapter

import "liquiswiss/pkg/models"

func (d *DatabaseAdapter) CreateOrganisationChatbot(organisationID int64, chatbotID string, skillID *string) error {
	query, err := sqlQueries.ReadFile("queries/create_organisation_chatbot.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(organisationID, chatbotID, skillID)
	return err
}

func (d *DatabaseAdapter) HasOrganisationChatbots(organisationID int64) (bool, error) {
	query, err := sqlQueries.ReadFile("queries/has_organisation_chatbots.sql")
	if err != nil {
		return false, err
	}

	var exists bool
	err = d.db.QueryRow(string(query), organisationID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (d *DatabaseAdapter) ListOrganisationChatbots(organisationID int64) ([]models.OrganisationChatbot, error) {
	query, err := sqlQueries.ReadFile("queries/list_organisation_chatbots.sql")
	if err != nil {
		return nil, err
	}

	rows, err := d.db.Query(string(query), organisationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chatbots []models.OrganisationChatbot
	for rows.Next() {
		var c models.OrganisationChatbot
		if err := rows.Scan(&c.ID, &c.OrganisationID, &c.ChatbotID, &c.SkillID); err != nil {
			return nil, err
		}
		chatbots = append(chatbots, c)
	}
	return chatbots, rows.Err()
}

func (d *DatabaseAdapter) DeleteOrganisationChatbots(organisationID int64) error {
	query, err := sqlQueries.ReadFile("queries/delete_organisation_chatbots.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(organisationID)
	return err
}

func (d *DatabaseAdapter) GetOrganisationChatbot(organisationID int64) (string, error) {
	query, err := sqlQueries.ReadFile("queries/get_organisation_chatbot.sql")
	if err != nil {
		return "", err
	}

	var chatbotID string
	err = d.db.QueryRow(string(query), organisationID).Scan(&chatbotID)
	if err != nil {
		return "", err
	}

	return chatbotID, nil
}
