package models

import "github.com/devmarvs/bblog/db"

type LogType struct {
	LogTypeID int64  `json:"log_type_id"`
	LogName   string `json:"log_name"`
}

func ListLogTypes() ([]LogType, error) {
	const query = `
		SELECT log_type_id, log_name
		FROM bblog.log_types
		ORDER BY log_type_id ASC
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logTypes []LogType
	for rows.Next() {
		var lt LogType
		if err := rows.Scan(&lt.LogTypeID, &lt.LogName); err != nil {
			return nil, err
		}
		logTypes = append(logTypes, lt)
	}

	return logTypes, nil
}
