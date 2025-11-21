package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/devmarvs/bblog/db"
)

var (
	ErrVerificationTokenInvalid = errors.New("invalid verification token")
	ErrVerificationTokenExpired = errors.New("verification token expired")
	ErrVerificationAlreadyUsed  = errors.New("verification token already used")
)

func CreateEmailVerification(userID int64, token string, expiresAt time.Time) error {
	const query = `
        INSERT INTO bblog.email_verifications(user_id, token_hash, expires_at)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id)
        DO UPDATE SET
            token_hash = EXCLUDED.token_hash,
            expires_at = EXCLUDED.expires_at,
            consumed_at = NULL,
            created_ts = NOW()
    `

	_, err := db.DB.Exec(query, userID, hashToken(token), expiresAt)
	return err
}

func VerifyEmailToken(token string) (int64, error) {
	if strings.TrimSpace(token) == "" {
		return 0, ErrVerificationTokenInvalid
	}

	const selectQuery = `
        SELECT user_id, expires_at, consumed_at
        FROM bblog.email_verifications
        WHERE token_hash = $1
    `

	var (
		userID    int64
		expiresAt time.Time
		consumed  sql.NullTime
	)

	err := db.DB.QueryRow(selectQuery, hashToken(token)).Scan(&userID, &expiresAt, &consumed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrVerificationTokenInvalid
		}
		return 0, err
	}

	if consumed.Valid {
		return 0, ErrVerificationAlreadyUsed
	}

	if expiresAt.Before(time.Now()) {
		return 0, ErrVerificationTokenExpired
	}

	if err := markVerificationUsed(userID); err != nil {
		return 0, err
	}

	if err := activateUser(userID); err != nil {
		return 0, err
	}

	return userID, nil
}

func markVerificationUsed(userID int64) error {
	const query = `
        UPDATE bblog.email_verifications
        SET consumed_at = NOW()
        WHERE user_id = $1
    `

	_, err := db.DB.Exec(query, userID)
	return err
}

func activateUser(userID int64) error {
	const query = `
        UPDATE bblog.users
        SET is_active = TRUE,
            updated_ts = NOW()
        WHERE user_id = $1
    `

	_, err := db.DB.Exec(query, userID)
	return err
}
