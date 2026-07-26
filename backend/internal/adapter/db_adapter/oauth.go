package db_adapter

import (
	"database/sql"
	"encoding/json"
	"errors"
	"liquiswiss/pkg/models"
	"time"
)

func (d *DatabaseAdapter) CreateOAuthClient(clientID, clientName string, redirectURIs []string) error {
	query, err := sqlQueries.ReadFile("queries/create_oauth_client.sql")
	if err != nil {
		return err
	}

	uris, err := json.Marshal(redirectURIs)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(string(query), clientID, clientName, string(uris))

	return err
}

func (d *DatabaseAdapter) GetOAuthClient(clientID string) (*models.OAuthClient, error) {
	query, err := sqlQueries.ReadFile("queries/get_oauth_client.sql")
	if err != nil {
		return nil, err
	}

	var client models.OAuthClient
	var uris string
	err = d.db.QueryRow(string(query), clientID).Scan(&client.ClientID, &client.ClientName, &uris)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal([]byte(uris), &client.RedirectURIs); err != nil {
		return nil, err
	}

	return &client, nil
}

func (d *DatabaseAdapter) CreateOAuthAuthCode(code models.OAuthAuthCode) error {
	query, err := sqlQueries.ReadFile("queries/create_oauth_auth_code.sql")
	if err != nil {
		return err
	}

	_, err = d.db.Exec(string(query), code.CodeHash, code.ClientID, code.UserID, code.CodeChallenge, code.RedirectURI, code.Resource, code.ExpiresAt)

	return err
}

func (d *DatabaseAdapter) GetOAuthAuthCode(codeHash string) (*models.OAuthAuthCode, error) {
	query, err := sqlQueries.ReadFile("queries/get_oauth_auth_code.sql")
	if err != nil {
		return nil, err
	}

	var code models.OAuthAuthCode
	err = d.db.QueryRow(string(query), codeHash).Scan(
		&code.CodeHash, &code.ClientID, &code.UserID, &code.CodeChallenge,
		&code.RedirectURI, &code.Resource, &code.ExpiresAt, &code.UsedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &code, nil
}

// MarkOAuthAuthCodeUsed marks the code as used and reports whether this call was the one consuming it,
// so concurrent reuse attempts can be detected atomically
func (d *DatabaseAdapter) MarkOAuthAuthCodeUsed(codeHash string) (bool, error) {
	query, err := sqlQueries.ReadFile("queries/mark_oauth_auth_code_used.sql")
	if err != nil {
		return false, err
	}

	res, err := d.db.Exec(string(query), codeHash)
	if err != nil {
		return false, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return affected == 1, nil
}

func (d *DatabaseAdapter) CreateOAuthRefreshToken(token models.OAuthRefreshToken) error {
	query, err := sqlQueries.ReadFile("queries/create_oauth_refresh_token.sql")
	if err != nil {
		return err
	}

	_, err = d.db.Exec(string(query), token.TokenHash, token.ClientID, token.UserID, token.ExpiresAt, token.RotatedFrom)

	return err
}

func (d *DatabaseAdapter) GetOAuthRefreshToken(tokenHash string) (*models.OAuthRefreshToken, error) {
	query, err := sqlQueries.ReadFile("queries/get_oauth_refresh_token.sql")
	if err != nil {
		return nil, err
	}

	var token models.OAuthRefreshToken
	err = d.db.QueryRow(string(query), tokenHash).Scan(
		&token.TokenHash, &token.ClientID, &token.UserID, &token.ExpiresAt, &token.RevokedAt, &token.RotatedFrom,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &token, nil
}

func (d *DatabaseAdapter) RevokeOAuthRefreshToken(tokenHash string) error {
	query, err := sqlQueries.ReadFile("queries/revoke_oauth_refresh_token.sql")
	if err != nil {
		return err
	}

	_, err = d.db.Exec(string(query), tokenHash)

	return err
}

func (d *DatabaseAdapter) RevokeOAuthConnection(userID int64, clientID string) error {
	query, err := sqlQueries.ReadFile("queries/revoke_oauth_refresh_tokens_for_connection.sql")
	if err != nil {
		return err
	}

	_, err = d.db.Exec(string(query), userID, clientID)

	return err
}

func (d *DatabaseAdapter) ListOAuthConnections(userID int64) ([]models.OAuthConnection, error) {
	query, err := sqlQueries.ReadFile("queries/list_oauth_connections.sql")
	if err != nil {
		return nil, err
	}

	rows, err := d.db.Query(string(query), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	connections := []models.OAuthConnection{}
	for rows.Next() {
		var conn models.OAuthConnection
		var createdAt, lastUsedAt time.Time
		if err := rows.Scan(&conn.ClientID, &conn.ClientName, &createdAt, &lastUsedAt); err != nil {
			return nil, err
		}
		conn.CreatedAt = createdAt
		conn.LastUsedAt = lastUsedAt
		connections = append(connections, conn)
	}

	return connections, rows.Err()
}

// HasActiveOAuthConnection reports whether the user still has an unrevoked
// refresh token for the client, making connection revocation immediate even
// for outstanding access tokens
func (d *DatabaseAdapter) HasActiveOAuthConnection(userID int64, clientID string) (bool, error) {
	query, err := sqlQueries.ReadFile("queries/has_active_oauth_connection.sql")
	if err != nil {
		return false, err
	}

	var active bool
	err = d.db.QueryRow(string(query), userID, clientID).Scan(&active)

	return active, err
}
