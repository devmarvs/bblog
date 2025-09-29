package routes

import (
	"net/http"
	"strconv"

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
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could not parse request data", "error": err.Error()})
		return
	}

	subUser, err := models.GetSubUserById(userLog.SubUserId)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could not fetch Sub Users", "error": err.Error()})
		return
	}

	if subUser.UserId != authenticatedID || !subUser.IsActive || subUser.IsDeleted {
		context.JSON(http.StatusForbidden, gin.H{"data": nil, "message": "Not Authorized"})
		return
	}

	userLog.UserId = authenticatedID

	if err := userLog.Save(); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could not save user log", "error": err.Error()})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"data": userLog, "message": "User Log created Successfully"})
}

func getLogByUser(context *gin.Context) {
	userId, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could Not Parse User Id", "error": err.Error()})
		return
	}

	authUserID, ok := context.Get("userId")
	if !ok {
		context.JSON(http.StatusUnauthorized, gin.H{"data": nil, "message": "Not Authorized"})
		return
	}

	authenticatedID, ok := authUserID.(int64)
	if !ok || authenticatedID != userId {
		context.JSON(http.StatusForbidden, gin.H{"data": nil, "message": "Not Authorized"})
		return
	}

	subUserId, err := strconv.ParseInt(context.Param("subuserid"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could Not Parse Sub User Id", "error": err.Error()})
		return
	}

	if _, err := models.GetUserById(userId); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not fetch User", "error": err.Error()})
		return
	}

	subUser, err := models.GetSubUserById(subUserId)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could not fetch Sub Users", "error": err.Error()})
		return
	}

	if subUser.UserId != authenticatedID || !subUser.IsActive || subUser.IsDeleted {
		context.JSON(http.StatusForbidden, gin.H{"data": nil, "message": "Not Authorized"})
		return
	}

	user, err := models.GetLogByUserAndSubUser(userId, subUserId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not fetch User", "error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"data": user})
}
