package routes

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/devmarvs/bblog/models"
	"github.com/devmarvs/bblog/utils"
	"github.com/gin-gonic/gin"
)

const (
	minPasswordLength      = 8
	verificationTokenSize  = 32
	verificationTokenTTL   = 24 * time.Hour
	passwordResetTokenSize = 32
	passwordResetTokenTTL  = 1 * time.Hour
)

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

	if err := sendVerification(&user); err != nil {
		context.Error(err)
		message := "Could not send verification email"
		if errors.Is(err, utils.ErrMissingSMTPConfig) {
			message = "Email service is not configured"
		}
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": message})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"data": nil, "message": "User created. Check your email to verify the account."})
}

func forgotPassword(context *gin.Context) {
	var payload struct {
		Email string `json:"email"`
	}

	if err := context.ShouldBindJSON(&payload); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(payload.Email))
	if email == "" {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Email is required"})
		return
	}

	const genericMessage = "If the account exists, a reset email has been sent"

	user, err := models.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			context.JSON(http.StatusOK, gin.H{"message": genericMessage})
			return
		}

		context.Error(err)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not process password reset"})
		return
	}

	if err := sendPasswordReset(user); err != nil {
		context.Error(err)
		message := "Could not send password reset email"
		if errors.Is(err, utils.ErrMissingSMTPConfig) {
			message = "Email service is not configured"
		}
		context.JSON(http.StatusInternalServerError, gin.H{"message": message})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": genericMessage})
}

func resetPassword(context *gin.Context) {
	var payload struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}

	if err := context.ShouldBindJSON(&payload); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data"})
		return
	}

	payload.Password = strings.TrimSpace(payload.Password)
	payload.Token = strings.TrimSpace(payload.Token)

	if payload.Token == "" {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Reset token is required"})
		return
	}

	if len(payload.Password) < minPasswordLength {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Password must be at least 8 characters long"})
		return
	}

	if err := models.ResetPassword(payload.Token, payload.Password); err != nil {
		switch {
		case errors.Is(err, models.ErrPasswordResetTokenInvalid):
			context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid reset token"})
		case errors.Is(err, models.ErrPasswordResetTokenExpired):
			context.JSON(http.StatusBadRequest, gin.H{"message": "Reset link has expired"})
		case errors.Is(err, models.ErrPasswordResetTokenUsed):
			context.JSON(http.StatusConflict, gin.H{"message": "Reset link already used"})
		default:
			context.Error(err)
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not reset password"})
		}
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

func resendVerificationEmail(context *gin.Context) {
	var payload struct {
		Email string `json:"email"`
	}

	if err := context.ShouldBindJSON(&payload); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(payload.Email))
	if email == "" {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Email is required"})
		return
	}

	user, err := models.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			context.JSON(http.StatusOK, gin.H{"message": "If the account exists, a verification email has been sent"})
			return
		}

		context.Error(err)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not resend verification email"})
		return
	}

	if user.IsActive {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Account is already verified"})
		return
	}

	if err := sendVerification(user); err != nil {
		context.Error(err)
		message := "Could not send verification email"
		if errors.Is(err, utils.ErrMissingSMTPConfig) {
			message = "Email service is not configured"
		}
		context.JSON(http.StatusInternalServerError, gin.H{"message": message})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Verification email sent"})
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

		if errors.Is(err, models.ErrEmailNotVerified) {
			context.JSON(http.StatusForbidden, gin.H{"message": "Please verify your email before logging in"})
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

func sendVerification(user *models.Users) error {
	token, err := utils.GenerateRandomToken(verificationTokenSize)
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(verificationTokenTTL)
	if err := models.CreateEmailVerification(user.UserId, token, expiresAt); err != nil {
		return err
	}

	verifyURL := buildVerificationURL(token)
	return utils.SendVerificationEmail(user.Email, verifyURL)
}

func buildVerificationURL(token string) string {
	return fmt.Sprintf("%s/bblog/user/verify?token=%s", applicationBaseURL(), token)
}

func sendPasswordReset(user *models.Users) error {
	token, err := utils.GenerateRandomToken(passwordResetTokenSize)
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(passwordResetTokenTTL)
	if err := models.CreatePasswordReset(user.UserId, token, expiresAt); err != nil {
		return err
	}

	resetURL := buildPasswordResetURL(token)
	return utils.SendPasswordResetEmail(user.Email, resetURL)
}

func buildPasswordResetURL(token string) string {
	return fmt.Sprintf("%s/reset-password?token=%s", applicationBaseURL(), token)
}

func applicationBaseURL() string {
	baseURL := strings.TrimSpace(os.Getenv("APP_BASE_URL"))
	if baseURL == "" {
		if gin.Mode() == gin.ReleaseMode {
			baseURL = "https://api.devmarvs.com"
		} else {
			baseURL = "http://localhost:8080"
		}
	}

	if gin.Mode() == gin.ReleaseMode && strings.Contains(baseURL, "localhost") {
		baseURL = "https://api.devmarvs.com"
	}

	return strings.TrimRight(baseURL, "/")
}

func verifyUserEmail(context *gin.Context) {
	token := context.Query("token")
	if strings.TrimSpace(token) == "" {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Missing verification token"})
		return
	}

	if _, err := models.VerifyEmailToken(token); err != nil {
		switch {
		case errors.Is(err, models.ErrVerificationAlreadyUsed):
			context.JSON(http.StatusConflict, gin.H{"message": "Verification link already used"})
		case errors.Is(err, models.ErrVerificationTokenExpired):
			context.JSON(http.StatusBadRequest, gin.H{"message": "Verification link has expired"})
		case errors.Is(err, models.ErrVerificationTokenInvalid):
			context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid verification link"})
		default:
			context.Error(err)
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not verify email"})
		}
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Email verified successfully. You can now log in."})
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

func listUserTypes(context *gin.Context) {
	userTypes, err := models.ListUserTypes()
	if err != nil {
		context.Error(err)
		context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Could not fetch user types"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"data": userTypes})
}
