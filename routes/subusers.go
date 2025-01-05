package routes

import (
	"log"
	"net/http"
	"strconv"

	"github.com/devmarvs/bblog/models"
	"github.com/gin-gonic/gin"
)

func createLog(context *gin.Context) {

	var userLog models.UserLog

	err := context.ShouldBindJSON(&userLog)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": "null", "message": "Could not parse request data"})
		return
	}

	err = userLog.Save()

	if err != nil {
		log.Fatal(err)
		context.JSON(http.StatusBadRequest, gin.H{"data": "null", "message": "Could not save user log", "error": err})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"data": userLog, "message": "User Log created Successfully"})

}

func getLogByUser(context *gin.Context) {

	userId, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could Not Parse User Id", "error": err})
	}

	subUserId, err := strconv.ParseInt(context.Param("subuserid"), 10, 64)
	if err != nil {
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could Not Parse Sub User Id", "error": err})
		}
	}

	_, err = models.GetUserById(userId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not fetch User", "error": err})
		return
	}

	_, err = models.GetSubUserById(subUserId)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could not fetch Sub Users", "error": err})
		return
	}

	user, err := models.GetLogByUserAndSubUser(userId, subUserId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not fetch User", "error": err})
	}
	context.JSON(http.StatusOK, gin.H{"data": user})
}
