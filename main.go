package main

import (
	"log"
	"os"

	"github.com/devmarvs/bblog/db"
	"github.com/devmarvs/bblog/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	os.MkdirAll("./uploads", os.ModePerm)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db.InitDb()
	server := gin.Default()

	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = gin.DebugMode // Default to debug if GIN_MODE is not set
	}
	gin.SetMode(mode)

	err = server.SetTrustedProxies(nil) // Trust all proxies
	// err = server.SetTrustedProxies([]string{"192.168.0.1", "10.0.0.0/8"})
	if err != nil {
		panic(err)
	}
	routes.RegisterRoutes(server)
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}
	server.Run(":" + port)
}
