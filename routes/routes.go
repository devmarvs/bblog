package routes

import (
	"github.com/devmarvs/bblog/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(server *gin.Engine) {
	// Authenticated endpoints
	api := server.Group("/bblog")
	api.Use(middlewares.Authenticate)
	{
		api.POST("/user/:id/subuser", createSubUser)
		api.POST("/subuser/log", createLog)
		api.GET("/user/:id/subuser", getSubUserByUser)
		api.GET("/user/:id/subuser/:subuserid/log", getLogByUser)
		api.GET("/user/:id", getUserById)
		api.GET("/user/all", getUsers)
		api.GET("/user/types", listUserTypes)
		api.GET("/log/types", listLogTypes)
		api.POST("/logout", logout)
		api.POST("/version", createAppVersion)
	}

	// Public endpoints
	public := server.Group("/bblog")
	{
		public.POST("/user/create", createUser)
		public.POST("/user/forgot-password", forgotPassword)
		public.POST("/user/reset-password", resetPassword)
		public.POST("/login", login)
		public.GET("/version", getLatestAppVersion)

		// Verification routes
		registerVerificationRoutes(public)
	}

	// Alias endpoints without /bblog prefix (in case a reverse proxy strips it)
	publicAlias := server.Group("/")
	{
		registerVerificationRoutes(publicAlias)
	}
}

func registerVerificationRoutes(group *gin.RouterGroup) {
	group.POST("/user/verify-email", resendVerificationEmail)
	group.POST("/user/verify-email/request", resendVerificationEmail)
	group.GET("/user/verify-email", resendVerificationEmail)
	group.POST("/user/resend-verification", resendVerificationEmail)
	group.GET("/user/verify", verifyUserEmail)
	group.POST("/user/verify", verifyUserEmail)
}
