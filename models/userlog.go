package models

import "github.com/devmarvs/bblog/db"

type UserLog struct {
	UserLogId      int64  `json:"user_log_id"`
	UserId         int64  `json:"user_id"`
	SubUserId      int64  `json:"sub_user_id"`
	SubUserName    string `json:"name"`
	LogTypeId      int64  `json:"log_type_id"`
	LogTypeName    string `json:"log_name"`
	LogTime        string `json:"log_time"`
	LogDescription string `json:"log_description"`
	CreatedTs      string `json:"created_ts"`
	UpdatedTs      string `json:"updated_ts"`
}

func (ul *UserLog) Save() error {

	query := `
		INSERT INTO bblog.user_log
		(user_id, sub_user_id, log_type_id, log_time, log_description)
		VALUES
		($1, $2, $3, $4, $5)
		RETURNING user_log_id, sub_user_id, log_type_id, log_time, log_description, created_ts
	`
	var userLogId int64
	var subUserId int64
	var logTypeId int64
	var logTime string
	var logDescription string
	var createdTs string
	err := db.DB.QueryRow(
		query,
		ul.UserId,
		ul.SubUserId,
		ul.LogTypeId,
		ul.LogTime,
		ul.LogDescription).Scan(
		&userLogId,
		&subUserId,
		&logTypeId,
		&logTime,
		&logDescription,
		&createdTs,
	)

	if err != nil {
		return err
	}

	ul.UserLogId = userLogId
	ul.SubUserId = subUserId
	ul.LogTypeId = logTypeId
	ul.LogTime = logTime
	ul.LogDescription = logDescription
	ul.CreatedTs = createdTs
	return nil
}

func GetLogByUserAndSubUser(userId, subUserId int64) ([]UserLog, error) {

	query := `
		SELECT 
		ul.user_log_id, 
		ul.user_id,
		ul.sub_user_id,
		su.name,
		ul.log_type_id,
		lt.log_name,
		ul.log_time,
		ul.log_description,
		ul.created_ts
		FROM bblog.user_log ul 
		LEFT JOIN bblog.sub_users su ON ul.sub_user_id = su.sub_user_id
		LEFT JOIN bblog.log_types lt ON ul.log_type_id = lt.log_type_id
		WHERE ul.user_id = $1 AND ul.sub_user_id = $2
		ORDER BY ul.user_log_id DESC
	`
	rows, err := db.DB.Query(query, userId, subUserId)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var userLogs []UserLog
	for rows.Next() {
		var userLog UserLog
		err := rows.Scan(
			&userLog.UserLogId,
			&userLog.UserId,
			&userLog.SubUserId,
			&userLog.SubUserName,
			&userLog.LogTypeId,
			&userLog.LogTypeName,
			&userLog.LogTime,
			&userLog.LogDescription,
			&userLog.CreatedTs,
		)
		if err != nil {
			return nil, err
		}
		userLogs = append(userLogs, userLog)
	}
	return userLogs, nil

}
