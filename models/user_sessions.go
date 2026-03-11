package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/devmarvs/bblog/db"
)

func IsTokenCurrentForUser(userID int64, issuedAt time.Time) (bool, error) {
	if userID <= 0 || issuedAt.IsZero() {
		return false, nil
	}

	const query = `
		SELECT COALESCE(token_valid_after, created_ts, NOW())
		FROM bblog.users
		WHERE user_id = $1
			AND is_active IS TRUE
			AND is_deleted IS FALSE
	`

	var tokenValidAfter time.Time
	if err := db.DB.QueryRow(query, userID).Scan(&tokenValidAfter); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return issuedAt.UTC().Unix() >= tokenValidAfter.UTC().Unix(), nil
}
