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

func UpdatePassword(c *gin.Context) {
	// Request body structure
	var passwordInput models.UpdatePasswordStruct

	// Bind the JSON input to the struct
	if err := c.ShouldBindJSON(&passwordInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Find the user in the database
	var user models.User
	if err := config.DB.Where("username = ?", passwordInput.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username"})
		return
	}

	// Check if the old password is correct
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(passwordInput.OldPassword)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect old password"})
		return
	}

	// Hash the new password
	hashedPassword, err := hashPassword(passwordInput.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}

	// Update the password and save changes
	user.Password = hashedPassword
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
