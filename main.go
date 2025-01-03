package main

import (
	"github.com/devmarvs/bblog/db"
	"github.com/devmarvs/bblog/routes"
	"github.com/gin-gonic/gin"
)

func main() {

	db.InitDb()
	server := gin.Default()
	routes.RegisterRoutes(server)
	server.Run(":8080")
}
