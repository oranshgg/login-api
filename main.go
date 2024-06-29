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

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.JSON(200, gin.H{})
			return
		}

		c.Next()
	})

	r.POST("/register", controller.Register)
	r.POST("/login", controller.Login)
	r.GET("/validate-token", controller.ValidateToken)
	r.GET("/ping", controller.Ping)

	err := r.Run(":8081")
	if err != nil {
		log.Fatal("error instantiating port")
	}
}
