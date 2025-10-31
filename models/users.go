package models

import (
	"database/sql"
	"errors"

	"github.com/devmarvs/bblog/db"
	"github.com/devmarvs/bblog/utils"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

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
        INSERT INTO bblog.users(user_type_id, username, password, email, mobile, country_code)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING user_id
    `

	hashedPassword, err := utils.HashPassword(u.Password)
	if err != nil {
		return err
	}

	var insertedId int64
	if err := db.DB.QueryRow(query, u.UserTypeId, u.UserName, hashedPassword, u.Email, u.Mobile, u.CountryCode).Scan(&insertedId); err != nil {
		return err
	}

	u.UserId = insertedId
	return nil
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
        FROM bblog.users
        WHERE is_active IS TRUE
            AND is_deleted IS FALSE
        ORDER BY user_id DESC
    `
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []Users
	for rows.Next() {
		var user Users
		if err := rows.Scan(
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
		); err != nil {
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
        FROM bblog.users
        WHERE user_id = $1
            AND is_deleted IS FALSE
            AND is_active IS TRUE
    `
	row := db.DB.QueryRow(query, userId)
	var user Users
	if err := row.Scan(
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
	); err != nil {
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
            created_ts,
            COALESCE(updated_ts::text, '')
        FROM bblog.sub_users
        WHERE user_id = $1
            AND is_active IS TRUE
            AND is_deleted IS FALSE
    `

	rows, err := db.DB.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subUsers []SubUsers
	for rows.Next() {
		var subUser SubUsers
		if err := rows.Scan(
			&subUser.SubUserId,
			&subUser.UserId,
			&subUser.Name,
			&subUser.UserTypeId,
			&subUser.IsActive,
			&subUser.IsDeleted,
			&subUser.CreatedTs,
			&subUser.UpdatedTs,
		); err != nil {
			return nil, err
		}

		subUsers = append(subUsers, subUser)
	}

	return subUsers, nil
}

func (u *Users) ValidateCredentials() error {
	query := `
        SELECT user_id, password FROM bblog.users
        WHERE email = $1
            AND is_deleted IS FALSE
            AND is_active IS TRUE
    `
	row := db.DB.QueryRow(query, u.Email)

	var retrievedPassword string
	if err := row.Scan(&u.UserId, &retrievedPassword); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidCredentials
		}
		return err
	}

	passwordIsValid := utils.CheckPasswordHash(u.Password, retrievedPassword)
	if !passwordIsValid {
		return ErrInvalidCredentials
	}

	return nil
}
