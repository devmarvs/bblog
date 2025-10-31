package middlewares

import (
	"net/http"

	"github.com/devmarvs/bblog/models"
	"github.com/devmarvs/bblog/utils"
	"github.com/gin-gonic/gin"
)

func Authenticate(context *gin.Context) {
	token := context.Request.Header.Get("Authorization")

	if token == "" {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"data": nil, "message": "Not Authorized"})
		return
	}

	tokenDetails, err := utils.VerifyToken(token)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"data": nil, "message": "Not Authorized"})
		return
	}

	isRevoked, err := models.IsTokenRevoked(tokenDetails.Token)
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
