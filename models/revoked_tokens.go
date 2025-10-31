package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/devmarvs/bblog/db"
)

func RevokeToken(token string, expiresAt time.Time) error {
	if token == "" {
		return errors.New("missing token")
	}

	const query = `
		INSERT INTO bblog.revoked_tokens (token_hash, expires_at)
		VALUES ($1, $2)
		ON CONFLICT (token_hash) DO UPDATE SET expires_at = EXCLUDED.expires_at
	`

	_, err := db.DB.Exec(query, hashToken(token), expiresAt)
	return err
}

func IsTokenRevoked(token string) (bool, error) {
	if token == "" {
		return false, nil
	}

	const query = `
		SELECT 1
		FROM bblog.revoked_tokens
		WHERE token_hash = $1
			AND expires_at > NOW()
	`

	row := db.DB.QueryRow(query, hashToken(token))
	var exists int
	if err := row.Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
