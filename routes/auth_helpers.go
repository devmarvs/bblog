package routes

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// getAuthenticatedUserID retrieves the user ID from the context.
// Returns 0 and false if not found or invalid type.
func getAuthenticatedUserID(context *gin.Context) (int64, bool) {
	authUserID, ok := context.Get("userId")
	if !ok {
		return 0, false
	}
	id, ok := authUserID.(int64)
	return id, ok
}

// extractAndVerifyUserID extracts the user ID from the URL parameter "id"
// and verifies it matches the authenticated user ID.
// Returns the user ID if successful, otherwise handles the error response and returns 0.
func extractAndVerifyUserID(context *gin.Context) int64 {
	paramID := context.Param("id")
	userId, err := strconv.ParseInt(paramID, 10, 64)
	if err != nil {
		respondBadRequest(context, ErrParseUserID)
		return 0
	}

	authID, ok := getAuthenticatedUserID(context)
	if !ok {
		respondUnauthorized(context)
		return 0
	}

	if authID != userId {
		respondForbidden(context)
		return 0
	}

	return userId
}
