package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/devmarvs/bblog/db"
	"github.com/devmarvs/bblog/utils"
)

var (
	ErrPasswordResetTokenInvalid = errors.New("invalid password reset token")
	ErrPasswordResetTokenExpired = errors.New("password reset token expired")
	ErrPasswordResetTokenUsed    = errors.New("password reset token already used")
)

func CreatePasswordReset(userID int64, token string, expiresAt time.Time) error {
	const query = `
        INSERT INTO bblog.password_resets(user_id, token_hash, expires_at)
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

func ResetPassword(token string, newPassword string) error {
	sanitizedToken := strings.TrimSpace(token)
	if sanitizedToken == "" {
		return ErrPasswordResetTokenInvalid
	}

	const selectQuery = `
        SELECT user_id, expires_at, consumed_at
        FROM bblog.password_resets
        WHERE token_hash = $1
    `

	var (
		userID    int64
		expiresAt time.Time
		consumed  sql.NullTime
	)

	err := db.DB.QueryRow(selectQuery, hashToken(sanitizedToken)).Scan(&userID, &expiresAt, &consumed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPasswordResetTokenInvalid
		}
		return err
	}

	if consumed.Valid {
		return ErrPasswordResetTokenUsed
	}

	if expiresAt.Before(time.Now()) {
		return ErrPasswordResetTokenExpired
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const updateUserPassword = `
        UPDATE bblog.users
        SET password = $1,
            token_valid_after = NOW(),
            updated_ts = NOW()
        WHERE user_id = $2
    `

	if _, err := tx.Exec(updateUserPassword, hashedPassword, userID); err != nil {
		return err
	}

	const markConsumed = `
        UPDATE bblog.password_resets
        SET consumed_at = NOW()
        WHERE user_id = $1
    `

	if _, err := tx.Exec(markConsumed, userID); err != nil {
		return err
	}

	return tx.Commit()
}
