package routes

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/devmarvs/bblog/models"
	"github.com/gin-gonic/gin"
)

func createLog(context *gin.Context) {
	authUserID, ok := context.Get("userId")
	if !ok {
		context.JSON(http.StatusUnauthorized, gin.H{"data": nil, "message": "Not Authorized"})
		return
	}

	authenticatedID, ok := authUserID.(int64)
	if !ok {
		context.JSON(http.StatusUnauthorized, gin.H{"data": nil, "message": "Not Authorized"})
		return
	}

	var userLog models.UserLog
	if err := context.ShouldBindJSON(&userLog); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could not parse request data"})
		return
	}

	sanitizeUserLog(&userLog)
	if status, message := validateUserLog(&userLog); status != 0 {
		context.JSON(status, gin.H{"data": nil, "message": message})
		return
	}

	subUser, err := models.GetSubUserById(userLog.SubUserId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			context.JSON(http.StatusNotFound, gin.H{"data": nil, "message": "Sub user not found"})
			return
		}

		context.Error(err)
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not fetch sub user"})
		return
	}

	if subUser.UserId != authenticatedID || !subUser.IsActive || subUser.IsDeleted {
		context.JSON(http.StatusForbidden, gin.H{"data": nil, "message": "Not Authorized"})
		return
	}

	userLog.UserId = authenticatedID

	if err := userLog.Save(); err != nil {
		context.Error(err)
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not save user log"})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"data": userLog, "message": "User log created successfully"})
}

func getLogByUser(context *gin.Context) {
	userId, ok := requireSameUser(context, "id")
	if !ok {
		return
	}

	subUserId, err := strconv.ParseInt(context.Param("subuserid"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could not parse sub user id"})
		return
	}

	if _, err := models.GetUserById(userId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			context.JSON(http.StatusNotFound, gin.H{"data": nil, "message": "User not found"})
			return
		}

		context.Error(err)
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not fetch user"})
		return
	}

	subUser, err := models.GetSubUserById(subUserId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			context.JSON(http.StatusNotFound, gin.H{"data": nil, "message": "Sub user not found"})
			return
		}

		context.Error(err)
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not fetch sub user"})
		return
	}

	if subUser.UserId != userId || !subUser.IsActive || subUser.IsDeleted {
		context.JSON(http.StatusForbidden, gin.H{"data": nil, "message": "Not Authorized"})
		return
	}

	user, err := models.GetLogByUserAndSubUser(userId, subUserId)
	if err != nil {
		context.Error(err)
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not fetch user log"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"data": user})
}

func sanitizeUserLog(userLog *models.UserLog) {
	userLog.LogDescription = strings.TrimSpace(userLog.LogDescription)
	userLog.LogTime = strings.TrimSpace(userLog.LogTime)
}

func validateUserLog(userLog *models.UserLog) (int, string) {
	if userLog.SubUserId <= 0 {
		return http.StatusBadRequest, "Invalid sub user id"
	}

	if userLog.LogTypeId <= 0 {
		return http.StatusBadRequest, "Invalid log type id"
	}

	return 0, ""
}

func listLogTypes(context *gin.Context) {
	logTypes, err := models.ListLogTypes()
	if err != nil {
		context.Error(err)
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not fetch log types"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"data": logTypes})
}
