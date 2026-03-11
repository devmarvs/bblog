package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/devmarvs/bblog/db"
)

var (
	ErrVerificationCodeInvalid     = errors.New("invalid verification code")
	ErrVerificationCodeExpired     = errors.New("verification code expired")
	ErrVerificationAlreadyUsed     = errors.New("verification code already used")
	ErrVerificationTooManyAttempts = errors.New("too many verification attempts")
)

const verificationMaxAttempts = 5

func CreateEmailVerification(userID int64, code string, expiresAt time.Time) error {
	sanitizedCode := strings.TrimSpace(code)
	if userID <= 0 || sanitizedCode == "" {
		return ErrVerificationCodeInvalid
	}

	const query = `
        INSERT INTO bblog.email_verifications(user_id, token_hash, expires_at, attempt_count, last_sent_at)
        VALUES ($1, $2, $3, 0, NOW())
        ON CONFLICT (user_id)
        DO UPDATE SET
            token_hash = EXCLUDED.token_hash,
            expires_at = EXCLUDED.expires_at,
            consumed_at = NULL,
            attempt_count = 0,
            last_sent_at = NOW(),
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
        SELECT token_hash, expires_at, consumed_at, attempt_count
        FROM bblog.email_verifications
        WHERE user_id = $1
    `

	var (
		tokenHash string
		expiresAt time.Time
		consumed  sql.NullTime
		attempts  int
	)

	err := db.DB.QueryRow(selectQuery, userID).Scan(&tokenHash, &expiresAt, &consumed, &attempts)
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

	if attempts >= verificationMaxAttempts {
		return ErrVerificationTooManyAttempts
	}

	if tokenHash != hashToken(sanitizedCode) {
		updatedAttempts, err := incrementVerificationAttempts(userID)
		if err != nil {
			return err
		}

		if updatedAttempts >= verificationMaxAttempts {
			return ErrVerificationTooManyAttempts
		}

		return ErrVerificationCodeInvalid
	}

	if err := markVerificationUsed(userID); err != nil {
		return err
	}

	return activateUser(userID)
}

func CanSendEmailVerification(userID int64, minInterval time.Duration) (bool, error) {
	if userID <= 0 {
		return false, ErrVerificationCodeInvalid
	}

	const query = `
		SELECT last_sent_at
		FROM bblog.email_verifications
		WHERE user_id = $1
	`

	var lastSent sql.NullTime
	err := db.DB.QueryRow(query, userID).Scan(&lastSent)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true, nil
		}
		return false, err
	}

	if !lastSent.Valid {
		return true, nil
	}

	return time.Since(lastSent.Time) >= minInterval, nil
}

func incrementVerificationAttempts(userID int64) (int, error) {
	const query = `
		UPDATE bblog.email_verifications
		SET attempt_count = attempt_count + 1
		WHERE user_id = $1
		RETURNING attempt_count
	`

	var attempts int
	if err := db.DB.QueryRow(query, userID).Scan(&attempts); err != nil {
		return 0, err
	}

	return attempts, nil
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
