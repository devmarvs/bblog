package routes

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/devmarvs/bblog/models"
	"github.com/devmarvs/bblog/utils"
	"github.com/gin-gonic/gin"
)

func createUser(context *gin.Context) {
	var user models.Users
	if err := context.ShouldBindJSON(&user); err != nil {
		respondBadRequest(context, ErrRequestDataParse)
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
		handleEmailError(context, err, ErrVerificationSend)
		return
	}

	context.JSON(http.StatusCreated, gin.H{"data": nil, "message": "User created. Check your email for the verification code."})
}

func forgotPassword(context *gin.Context) {
	email, err := extractEmail(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data"})
		return
	}

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
		respondInternalServerError(context, ErrPasswordResetProcess)
		return
	}

	if err := sendPasswordReset(user); err != nil {
		handleEmailError(context, err, ErrPasswordResetSend)
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
		respondBadRequest(context, ErrRequestDataParse)
		return
	}

	payload.Password = strings.TrimSpace(payload.Password)
	payload.Token = strings.TrimSpace(payload.Token)

	if payload.Token == "" {
		respondBadRequest(context, ErrResetTokenRequired)
		return
	}

	if len(payload.Password) < MinPasswordLength {
		respondBadRequest(context, ErrPasswordLength)
		return
	}

	if err := models.ResetPassword(payload.Token, payload.Password); err != nil {
		switch {
		case errors.Is(err, models.ErrPasswordResetTokenInvalid):
			respondBadRequest(context, ErrInvalidResetToken)
		case errors.Is(err, models.ErrPasswordResetTokenExpired):
			respondBadRequest(context, ErrExpiredResetLink)
		case errors.Is(err, models.ErrPasswordResetTokenUsed):
			context.JSON(http.StatusConflict, gin.H{"message": ErrUsedResetLink})
		default:
			context.Error(err)
			respondInternalServerError(context, ErrPasswordReset)
		}
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

func resendVerificationEmail(context *gin.Context) {
	email, err := extractEmail(context)
	if err != nil {
		respondBadRequest(context, ErrRequestDataParse)
		return
	}

	if email == "" {
		respondBadRequest(context, ErrEmailRequired)
		return
	}

	user, err := models.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			context.JSON(http.StatusOK, gin.H{"message": "If the account exists, a verification code has been sent"})
			return
		}

		context.Error(err)
		respondInternalServerError(context, ErrResendVerification)
		return
	}

	if user.IsActive {
		respondBadRequest(context, ErrAccountAlreadyVerified)
		return
	}

	if err := sendVerification(user); err != nil {
		handleEmailError(context, err, ErrVerificationSend)
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Verification code sent"})
}

func getUsers(context *gin.Context) {
	users, err := models.GetUsers()
	if err != nil {
		context.Error(err)
		respondInternalServerError(context, ErrFetchUsers)
		return
	}

	if len(users) == 0 {
		respondWithData(context, http.StatusOK, nil)
		return
	}

	respondWithData(context, http.StatusOK, users)
}

func getUserById(context *gin.Context) {
	userId := extractAndVerifyUserID(context)
	if userId == 0 {
		return
	}

	user, err := models.GetUserById(userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondNotFound(context, ErrUserNotFound)
			return
		}

		context.Error(err)
		respondInternalServerErrorWithData(context, ErrFetchUser)
		return
	}

	respondWithData(context, http.StatusOK, user)
}

