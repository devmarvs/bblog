package routes

import (
	"net/http"

	"github.com/devmarvs/bblog/models"
)

func validateNewUser(user *models.Users) (int, string) {
	if user.Email == "" {
		return http.StatusBadRequest, ErrEmailRequired
	}

	if len(user.Password) < MinPasswordLength {
		return http.StatusBadRequest, ErrPasswordLength
	}

	return 0, ""
}

func validateSubUser(subUser *models.SubUsers) (int, string) {
	if subUser.Name == "" {
		return http.StatusBadRequest, ErrSubUserNameRequired
	}

	if subUser.UserTypeId <= 0 {
		return http.StatusBadRequest, ErrInvalidSubUserType
	}

	return 0, ""
}

func validateUserLog(userLog *models.UserLog) (int, string) {
	if userLog.SubUserId <= 0 {
		return http.StatusBadRequest, ErrInvalidSubUserID
	}

	if userLog.LogTypeId <= 0 {
		return http.StatusBadRequest, ErrInvalidLogTypeID
	}

	return 0, ""
}
