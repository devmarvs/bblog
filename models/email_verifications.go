package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/devmarvs/bblog/db"
)

var (
	ErrVerificationCodeInvalid = errors.New("invalid verification code")
	ErrVerificationCodeExpired = errors.New("verification code expired")
	ErrVerificationAlreadyUsed = errors.New("verification code already used")
)

func CreateEmailVerification(userID int64, code string, expiresAt time.Time) error {
	sanitizedCode := strings.TrimSpace(code)
	if userID <= 0 || sanitizedCode == "" {
		return ErrVerificationCodeInvalid
	}

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

	_, err := db.DB.Exec(query, userID, hashToken(sanitizedCode), expiresAt)
	return err
}

func VerifyEmailCode(userID int64, code string) error {
	sanitizedCode := strings.TrimSpace(code)
	if userID <= 0 || sanitizedCode == "" {
		return ErrVerificationCodeInvalid
	}

	const selectQuery = `
        SELECT token_hash, expires_at, consumed_at
        FROM bblog.email_verifications
        WHERE user_id = $1
    `

	var (
		tokenHash string
		expiresAt time.Time
		consumed  sql.NullTime
	)

	err := db.DB.QueryRow(selectQuery, userID).Scan(&tokenHash, &expiresAt, &consumed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrVerificationCodeInvalid
		}
		return err
	}

	if consumed.Valid {
		return ErrVerificationAlreadyUsed
	}

	if expiresAt.Before(time.Now()) {
		return ErrVerificationCodeExpired
	}

	if tokenHash != hashToken(sanitizedCode) {
		return ErrVerificationCodeInvalid
	}

	if err := markVerificationUsed(userID); err != nil {
		return err
	}

	return activateUser(userID)
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
