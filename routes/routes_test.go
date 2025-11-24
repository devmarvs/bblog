package routes

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/smtp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/devmarvs/bblog/db"
	"github.com/devmarvs/bblog/middlewares"
	"github.com/devmarvs/bblog/utils"
	"github.com/gin-gonic/gin"
)

func setupRouter(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)

	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	db.DB = database
	t.Cleanup(func() { _ = database.Close() })

	t.Setenv("SMTP_HOST", "smtp.test")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_USER", "user")
	t.Setenv("SMTP_PASS", "pass")
	t.Setenv("MAIL_FROM", "from@test.com")
	t.Setenv("APP_BASE_URL", "http://example.com")
	t.Setenv("JWT_SECRET", "test-secret")

	utils.SetSendMail(func(string, smtp.Auth, string, []string, []byte) error { return nil })
	t.Cleanup(func() { utils.SetSendMail(nil) })

	router := gin.New()
	router.Use(gin.Recovery())
	RegisterRoutes(router)
	return router, mock
}

func stubAuthAllow(t *testing.T) {
	middlewares.SetAuthDependencies(
		func(token string) (*utils.TokenDetails, error) {
			return &utils.TokenDetails{
				UserID:    1,
				Token:     token,
				ExpiresAt: time.Now().Add(time.Hour),
			}, nil
		},
		func(string) (bool, error) { return false, nil },
	)

	t.Cleanup(func() {
		middlewares.SetAuthDependencies(nil, nil)
	})
}

