package routes

import "github.com/gin-gonic/gin"

func RegisterRoutes(server *gin.Engine) {

	server.POST("/user/create", createUser)
	server.GET("/user/all", getUsers)
	server.GET("/user/:id", getUserById)
}
