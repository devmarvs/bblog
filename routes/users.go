package routes

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/devmarvs/bblog/models"
	"github.com/devmarvs/bblog/utils"
	"github.com/gin-gonic/gin"
)

const minPasswordLength = 8

func createUser(context *gin.Context) {
	var user models.Users
	if err := context.ShouldBindJSON(&user); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could not parse request data"})
		return
	}

	sanitizeUserPayload(&user)

	if status, message := validateNewUser(&user); status != 0 {
		context.JSON(status, gin.H{"data": nil, "message": message})
		return
	}

	if err := user.Save(); err != nil {
		context.Error(err)
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could not save user"})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"data": nil, "message": "User created successfully"})
}

func getUsers(context *gin.Context) {
	users, err := models.GetUsers()
	if err != nil {
		context.Error(err)
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
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could not parse user id"})
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
		if errors.Is(err, sql.ErrNoRows) {
			context.JSON(http.StatusNotFound, gin.H{"data": nil, "message": "User not found"})
			return
		}

		context.Error(err)
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not fetch user"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"data": user})
}

func createSubUser(context *gin.Context) {
	userId, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could not parse user id"})
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
		if errors.Is(err, sql.ErrNoRows) {
			context.JSON(http.StatusNotFound, gin.H{"data": nil, "message": "User not found"})
			return
		}

		context.Error(err)
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not fetch user"})
		return
	}

	var subUser models.SubUsers
	if err := context.ShouldBindJSON(&subUser); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data"})
		return
	}

	sanitizeSubUserPayload(&subUser)
	if status, message := validateSubUser(&subUser); status != 0 {
		context.JSON(status, gin.H{"data": nil, "message": message})
		return
	}

	if err := subUser.Save(userId); err != nil {
		context.Error(err)
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not create sub user. Try again later"})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"data": subUser})
}

func getSubUserByUser(context *gin.Context) {
	userId, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could not parse user id"})
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
		if errors.Is(err, sql.ErrNoRows) {
			context.JSON(http.StatusNotFound, gin.H{"data": nil, "message": "User not found"})
			return
		}

		context.Error(err)
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not fetch user"})
		return
	}

	subUsers, err := models.GetSubUserByUser(userId)
	if err != nil {
		context.Error(err)
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not fetch sub users"})
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
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could not parse request data"})
		return
	}

	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	user.Password = strings.TrimSpace(user.Password)

	if err := user.ValidateCredentials(); err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			context.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid email or password"})
			return
		}

		context.Error(err)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not authenticate user."})
		return
	}

	token, err := utils.GenerateToken(user.Email, user.UserId)
	if err != nil {
		context.Error(err)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not authenticate user."})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Login Successful", "token": token})
}

func logout(context *gin.Context) {
	tokenValue, ok := context.Get("token")
	if !ok {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Not Authorized"})
		return
	}

	expiresValue, ok := context.Get("tokenExpiresAt")
	if !ok {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Not Authorized"})
		return
	}

	sanitizedToken, ok := tokenValue.(string)
	if !ok || sanitizedToken == "" {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Not Authorized"})
		return
	}

	expiresAt, ok := expiresValue.(time.Time)
	if !ok {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Not Authorized"})
		return
	}

	if err := models.RevokeToken(sanitizedToken, expiresAt); err != nil {
		context.Error(err)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not log out"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

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

func validateNewUser(user *models.Users) (int, string) {
	if user.Email == "" {
		return http.StatusBadRequest, "Email is required"
	}

	if len(user.Password) < minPasswordLength {
		return http.StatusBadRequest, "Password must be at least 8 characters long"
	}

	return 0, ""
}

func validateSubUser(subUser *models.SubUsers) (int, string) {
	if subUser.Name == "" {
		return http.StatusBadRequest, "Sub user name is required"
	}

	if subUser.UserTypeId <= 0 {
		return http.StatusBadRequest, "Invalid sub user type"
	}

	return 0, ""
}
