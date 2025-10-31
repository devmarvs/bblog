package utils

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var (
	secretKeyOnce sync.Once
	secretKey     string
	secretKeyErr  error
)

func GenerateToken(email string, userId int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":  email,
		"userId": userId,
		"exp":    time.Now().Add(time.Hour * 2).Unix(),
	})

	key, err := getSecretKey()
	if err != nil {
		return "", err
	}

	return token.SignedString([]byte(key))
}

func VerifyToken(token string) (int64, error) {
	sanitizedToken := sanitizeToken(token)
	if sanitizedToken == "" {
		return 0, errors.New("Invalid token.")
	}

	key, err := getSecretKey()
	if err != nil {
		return 0, err
	}

	parsedToken, err := jwt.Parse(
		sanitizedToken,
		func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(key), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)

	if err != nil || !parsedToken.Valid {
		return 0, errors.New("Invalid token.")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("Invalid token claims.")
	}

	expiration, err := claims.GetExpirationTime()
	if err != nil || expiration == nil {
		return 0, errors.New("Invalid token claims.")
	}

	if time.Now().After(expiration.Time) {
		return 0, errors.New("Invalid token.")
	}

	userIDClaim, ok := claims["userId"]
	if !ok {
		return 0, errors.New("Invalid token claims.")
	}

	userId, ok := toInt64(userIDClaim)
	if !ok || userId <= 0 {
		return 0, errors.New("Invalid token claims.")
	}

	return userId, nil
}

func getSecretKey() (string, error) {
	secretKeyOnce.Do(func() {
		_ = godotenv.Load()
		secretKey = strings.TrimSpace(os.Getenv("JWT_SECRET"))
		if secretKey == "" {
			secretKeyErr = errors.New("JWT secret not configured")
		}
	})

	return secretKey, secretKeyErr
}

func sanitizeToken(token string) string {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(strings.ToLower(trimmed), "bearer ") {
		trimmed = strings.TrimSpace(trimmed[7:])
	}

	return trimmed
}

func toInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int64:
		return v, true
	case float64:
		return int64(v), true
	case int:
		return int64(v), true
	case string:
		if v == "" {
			return 0, false
		}
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	case json.Number:
		parsed, err := v.Int64()
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}
