package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Import the pq driver
)

var DB *sql.DB

func InitDb() {

	var err error
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PWORD")
	dbName := os.Getenv("DB_NAME")
	connString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost,
		dbPort,
		dbUser,
		dbPass,
		dbName,
		databaseSSLMode(dbHost),
	)

	// Open connection to PostgreSQL
	DB, err = sql.Open("postgres", connString)
	if err != nil {
		// log.Fatal("Error connecting to database:", err)
		panic("Could not connect to database.")
	}

	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)

	createSchema()
	createTables()

}

func databaseSSLMode(host string) string {
	if sslMode := strings.TrimSpace(os.Getenv("DB_SSLMODE")); sslMode != "" {
		return sslMode
	}

	normalizedHost := strings.ToLower(strings.TrimSpace(host))
	switch {
	case normalizedHost == "", normalizedHost == "localhost", normalizedHost == "::1":
		return "disable"
	case strings.HasPrefix(normalizedHost, "127."):
		return "disable"
	default:
		return "require"
	}
}

func createSchema() {
	const createSchema = `
		CREATE SCHEMA IF NOT EXISTS bblog;
	`

	if _, err := DB.Exec(createSchema); err != nil {
		log.Fatalf("Error ensuring schema exists: %v", err)
	}
}

func createTables() {

	createUsersTable := `
		CREATE TABLE IF NOT EXISTS bblog.users (
			user_id BIGSERIAL PRIMARY KEY,
			created_ts TIMESTAMPTZ NULL DEFAULT NOW(),
			updated_ts TIMESTAMPTZ NULL,
			token_valid_after TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			user_type_id INTEGER NULL,
			username VARCHAR NULL,
			password VARCHAR NOT NULL,
			email VARCHAR NULL UNIQUE,
			mobile VARCHAR NULL UNIQUE,
			country_code VARCHAR NULL,
			is_online BOOL NOT NULL DEFAULT FALSE,
			is_active BOOL NOT NULL DEFAULT FALSE,
			is_deleted BOOL NOT NULL DEFAULT FALSE,
			is_premium BOOL NOT NULL DEFAULT FALSE
		);
		CREATE INDEX IF NOT EXISTS bb_user_idx ON bblog.users USING btree(user_id,created_ts, updated_ts,email,mobile);
	`

	_, err := DB.Exec(createUsersTable)

	if err != nil {
		log.Fatalf("Error creating schema: %v", err) // Log the real error

	}

	ensureUsersTokenValidityColumn := `
		ALTER TABLE bblog.users
		ADD COLUMN IF NOT EXISTS token_valid_after TIMESTAMPTZ;

		UPDATE bblog.users
		SET token_valid_after = COALESCE(token_valid_after, created_ts, NOW())
		WHERE token_valid_after IS NULL;

		ALTER TABLE bblog.users
		ALTER COLUMN token_valid_after SET DEFAULT NOW();

		ALTER TABLE bblog.users
		ALTER COLUMN token_valid_after SET NOT NULL;
	`

	if _, err = DB.Exec(ensureUsersTokenValidityColumn); err != nil {
		log.Fatalf("Error ensuring token validity column: %v", err)
	}

	// if err != nil {
	// 	panic("Could not create users table")
	// }

	createUserTypeTable := `
		CREATE TABLE IF NOT EXISTS bblog.user_type (
			user_type_id SERIAL PRIMARY KEY,
			description VARCHAR UNIQUE,
			created_ts TIMESTAMPTZ DEFAULT NOW(),
			updated_ts TIMESTAMPTZ NULL
		)
	`

	_, err = DB.Exec(createUserTypeTable)

	if err != nil {
		log.Fatalf("Error creating user_type table: %v", err)
	}

	seedUserTypeTable := `
		INSERT INTO bblog.user_type(description) VALUES 
		('user'),('baby'),('pet')
		ON CONFLICT (description) DO NOTHING
	`
	_, err = DB.Exec(seedUserTypeTable)
	if err != nil { //DO NOTHING
		// log.Fatalf("Error inserting into user_type table: %v", err)
	}

	createSubUsersTable := `
		CREATE TABLE IF NOT EXISTS bblog.sub_users (
			sub_user_id BIGSERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			user_type_id SMALLINT NOT NULL,
			is_active BOOL NOT NULL DEFAULT true,
			is_deleted BOOL NOT NULL DEFAULT false,
			created_ts TIMESTAMPTZ DEFAULT NOW(),
			updated_ts TIMESTAMPTZ,
			name VARCHAR NOT NULL,
			CONSTRAINT user_type_id_fkey FOREIGN KEY(user_type_id) REFERENCES bblog.user_type(user_type_id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS sub_user_id_namex ON bblog.sub_users USING btree(sub_user_id, user_id, name)
	`

	_, err = DB.Exec(createSubUsersTable)

	if err != nil {
		log.Fatalf("Error creating sub_users table: %v", err)
	}

	createLogTypesTable := `
		CREATE TABLE IF NOT EXISTS bblog.log_types (
			log_type_id SERIAL PRIMARY KEY,
			log_name VARCHAR UNIQUE NOT NULL,
			created_ts TIMESTAMPTZ DEFAULT NOW(),
			updated_ts TIMESTAMPTZ NULL
		)
	`

	_, err = DB.Exec(createLogTypesTable)

	if err != nil {
		log.Fatalf("Error creating log type table: %v", err)
	}

	seedLogTypeTable := `
		INSERT INTO bblog.log_types(log_name) VALUES
		('milk'),
		('medicine'),
		('vaccine'),
		('diaper'),
		('vitamins'),
		('poop'),
		('pee'),
		('temperature'),
		('height'),
		('weight'),
		('solid food'),
		('snack'),
		('meal'),
		('cough'),
		('vomit'),
		('rashes'),
		('injury'),
		('bath'),
		('hospital'),
		('veterinary'),
		('walks'),
		('others')
		ON CONFLICT (log_name) DO NOTHING
	`

	_, err = DB.Exec(seedLogTypeTable)

	if err != nil { //DO NOTHING
		// log.Fatalf("Error inserting into log type table: %v", err)
	}

	createLogTable := `
		CREATE TABLE IF NOT EXISTS bblog.user_log(
			user_log_id BIGSERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			sub_user_id INTEGER NOT NULL,
			log_type_id INTEGER NOT NULL,
			log_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			log_description TEXT NULL,
			created_ts TIMESTAMPTZ DEFAULT NOW(),
			updated_ts TIMESTAMPTZ NULL,
			CONSTRAINT user_id_fkey FOREIGN KEY(user_id) REFERENCES bblog.users(user_id) ON DELETE CASCADE,
			CONSTRAINT sub_user_id_fkey FOREIGN KEY(sub_user_id) REFERENCES bblog.sub_users(sub_user_id) ON DELETE CASCADE,
			CONSTRAINT log_type_id_fkey FOREIGN KEY(log_type_id) REFERENCES bblog.log_types(log_type_id)
		);
		CREATE INDEX IF NOT EXISTS user_log_user_idx ON bblog.user_log (user_id, log_type_id);
		CREATE INDEX IF NOT EXISTS user_log_sub_idx ON bblog.user_log (sub_user_id);

	`
	_, err = DB.Exec(createLogTable)

	if err != nil {
		log.Fatalf("Error creating log table: %v", err)
	}

	createRevokedTokensTable := `
		CREATE TABLE IF NOT EXISTS bblog.revoked_tokens(
			token_hash VARCHAR PRIMARY KEY,
			expires_at TIMESTAMPTZ NOT NULL
		)
	`

	if _, err = DB.Exec(createRevokedTokensTable); err != nil {
		log.Fatalf("Error creating revoked tokens table: %v", err)
	}

	createEmailVerificationTable := `
		CREATE TABLE IF NOT EXISTS bblog.email_verifications (
			user_id BIGINT PRIMARY KEY REFERENCES bblog.users(user_id) ON DELETE CASCADE,
			token_hash VARCHAR NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			consumed_at TIMESTAMPTZ NULL,
			attempt_count INTEGER NOT NULL DEFAULT 0,
			last_sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			created_ts TIMESTAMPTZ DEFAULT NOW()
		);
	`

	if _, err = DB.Exec(createEmailVerificationTable); err != nil {
		log.Fatalf("Error creating email verification table: %v", err)
	}

	ensureEmailVerificationSecurityColumns := `
		ALTER TABLE bblog.email_verifications
		ADD COLUMN IF NOT EXISTS attempt_count INTEGER NOT NULL DEFAULT 0;

		ALTER TABLE bblog.email_verifications
		ADD COLUMN IF NOT EXISTS last_sent_at TIMESTAMPTZ;

		UPDATE bblog.email_verifications
		SET last_sent_at = COALESCE(last_sent_at, created_ts, NOW())
		WHERE last_sent_at IS NULL;

		ALTER TABLE bblog.email_verifications
		ALTER COLUMN last_sent_at SET DEFAULT NOW();

		ALTER TABLE bblog.email_verifications
		ALTER COLUMN last_sent_at SET NOT NULL;
	`

	if _, err = DB.Exec(ensureEmailVerificationSecurityColumns); err != nil {
		log.Fatalf("Error ensuring email verification security columns: %v", err)
	}

	createPasswordResetTable := `
		CREATE TABLE IF NOT EXISTS bblog.password_resets (
			user_id BIGINT PRIMARY KEY REFERENCES bblog.users(user_id) ON DELETE CASCADE,
			token_hash VARCHAR NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			consumed_at TIMESTAMPTZ NULL,
			created_ts TIMESTAMPTZ DEFAULT NOW()
		);
	`

	if _, err = DB.Exec(createPasswordResetTable); err != nil {
		log.Fatalf("Error creating password reset table: %v", err)
	}

	createAppVersionsTable := `
		CREATE TABLE IF NOT EXISTS bblog.app_versions (
			version_id BIGSERIAL PRIMARY KEY,
			api_version VARCHAR NOT NULL,
			mobile_version VARCHAR NOT NULL,
			created_ts TIMESTAMPTZ DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS app_versions_created_idx ON bblog.app_versions (created_ts DESC, version_id DESC);
	`

	if _, err = DB.Exec(createAppVersionsTable); err != nil {
		log.Fatalf("Error creating app_versions table: %v", err)
	}
}
