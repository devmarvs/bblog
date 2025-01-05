package routes

import (
	"github.com/devmarvs/bblog/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(server *gin.Engine) {

	authenticated := server.Group("/")
	authenticated.Use(middlewares.Authenticate)
	authenticated.POST("/user/:id/subuser", createSubUser)
	authenticated.POST("/subuser/log", createLog)
	authenticated.GET("/user/:id/subuser", getSubUserByUser)
	authenticated.GET("/user/:id/subuser/:subuserid/log", getLogByUser)
	authenticated.GET("/user/:id", getUserById)

	server.POST("/user/create", createUser)
	server.GET("/user/all", getUsers)
	server.POST("/login", login)

}
