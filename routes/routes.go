package routes

import (
	"github.com/devmarvs/bblog/middlewares"
	"github.com/gin-gonic/gin"
)

// func RegisterRoutes(server *gin.Engine) {

// 	authenticated := server.Group("/")
// 	authenticated.Use(middlewares.Authenticate)
// 	authenticated.POST("/user/:id/subuser", createSubUser)
// 	authenticated.POST("/subuser/log", createLog)
// 	authenticated.GET("/user/:id/subuser", getSubUserByUser)
// 	authenticated.GET("/user/:id/subuser/:subuserid/log", getLogByUser)
// 	authenticated.GET("/user/:id", getUserById)

// 	server.POST("/user/create", createUser)
// 	server.GET("/user/all", getUsers)
// 	server.POST("/login", login)

// }

func RegisterRoutes(server *gin.Engine) {
	//authenticated endpoints
	api := server.Group("/bblog")
	api.Use(middlewares.Authenticate)
	api.POST("/user/:id/subuser", createSubUser)
	api.POST("/subuser/log", createLog)
	api.GET("/user/:id/subuser", getSubUserByUser)
	api.GET("/user/:id/subuser/:subuserid/log", getLogByUser)
	api.GET("/user/:id", getUserById)
	api.GET("/user/all", getUsers)

	//public endpoints
	public := server.Group("/bblog")
	public.POST("/user/create", createUser)
	public.POST("/login", login)

}
