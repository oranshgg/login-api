package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"login-api/config"
	"login-api/controller"
)

func main() {
	errEnv := godotenv.Load()
	if errEnv != nil {
		log.Fatal("Error loading .env file")
	}

	r := gin.Default()

	config.ConnectDatabase()

	r.POST("/register", controller.Register)
	r.POST("/login", controller.Login)
	r.GET("/ping", controller.Ping)

	err := r.Run(":8080")
	if err != nil {
		log.Fatal("error instantiating port")
	}
}