func createSubUser(context *gin.Context) {
	userId := extractAndVerifyUserID(context)
	if userId == 0 {
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

	var subUser models.SubUsers
	if err := context.ShouldBindJSON(&subUser); err != nil {
		respondBadRequest(context, ErrRequestDataParse)
		return
	}

	sanitizeSubUserPayload(&subUser)
	if status, message := validateSubUser(&subUser); status != 0 {
		respondWithError(context, status, message)
		return
	}

	if err := subUser.Save(userId); err != nil {
		context.Error(err)
		respondInternalServerErrorWithData(context, ErrCreateSubUser)
		return
	}

	respondWithData(context, http.StatusCreated, subUser)
}

func getSubUserByUser(context *gin.Context) {
	userId := extractAndVerifyUserID(context)
	if userId == 0 {
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

	subUsers, err := models.GetSubUserByUser(userId)
	if err != nil {
		context.Error(err)
		respondInternalServerErrorWithData(context, ErrFetchSubUsers)
		return
	}

	if len(subUsers) == 0 {
		respondWithData(context, http.StatusOK, nil)
		return
	}

	respondWithData(context, http.StatusOK, subUsers)
}

func login(context *gin.Context) {
	var user models.Users
	if err := context.ShouldBindJSON(&user); err != nil {
		respondBadRequest(context, ErrRequestDataParse)
		return
	}

	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	user.Password = strings.TrimSpace(user.Password)

	if err := user.ValidateCredentials(); err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			respondWithError(context, http.StatusUnauthorized, ErrInvalidCredentials)
			return
		}

		if errors.Is(err, models.ErrEmailNotVerified) {
			respondWithError(context, http.StatusForbidden, ErrEmailNotVerified)
			return
		}

		context.Error(err)
		respondInternalServerError(context, ErrAuthUser)
		return
	}

	token, err := utils.GenerateToken(user.Email, user.UserId)
	if err != nil {
		context.Error(err)
		respondInternalServerError(context, ErrAuthUser)
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Login Successful", "token": token})
}

func logout(context *gin.Context) {
	tokenValue, ok := context.Get("token")
	if !ok {
		respondUnauthorized(context)
		return
	}

	expiresValue, ok := context.Get("tokenExpiresAt")
	if !ok {
		respondUnauthorized(context)
		return
	}

	sanitizedToken, ok := tokenValue.(string)
	if !ok || sanitizedToken == "" {
		respondUnauthorized(context)
		return
	}

	expiresAt, ok := expiresValue.(time.Time)
	if !ok {
		respondUnauthorized(context)
		return
	}

	if err := models.RevokeToken(sanitizedToken, expiresAt); err != nil {
		context.Error(err)
		respondInternalServerError(context, ErrLogOut)
		return
	}

	respondWithMessage(context, http.StatusOK, "Logout successful")
}

func sendVerification(user *models.Users) error {
	code, err := utils.GenerateNumericCode(VerificationCodeLength)
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(VerificationCodeTTL)
	if err := models.CreateEmailVerification(user.UserId, code, expiresAt); err != nil {
		return err
	}

	return utils.SendVerificationEmail(user.Email, code, VerificationCodeTTL)
}

func sendPasswordReset(user *models.Users) error {
	token, err := utils.GenerateRandomToken(PasswordResetTokenSize)
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(PasswordResetTokenTTL)
	if err := models.CreatePasswordReset(user.UserId, token, expiresAt); err != nil {
		return err
	}

	resetURL := buildPasswordResetURL(token)
	return utils.SendPasswordResetEmail(user.Email, resetURL)
}

func buildPasswordResetURL(token string) string {
	return fmt.Sprintf("%s/reset-password?token=%s", applicationBaseURL(), token)
}

func extractVerificationPayload(context *gin.Context) (string, string, error) {
	email := strings.ToLower(strings.TrimSpace(context.Query("email")))
	code := strings.TrimSpace(context.Query("code"))
	if email != "" || code != "" {
		return email, code, nil
	}

	if context.Request.ContentLength == 0 {
		return "", "", nil
	}

	var payload struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}

	if err := context.ShouldBindJSON(&payload); err != nil {
		return "", "", err
	}

	return strings.ToLower(strings.TrimSpace(payload.Email)), strings.TrimSpace(payload.Code), nil
}

func extractEmail(context *gin.Context) (string, error) {
	if email := strings.ToLower(strings.TrimSpace(context.Query("email"))); email != "" {
		return email, nil
	}

	var payload struct {
		Email string `json:"email"`
	}

	if context.Request.ContentLength == 0 && context.Request.Method == http.MethodGet {
		return "", nil
	}

	if err := context.ShouldBindJSON(&payload); err != nil {
		return "", err
	}

	return strings.ToLower(strings.TrimSpace(payload.Email)), nil
}

func applicationBaseURL() string {
	baseURL := strings.TrimSpace(os.Getenv("APP_BASE_URL"))

	if baseURL == "" {
		if gin.Mode() == gin.ReleaseMode {
			baseURL = "https://api.devmarvs.com"
		} else {
			baseURL = "http://localhost:8080"
		}
		return baseURL
	}

	if gin.Mode() == gin.ReleaseMode && strings.Contains(baseURL, "localhost") {
		return "https://api.devmarvs.com"
	}

	return strings.TrimRight(baseURL, "/")
}

func handleEmailError(context *gin.Context, err error, defaultMessage string) {
	context.Error(err)
	message := defaultMessage
	if errors.Is(err, utils.ErrMissingSMTPConfig) {
		message = ErrEmailServiceConfig
	}
	respondInternalServerError(context, message)
}

func verifyUserEmail(context *gin.Context) {
	email, code, err := extractVerificationPayload(context)
	if err != nil {
		respondBadRequest(context, ErrRequestDataParse)
		return
	}

	if email == "" || code == "" {
		respondBadRequest(context, ErrVerificationCodeRequired)
		return
	}

	user, err := models.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondBadRequest(context, ErrInvalidVerificationCode)
			return
		}

		context.Error(err)
		respondInternalServerError(context, ErrVerifyEmail)
		return
	}

	if user.IsActive {
		respondWithError(context, http.StatusConflict, ErrAccountAlreadyVerified)
		return
	}

	if err := models.VerifyEmailCode(user.UserId, code); err != nil {
		switch {
		case errors.Is(err, models.ErrVerificationAlreadyUsed):
			respondWithError(context, http.StatusConflict, ErrVerificationCodeUsed)
		case errors.Is(err, models.ErrVerificationCodeExpired):
			respondBadRequest(context, ErrVerificationCodeExpired)
		case errors.Is(err, models.ErrVerificationCodeInvalid):
			respondBadRequest(context, ErrInvalidVerificationCode)
		default:
			context.Error(err)
			respondInternalServerError(context, ErrVerifyEmail)
		}
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Email verified successfully. You can now log in."})
}

func listUserTypes(context *gin.Context) {
	userTypes, err := models.ListUserTypes()
	if err != nil {
		context.Error(err)
		respondInternalServerErrorWithData(context, ErrFetchUserTypes)
		return
	}

	respondWithData(context, http.StatusOK, userTypes)
}
