package routes

import "time"

const (
	MinPasswordLength           = 8
	VerificationCodeLength      = 6
	VerificationCodeTTL         = 5 * time.Minute
	VerificationCodeMaxAttempts = 5
	VerificationResendCooldown  = 1 * time.Minute
	PasswordResetTokenSize      = 32
	PasswordResetTokenTTL       = 1 * time.Hour
)

const (
	ErrRequestDataParse            = "Could not parse request data"
	ErrEmailRequired               = "Email is required"
	ErrPasswordLength              = "Password must be at least 8 characters long"
	ErrUserSave                    = "Could not save user"
	ErrVerificationSend            = "Could not send verification email"
	ErrEmailServiceConfig          = "Email service is not configured"
	ErrPasswordResetProcess        = "Could not process password reset"
	ErrPasswordResetSend           = "Could not send password reset email"
	ErrResetTokenRequired          = "Reset token is required"
	ErrInvalidResetToken           = "Invalid reset token"
	ErrExpiredResetLink            = "Reset link has expired"
	ErrUsedResetLink               = "Reset link already used"
	ErrPasswordReset               = "Could not reset password"
	ErrResendVerification          = "Could not resend verification email"
	ErrAccountAlreadyVerified      = "Account is already verified"
	ErrFetchUsers                  = "Could not fetch users. Try again later"
	ErrParseUserID                 = "Could not parse user id"
	ErrNotAuthorized               = "Not Authorized"
	ErrUserNotFound                = "User not found"
	ErrFetchUser                   = "Could not fetch user"
	ErrSubUserNameRequired         = "Sub user name is required"
	ErrInvalidSubUserType          = "Invalid sub user type"
	ErrCreateSubUser               = "Could not create sub user. Try again later"
	ErrFetchSubUsers               = "Could not fetch sub users"
	ErrInvalidCredentials          = "Invalid email or password"
	ErrEmailNotVerified            = "Please verify your email before logging in"
	ErrAuthUser                    = "Could not authenticate user."
	ErrLogOut                      = "Could not log out"
	ErrInvalidSubUserID            = "Invalid sub user id"
	ErrInvalidLogTypeID            = "Invalid log type id"
	ErrSubUserNotFound             = "Sub user not found"
	ErrFetchSubUser                = "Could not fetch sub user"
	ErrSaveUserLog                 = "Could not save user log"
	ErrParseSubUserID              = "Could not parse sub user id"
	ErrFetchUserLog                = "Could not fetch user log"
	ErrFetchLogTypes               = "Could not fetch log types"
	ErrFetchUserTypes              = "Could not fetch user types"
	ErrVerificationCodeRequired    = "Email and verification code are required"
	ErrInvalidVerificationCode     = "Invalid verification code"
	ErrVerifyEmail                 = "Could not verify email"
	ErrVerificationCodeExpired     = "Verification code has expired"
	ErrVerificationCodeUsed        = "Verification code already used"
	ErrVerificationTooManyAttempts = "Too many invalid verification attempts. Request a new code"
	ErrVerificationResendTooSoon   = "Please wait before requesting another verification code"
	ErrVersionValuesRequired       = "API and mobile versions are required"
	ErrSaveAppVersion              = "Could not save version info"
	ErrFetchAppVersion             = "Could not fetch version info"
)
