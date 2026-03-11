package db

import "testing"

func TestDatabaseSSLMode_Defaults(t *testing.T) {
	t.Run("local host disables SSL by default", func(t *testing.T) {
		t.Setenv("DB_SSLMODE", "")

		if got := databaseSSLMode("localhost"); got != "disable" {
			t.Fatalf("expected disable for localhost, got %q", got)
		}
	})

	t.Run("remote host requires SSL by default", func(t *testing.T) {
		t.Setenv("DB_SSLMODE", "")

		if got := databaseSSLMode("db.internal"); got != "require" {
			t.Fatalf("expected require for remote host, got %q", got)
		}
	})

	t.Run("environment override wins", func(t *testing.T) {
		t.Setenv("DB_SSLMODE", "verify-full")

		if got := databaseSSLMode("localhost"); got != "verify-full" {
			t.Fatalf("expected env override, got %q", got)
		}
	})
}
