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
	UserName    string `json:"username"`
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
		INSERT INTO bblog.users(user_type_id,username, password,email,mobile,country_code) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING user_id
	`

	// stmt, err := db.DB.Prepare(query)

	// if err != nil {
	// 	log.Fatalf("Error preparing: %v", err) // Log the real error
	// 	return err
	// }

	// defer stmt.Close()

	hashedPassword, err := utils.HashPassword(u.Password)

	if err != nil {
		log.Fatalf("Error hashing: %v", err) // Log the real error
		return err
	}

	// result, err := stmt.Exec(u.UserTypeId, hashedPassword, u.Email, u.Mobile, u.CountryCode)

	// if err != nil {
	// 	log.Fatalf("Error saving users: %v", err) // Log the real error
	// 	// return err
	// }

	var insertedId int64
	err = db.DB.QueryRow(query, u.UserTypeId, u.UserName, hashedPassword, u.Email, u.Mobile, u.CountryCode).Scan(&insertedId)

	u.UserId = insertedId
	return err
}

func GetUsers() ([]Users, error) {

	query := `
		SELECT 
		user_id, 
		username,
		created_ts, 
		user_type_id,
		email,
		mobile,
		country_code,
		is_online,
		is_active,
		is_deleted,
		is_premium
		FROM bblog.users WHERE is_active IS TRUE 
		AND is_deleted IS FALSE ORDER BY user_id DESC
	`
	rows, err := db.DB.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []Users

	for rows.Next() {
		var user Users

		err := rows.Scan(
			&user.UserId,
			&user.UserName,
			&user.CreatedTs,
			&user.UserTypeId,
			&user.Email,
			&user.Mobile,
			&user.CountryCode,
			&user.IsOnline,
			&user.IsActive,
			&user.IsDeleted,
			&user.IsPremium,
		)

		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func GetUserById(userId int64) (*Users, error) {

	query := `
		SELECT
		user_id, 
		username,
		created_ts, 
		user_type_id,
		email,
		mobile,
		country_code,
		is_online,
		is_active,
		is_deleted,
		is_premium
		FROM bblog.users WHERE user_id = $1
	`
	row := db.DB.QueryRow(query, userId)
	var user Users
	err := row.Scan(
		&user.UserId,
		&user.UserName,
		&user.CreatedTs,
		&user.UserTypeId,
		&user.Email,
		&user.Mobile,
		&user.CountryCode,
		&user.IsOnline,
		&user.IsActive,
		&user.IsDeleted,
		&user.IsPremium,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func GetSubUserByUser(userId int64) ([]SubUsers, error) {

	query := `
		SELECT 
		sub_user_id,
		user_id,
		name,
		user_type_id,
		is_active,
		is_deleted,
		created_ts
		FROM bblog.sub_users 
		WHERE user_id = $1
	`

	rows, err := db.DB.Query(query, userId)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var subUsers []SubUsers

	for rows.Next() {
		var subUser SubUsers

		err := rows.Scan(
			&subUser.SubUserId,
			&subUser.UserId,
			&subUser.Name,
			&subUser.UserTypeId,
			&subUser.IsActive,
			&subUser.IsDeleted,
			&subUser.CreatedTs,
		)

		if err != nil {
			return nil, err
		}

		subUsers = append(subUsers, subUser)
	}

	return subUsers, nil
}
