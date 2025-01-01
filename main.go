package main

import (
	"github.com/devmarvs/bblog/db"
	"github.com/gin-gonic/gin"
)

func main() {

	db.InitDb()
	server := gin.Default()

	server.Run(":8080")
}
