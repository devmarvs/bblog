package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

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
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)

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

func createSchema() {
	
	createSchema := `
		CREATE SCHEMA IF NOT EXISTS bblog;
		SET search_path TO bblog, public;
		GRANT ALL ON SCHEMA bblog TO public;
		GRANT ALL ON ALL TABLES IN SCHEMA bblog TO public;
	`

	_, err := DB.Exec(createSchema)

	if err != nil {
		log.Fatalf("Error creating users table: %v", err) // Log the real error

	}
}

func createTables() {

	createSchema := `
		CREATE SCHEMA IF NOT EXISTS bblog;
	`

	_, err := DB.Exec(createSchema)

	if err != nil {
		log.Fatalf("Error creating schema: %v", err) // Log the real error

	}

	createUsersTable := `
		CREATE TABLE IF NOT EXISTS bblog.users (
			user_id BIGSERIAL PRIMARY KEY,
			created_ts TIMESTAMPTZ NULL DEFAULT NOW(),
			updated_ts TIMESTAMPTZ NULL,
			user_type_id INTEGER NULL,
			username VARCHAR NULL,
			password VARCHAR NOT NULL,
			email VARCHAR NULL UNIQUE,
			mobile VARCHAR NULL UNIQUE,
			country_code VARCHAR NULL,
			is_online BOOL NOT NULL DEFAULT FALSE,
			is_active BOOL NOT NULL DEFAULT TRUE,
			is_deleted BOOL NOT NULL DEFAULT FALSE,
			is_premium BOOL NOT NULL DEFAULT FALSE
		);
		CREATE INDEX IF NOT EXISTS bb_user_idx ON bblog.users USING btree(user_id,created_ts, updated_ts,email,mobile);
	`

	_, err = DB.Exec(createUsersTable)

	if err != nil {
		log.Fatalf("Error creating schema: %v", err) // Log the real error

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
}
