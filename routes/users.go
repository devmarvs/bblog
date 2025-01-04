package routes

import (
	"net/http"
	"strconv"

	"github.com/devmarvs/bblog/models"
	"github.com/gin-gonic/gin"
)

func createUser(context *gin.Context) {

	var user models.Users

	err := context.ShouldBindJSON(&user)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": "null", "message": "Could not parse request data"})
		return
	}

	err = user.Save()

	if err != nil {
		// log.Fatalf("cound not save user: %v", err) // Log the real error
		context.JSON(http.StatusBadRequest, gin.H{"data": "null", "message": "Could not save user", "error": err})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"data": "null", "message": "User created Successfully"})
}

func getUsers(context *gin.Context) {

	users, err := models.GetUsers()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not fetch users. Try again later"})
		return
	}

	if len(users) == 0 {
		context.JSON(http.StatusOK, gin.H{"data": nil})
		return
	}
	context.JSON(http.StatusOK, gin.H{"data": users})

}

func getUserById(context *gin.Context) {

	userId, err := strconv.ParseInt(context.Param("id"), 10, 64)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could Not Parse User Id", "error": err})
	}

	user, err := models.GetUserById(userId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not fetch User", "error": err})
	}
	context.JSON(http.StatusOK, gin.H{"data": user})
}

func createSubUser(context *gin.Context) {

	userId, err := strconv.ParseInt(context.Param("id"), 10, 64)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could Not Parse User Id", "error": err})
	}

	// log.Fatal(userId)
	user, err := models.GetUserById(userId)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not Parse User Id. Try again later"})
		return
	}

	user.UserId = userId
	var subUser models.SubUsers
	err = context.ShouldBindJSON(&subUser)
	// log.Fatal(err)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data"})
		return
	}
	err = subUser.Save(userId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not create sub user. Try again later", "error": err})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"data": subUser})
}

func getSubUserByUser(context *gin.Context) {

	userId, err := strconv.ParseInt(context.Param("id"), 10, 64)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could Not Parse User Id", "error": err})
	}

	_, err = models.GetUserById(userId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not fetch User", "error": err})
	}

	subUsers, err := models.GetSubUserByUser(userId)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could not fetch Sub Users", "error": err})
		return
	}

	if len(subUsers) == 0 {
		context.JSON(http.StatusOK, gin.H{"data": nil})
		return
	}
	context.JSON(http.StatusOK, gin.H{"data": subUsers})
	return
}
