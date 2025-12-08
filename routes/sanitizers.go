package routes

import (
	"strings"

	"github.com/devmarvs/bblog/models"
)

func sanitizeUserPayload(user *models.Users) {
	user.UserName = strings.TrimSpace(user.UserName)
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	user.Mobile = strings.TrimSpace(user.Mobile)
	user.CountryCode = strings.TrimSpace(user.CountryCode)
	user.Password = strings.TrimSpace(user.Password)
}

func sanitizeSubUserPayload(subUser *models.SubUsers) {
	subUser.Name = strings.TrimSpace(subUser.Name)
}

func sanitizeUserLog(userLog *models.UserLog) {
	userLog.LogDescription = strings.TrimSpace(userLog.LogDescription)
	userLog.LogTime = strings.TrimSpace(userLog.LogTime)
}

func sanitizeAppVersion(version *models.AppVersion) {
	version.APIVersion = strings.TrimSpace(version.APIVersion)
	version.MobileVersion = strings.TrimSpace(version.MobileVersion)
}
