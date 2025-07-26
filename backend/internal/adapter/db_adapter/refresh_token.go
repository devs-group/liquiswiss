package db_adapter

import "time"

// StoreRefreshTokenID stores the refresh token's token ID, user ID, device name and expiration time in the database
func (d *DatabaseAdapter) StoreRefreshTokenID(userID int64, tokenId string, expirationTime time.Time, deviceName string) error {
	query, err := sqlQueries.ReadFile("queries/create_refresh_token.sql")
	if err != nil {
		return err
	}

	stmt, err := d.db.Prepare(string(query))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, tokenId, expirationTime, deviceName)
	if err != nil {
		return err
	}

	return nil
}

func (d *DatabaseAdapter) CheckRefreshToken(tokenID string, userID int64) (bool, error) {
	query, err := sqlQueries.ReadFile("queries/get_refresh_token.sql")
	if err != nil {
		return false, err
	}

	var exists bool
	err = d.db.QueryRow(string(query), tokenID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (d *DatabaseAdapter) DeleteRefreshToken(tokenID string, userID int64) error {
	query, err := sqlQueries.ReadFile("queries/delete_refresh_token.sql")
	if err != nil {
		return err
	}

	_, err = d.db.Exec(string(query), tokenID, userID)

	return err
}
