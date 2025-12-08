package models

import "github.com/devmarvs/bblog/db"

type AppVersion struct {
	VersionID     int64  `json:"version_id"`
	APIVersion    string `json:"api_version"`
	MobileVersion string `json:"mobile_version"`
	CreatedTs     string `json:"created_ts"`
}

func (v *AppVersion) Save() error {
	const query = `
		INSERT INTO bblog.app_versions(api_version, mobile_version)
		VALUES ($1, $2)
		RETURNING version_id, created_ts
	`

	return db.DB.QueryRow(query, v.APIVersion, v.MobileVersion).Scan(&v.VersionID, &v.CreatedTs)
}

func GetLatestAppVersion() (*AppVersion, error) {
	const query = `
		SELECT version_id, api_version, mobile_version, created_ts
		FROM bblog.app_versions
		ORDER BY created_ts DESC, version_id DESC
		LIMIT 1
	`

	row := db.DB.QueryRow(query)
	var version AppVersion
	if err := row.Scan(&version.VersionID, &version.APIVersion, &version.MobileVersion, &version.CreatedTs); err != nil {
		return nil, err
	}

	return &version, nil
}
