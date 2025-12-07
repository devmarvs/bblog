package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func respondWithError(context *gin.Context, status int, message string) {
	context.JSON(status, gin.H{"message": message})
}

func respondWithData(context *gin.Context, status int, data interface{}) {
	context.JSON(status, gin.H{"data": data})
}

func respondWithDataAndMessage(context *gin.Context, status int, data interface{}, message string) {
	context.JSON(status, gin.H{"data": data, "message": message})
}

func respondWithMessage(context *gin.Context, status int, message string) {
	context.JSON(status, gin.H{"message": message})
}

func respondBadRequest(context *gin.Context, message string) {
	respondWithError(context, http.StatusBadRequest, message)
}

func respondNotFound(context *gin.Context, message string) {
	context.JSON(http.StatusNotFound, gin.H{"data": nil, "message": message})
}

func respondUnauthorized(context *gin.Context) {
	context.JSON(http.StatusUnauthorized, gin.H{"data": nil, "message": ErrNotAuthorized})
}

func respondForbidden(context *gin.Context) {
	context.JSON(http.StatusForbidden, gin.H{"data": nil, "message": ErrNotAuthorized})
}

func respondInternalServerError(context *gin.Context, message string) {
	context.JSON(http.StatusInternalServerError, gin.H{"message": message})
}

func respondInternalServerErrorWithData(context *gin.Context, message string) {
	context.JSON(http.StatusInternalServerError, gin.H{"data": nil, "message": message})
}
