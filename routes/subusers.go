package routes

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/devmarvs/bblog/models"
	"github.com/gin-gonic/gin"
)

func createLog(context *gin.Context) {
	authenticatedID, ok := getAuthenticatedUserID(context)
	if !ok {
		respondUnauthorized(context)
		return
	}

	var userLog models.UserLog
	if err := context.ShouldBindJSON(&userLog); err != nil {
		respondBadRequest(context, ErrRequestDataParse)
		return
	}

	sanitizeUserLog(&userLog)
	if status, message := validateUserLog(&userLog); status != 0 {
		respondWithError(context, status, message)
		return
	}

	subUser, err := models.GetSubUserById(userLog.SubUserId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondNotFound(context, ErrSubUserNotFound)
			return
		}

		context.Error(err)
		respondInternalServerErrorWithData(context, ErrFetchSubUser)
		return
	}

	if subUser.UserId != authenticatedID || !subUser.IsActive || subUser.IsDeleted {
		respondForbidden(context)
		return
	}

	userLog.UserId = authenticatedID

	if err := userLog.Save(); err != nil {
		context.Error(err)
		respondInternalServerErrorWithData(context, ErrSaveUserLog)
		return
	}

	respondWithDataAndMessage(context, http.StatusCreated, userLog, "User log created successfully")
}

func getLogByUser(context *gin.Context) {
	userId := extractAndVerifyUserID(context)
	if userId == 0 {
		return
	}

	subUserId, err := strconv.ParseInt(context.Param("subuserid"), 10, 64)
	if err != nil {
		respondBadRequest(context, ErrParseSubUserID)
		return
	}

	if _, err := models.GetUserById(userId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondNotFound(context, ErrUserNotFound)
			return
		}

		context.Error(err)
		respondInternalServerErrorWithData(context, ErrFetchUser)
		return
	}

	subUser, err := models.GetSubUserById(subUserId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondNotFound(context, ErrSubUserNotFound)
			return
		}

		context.Error(err)
		respondInternalServerErrorWithData(context, ErrFetchSubUser)
		return
	}

	if subUser.UserId != userId || !subUser.IsActive || subUser.IsDeleted {
		respondForbidden(context)
		return
	}

	user, err := models.GetLogByUserAndSubUser(userId, subUserId)
	if err != nil {
		context.Error(err)
		respondInternalServerErrorWithData(context, ErrFetchUserLog)
		return
	}

	respondWithData(context, http.StatusOK, user)
}

func listLogTypes(context *gin.Context) {
	logTypes, err := models.ListLogTypes()
	if err != nil {
		context.Error(err)
		respondInternalServerErrorWithData(context, ErrFetchLogTypes)
		return
	}

	respondWithData(context, http.StatusOK, logTypes)
}
