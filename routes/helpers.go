package routes

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// requireSameUser ensures the authenticated user matches the user id in the route param.
// It returns the parsed user id or writes the appropriate response on failure.
func requireSameUser(context *gin.Context, paramName string) (int64, bool) {
	userId, err := strconv.ParseInt(context.Param(paramName), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": nil, "message": "Could not parse user id"})
		return 0, false
	}

	authUserID, ok := context.Get("userId")
	if !ok {
		context.JSON(http.StatusUnauthorized, gin.H{"data": nil, "message": "Not Authorized"})
		return 0, false
	}

	authID, ok := authUserID.(int64)
	if !ok || authID != userId {
		context.JSON(http.StatusForbidden, gin.H{"data": nil, "message": "Not Authorized"})
		return 0, false
	}

	return userId, true
}
