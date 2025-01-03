package models

import (
	"log"

	"github.com/devmarvs/bblog/db"
	"github.com/devmarvs/bblog/utils"
)

type Users struct {
	UserId      int64  `json:"user_id"`
	CreatedTs   string `json:"created_ts"`
	UpdatedTs   string `json:"updated_ts"`
	UserTypeId  int64  `json:"user_type_id"`
	Password    string `json:"password"`
	Email       string `json:"email"`
	Mobile      string `json:"mobile"`
	CountryCode string `json:"country_code"`
	IsOnline    bool   `json:"is_online"`
	IsActive    bool   `json:"is_active"`
	IsDeleted   bool   `json:"is_deleted"`
	IsPremium   bool   `json:"is_premium"`
}

func (u *Users) Save() error {

	query := `
		INSERT INTO bblog.users
		(user_type_id, password, email, mobile, country_code) 
		VALUES (?, ?, ?, ?, ?)
	`

	stmt, err := db.DB.Prepare(query)

	if err != nil {
		log.Fatalf("Error preparing: %v", err) // Log the real error
		return err
	}

	defer stmt.Close()

	hashedPassword, err := utils.HashPassword(u.Password)

	if err != nil {
		log.Fatalf("Error hashing: %v", err) // Log the real error
		return err
	}

	result, err := stmt.Exec(u.UserTypeId, hashedPassword, u.Email, u.Mobile, u.CountryCode)

	if err != nil {
		log.Fatalf("Error saving users: %v", err) // Log the real error
		// return err
	}

	userId, err := result.LastInsertId()

	u.UserId = userId
	return err
}
