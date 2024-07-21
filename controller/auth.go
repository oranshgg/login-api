package controller

import (
	"login-api/config"
	"login-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "ping"})
}

func Register(c *gin.Context) {
	/*
		En realidad acá se deberían validar los parámetros de la request y desp llamar
		a un servicio para la consulta a la DB.
	*/
	var userInput models.User
	if err := c.ShouldBindJSON(&userInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := hashPassword(userInput.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	userInput.Password = hashedPassword

	if err := config.DB.Create(&userInput).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": userInput})
}

func Login(c *gin.Context) {
	/*
		En realidad acá se deberían validar los parámetros de la request y desp llamar
		a un servicio para la consulta a la DB.
	*/
	var user models.User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var storedUser models.User
	if err := config.DB.Where("username = ?", user.Username).First(&storedUser).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	jwtToken, err := GenerateToken(storedUser.Username, storedUser.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not authenticate user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": jwtToken, "id": storedUser.ID})
}

func ValidateToken(c *gin.Context) {
	jwtToken := c.Request.Header.Get("Authorization")

	if jwtToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}

	err := VerifyToken(jwtToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
