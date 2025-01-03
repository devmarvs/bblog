package routes

import (
	"net/http"

	"github.com/devmarvs/bblog/models"
	"github.com/gin-gonic/gin"
)

func createUser(context *gin.Context) {

	var user models.Users

	err := context.ShouldBindJSON(&user)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": "null", "message": "Could not parse request data"})
		return
	}

	err = user.Save()

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"data": "null", "message": "Could not save user"})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"data": "null", "message": "User created Successfully"})
}
