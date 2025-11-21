package middlewares

import (
	"net/http"

	"github.com/devmarvs/bblog/models"
	"github.com/devmarvs/bblog/utils"
	"github.com/gin-gonic/gin"
)

var (
	verifyTokenFunc    = utils.VerifyToken
	isTokenRevokedFunc = models.IsTokenRevoked
)

// SetAuthDependencies overrides auth helpers (used in tests).
func SetAuthDependencies(verify func(string) (*utils.TokenDetails, error), isRevoked func(string) (bool, error)) {
	if verify != nil {
		verifyTokenFunc = verify
	} else {
		verifyTokenFunc = utils.VerifyToken
	}

	if isRevoked != nil {
		isTokenRevokedFunc = isRevoked
	} else {
		isTokenRevokedFunc = models.IsTokenRevoked
	}
}

func Authenticate(context *gin.Context) {
	token := context.Request.Header.Get("Authorization")

	if token == "" {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"data": nil, "message": "Not Authorized"})
		return
	}

	tokenDetails, err := verifyTokenFunc(token)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"data": nil, "message": "Not Authorized"})
		return
	}

	isRevoked, err := isTokenRevokedFunc(tokenDetails.Token)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"data": nil, "message": "Unable to authenticate request"})
		return
	}

	if isRevoked {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"data": nil, "message": "Not Authorized"})
		return
	}

	context.Set("userId", tokenDetails.UserID)
	context.Set("token", tokenDetails.Token)
	context.Set("tokenExpiresAt", tokenDetails.ExpiresAt)
	context.Next()
}
