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

	createTables()

}

func createTables() {

	createUsersTable := `
		CREATE TABLE IF NOT EXISTS bblog.users (
			user_id BIGSERIAL PRIMARY KEY,
			created_ts TIMESTAMPTZ NULL DEFAULT NOW(),
			updated_ts TIMESTAMPTZ NULL,
			user_type_id INTEGER NULL,
			password VARCHAR NOT NULL,
			email VARCHAR NULL,
			mobile VARCHAR NULL,
			country_code VARCHAR NULL,
			is_online BOOL NOT NULL DEFAULT FALSE,
			is_active BOOL NOT NULL DEFAULT FALSE,
			is_deleted BOOL NOT NULL DEFAULT FALSE,
			is_premium BOOL NOT NULL DEFAULT FALSE
		);
		CREATE INDEX bb_user_idx ON bblog.users USING btree(user_id,created_ts, updated_ts,email,mobile);
	`

	_, err := DB.Exec(createUsersTable)

	if err != nil {
		log.Fatalf("Error creating users table: %v", err) // Log the real error

	}

	// if err != nil {
	// 	panic("Could not create users table")
	// }
}