func TestCreateUser_Success(t *testing.T) {
	router, mock := setupRouter(t)

	mock.ExpectQuery(`INSERT INTO bblog\.users`).
		WithArgs(int64(1), "parent", sqlmock.AnyArg(), "parent@example.com", sqlmock.AnyArg(), "US", false).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(1)))

	mock.ExpectExec(`INSERT INTO bblog\.email_verifications`).
		WithArgs(int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	payload := `{"user_type_id":1,"username":"parent","password":"password123","email":"parent@example.com","mobile":"+123456789","country_code":"US"}`
	req := httptest.NewRequest(http.MethodPost, "/bblog/user/create", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	router, mock := setupRouter(t)

	hashed, err := utils.HashPassword("password123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	mock.ExpectQuery(`SELECT user_id, password, is_active FROM bblog\.users`).
		WithArgs("user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "password", "is_active"}).AddRow(int64(42), hashed, true))

	payload := `{"email":"user@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/bblog/login", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("bad json response: %v", err)
	}

	token, ok := resp["token"].(string)
	if !ok || token == "" {
		t.Fatalf("expected token in response, got: %v", resp)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestResendVerification_Success(t *testing.T) {
	router, mock := setupRouter(t)

	mock.ExpectQuery(`(?s)SELECT\s+user_id.*FROM\s+bblog\.users\s+WHERE email = \$1`).
		WithArgs("parent@example.com").
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "username", "created_ts", "user_type_id", "email", "mobile", "country_code", "is_online", "is_active", "is_deleted", "is_premium",
		}).AddRow(int64(1), "parent", time.Now().Format(time.RFC3339), int64(1), "parent@example.com", "+123456789", "US", false, false, false, false))

	mock.ExpectExec(`INSERT INTO bblog\.email_verifications`).
		WithArgs(int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	payload := `{"email":"parent@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/bblog/user/resend-verification", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestVerifyEmailEndpoint_Post_Success(t *testing.T) {
	router, mock := setupRouter(t)

	mock.ExpectQuery(`(?s)SELECT\s+user_id.*FROM\s+bblog\.users\s+WHERE email = \$1`).
		WithArgs("parent@example.com").
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "username", "created_ts", "user_type_id", "email", "mobile", "country_code", "is_online", "is_active", "is_deleted", "is_premium",
		}).AddRow(int64(1), "parent", time.Now().Format(time.RFC3339), int64(1), "parent@example.com", "+123456789", "US", false, false, false, false))

	mock.ExpectExec(`INSERT INTO bblog\.email_verifications`).
		WithArgs(int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	payload := `{"email":"parent@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/bblog/user/verify-email", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestVerifyEmailEndpoint_Get_Success(t *testing.T) {
	router, mock := setupRouter(t)

	mock.ExpectQuery(`(?s)SELECT\s+user_id.*FROM\s+bblog\.users\s+WHERE email = \$1`).
		WithArgs("parent@example.com").
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "username", "created_ts", "user_type_id", "email", "mobile", "country_code", "is_online", "is_active", "is_deleted", "is_premium",
		}).AddRow(int64(1), "parent", time.Now().Format(time.RFC3339), int64(1), "parent@example.com", "+123456789", "US", false, false, false, false))

	mock.ExpectExec(`INSERT INTO bblog\.email_verifications`).
		WithArgs(int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodGet, "/bblog/user/verify-email?email=parent@example.com", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestVerifyEmailEndpoint_AliasWithoutPrefix(t *testing.T) {
	router, mock := setupRouter(t)

	mock.ExpectQuery(`(?s)SELECT\s+user_id.*FROM\s+bblog\.users\s+WHERE email = \$1`).
		WithArgs("parent@example.com").
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "username", "created_ts", "user_type_id", "email", "mobile", "country_code", "is_online", "is_active", "is_deleted", "is_premium",
		}).AddRow(int64(1), "parent", time.Now().Format(time.RFC3339), int64(1), "parent@example.com", "+123456789", "US", false, false, false, false))

	mock.ExpectExec(`INSERT INTO bblog\.email_verifications`).
		WithArgs(int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodPost, "/user/verify-email", bytes.NewBufferString(`{"email":"parent@example.com"}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestForgotPassword_Success(t *testing.T) {
	router, mock := setupRouter(t)

	mock.ExpectQuery(`(?s)SELECT\s+user_id.*FROM\s+bblog\.users\s+WHERE email = \$1`).
		WithArgs("user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "username", "created_ts", "user_type_id", "email", "mobile", "country_code", "is_online", "is_active", "is_deleted", "is_premium",
		}).AddRow(int64(1), "user", time.Now().Format(time.RFC3339), int64(1), "user@example.com", "+123456789", "US", false, true, false, false))

	mock.ExpectExec(`INSERT INTO bblog\.password_resets`).
		WithArgs(int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	payload := `{"email":"user@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/bblog/user/forgot-password", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestResetPassword_Success(t *testing.T) {
	router, mock := setupRouter(t)

	mock.ExpectQuery(`SELECT user_id, expires_at, consumed_at FROM bblog\.password_resets`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at", "consumed_at"}).
			AddRow(int64(1), time.Now().Add(time.Hour), sql.NullTime{}))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE bblog\.users`).
		WithArgs(sqlmock.AnyArg(), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE bblog\.password_resets`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	payload := `{"token":"reset-token","password":"newpassword123"}`
	req := httptest.NewRequest(http.MethodPost, "/bblog/user/reset-password", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestListUserTypes_WithAuth(t *testing.T) {
	router, mock := setupRouter(t)
	stubAuthAllow(t)

	mock.ExpectQuery(`SELECT user_type_id, description FROM bblog\.user_type`).
		WillReturnRows(sqlmock.NewRows([]string{"user_type_id", "description"}).
			AddRow(int64(1), "user").
			AddRow(int64(2), "baby"))

	req := httptest.NewRequest(http.MethodGet, "/bblog/user/types", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestListLogTypes_WithAuth(t *testing.T) {
	router, mock := setupRouter(t)
	stubAuthAllow(t)

	mock.ExpectQuery(`SELECT log_type_id, log_name FROM bblog\.log_types`).
		WillReturnRows(sqlmock.NewRows([]string{"log_type_id", "log_name"}).
			AddRow(int64(1), "milk").
			AddRow(int64(2), "diaper"))

	req := httptest.NewRequest(http.MethodGet, "/bblog/log/types", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
