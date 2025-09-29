package routes

import (
	"net/http"
	"strconv"

	"github.com/devmarvs/bblog/models"
	"github.com/devmarvs/bblog/utils"
	"github.com/gin-gonic/gin"
)

func createUser(context *gin.Context) {
	var user models.Users
	if err := context.ShouldBindJSON(&user); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could not parse request data", "error": err.Error()})
		return
	}

	if err := user.Save(); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could not save user", "error": err.Error()})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"data": nil, "message": "User created Successfully"})
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
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could Not Parse User Id", "error": err.Error()})
		return
	}

	authUserID, ok := context.Get("userId")
	if !ok {
		context.JSON(http.StatusUnauthorized, gin.H{"data": nil, "message": "Not Authorized"})
		return
	}

	if authID, ok := authUserID.(int64); !ok || authID != userId {
		context.JSON(http.StatusForbidden, gin.H{"data": nil, "message": "Not Authorized"})
		return
	}

	user, err := models.GetUserById(userId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not fetch User", "error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"data": user})
}

func createSubUser(context *gin.Context) {
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

	if authID, ok := authUserID.(int64); !ok || authID != userId {
		context.JSON(http.StatusForbidden, gin.H{"data": nil, "message": "Not Authorized"})
		return
	}

	if _, err := models.GetUserById(userId); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse User Id. Try again later", "error": err.Error()})
		return
	}

	var subUser models.SubUsers
	if err := context.ShouldBindJSON(&subUser); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data", "error": err.Error()})
		return
	}

	if err := subUser.Save(userId); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not create sub user. Try again later", "error": err.Error()})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"data": subUser})
}

func getSubUserByUser(context *gin.Context) {
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

	if authID, ok := authUserID.(int64); !ok || authID != userId {
		context.JSON(http.StatusForbidden, gin.H{"data": nil, "message": "Not Authorized"})
		return
	}

	if _, err := models.GetUserById(userId); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not fetch User", "error": err.Error()})
		return
	}

	subUsers, err := models.GetSubUserByUser(userId)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could not fetch Sub Users", "error": err.Error()})
		return
	}

	if len(subUsers) == 0 {
		context.JSON(http.StatusOK, gin.H{"data": nil})
		return
	}

	context.JSON(http.StatusOK, gin.H{"data": subUsers})
}

func login(context *gin.Context) {
	var user models.Users
	if err := context.ShouldBindJSON(&user); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could not parse request data", "error": err.Error()})
		return
	}

	if err := user.ValidateCredentials(); err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	token, err := utils.GenerateToken(user.Email, user.UserId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not authenticate user."})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Login Successful", "token": token})
}
