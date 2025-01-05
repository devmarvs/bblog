package models

import "github.com/devmarvs/bblog/db"

type SubUsers struct {
	SubUserId  int64  `json:"sub_user_id"`
	UserId     int64  `json:"user_id"`
	UserTypeId int64  `json:"user_type_id"`
	Name       string `json:"name"`
	IsActive   bool   `json:"is_active"`
	IsDeleted  bool   `json:"is_deleted"`
	CreatedTs  string `json:"created_ts"`
	UpdatedTs  string `json:"updated_ts"`
}

func (s *SubUsers) Save(userId int64) error {

	query := `
		INSERT INTO bblog.sub_users(user_id,user_type_id, name) 
		VALUES ($1, $2, $3) 
		RETURNING sub_user_id, created_ts
	`

	var insertedId int64
	var createdTs string
	err := db.DB.QueryRow(query, userId, s.UserTypeId, s.Name).Scan(&insertedId, &createdTs)

	s.SubUserId = insertedId
	s.CreatedTs = createdTs
	return err
}

func GetSubUserById(subUserId int64) (*SubUsers, error) {

	query := `
		SELECT 
		sub_user_id
		FROM bblog.sub_users 
		WHERE is_active IS TRUE 
		AND is_deleted IS FALSE
		AND sub_user_id = $1
	`

	row := db.DB.QueryRow(query, subUserId)
	var subUser SubUsers
	err := row.Scan(&subUser.SubUserId)
	if err != nil {
		return nil, err
	}

	return &subUser, nil
}
