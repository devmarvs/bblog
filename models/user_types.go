package models

import "github.com/devmarvs/bblog/db"

type UserType struct {
	UserTypeID  int64  `json:"user_type_id"`
	Description string `json:"description"`
}

func ListUserTypes() ([]UserType, error) {
	const query = `
		SELECT user_type_id, description
		FROM bblog.user_type
		ORDER BY user_type_id ASC
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userTypes []UserType
	for rows.Next() {
		var ut UserType
		if err := rows.Scan(&ut.UserTypeID, &ut.Description); err != nil {
			return nil, err
		}
		userTypes = append(userTypes, ut)
	}

	return userTypes, nil
}
