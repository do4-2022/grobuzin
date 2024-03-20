package user

import (
	"log"

	"github.com/do4-2022/grobuzin/database"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

const BCRYPT_COST = 12

func (cont *Controller) createUser(c *gin.Context) {

	var input User

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), BCRYPT_COST)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to hash password"})
		return
	}

	cont.DB.Create(&database.User{
		Username:       input.Username,
		HashedPassword: string(hashedPassword),
	})

	c.JSON(201, gin.H{"message": "User created"})
}

func (cont *Controller) login(c *gin.Context) {

	var input User

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var user database.User

	if cont.DB.First(&user, "username = ?", input.Username).Error != nil {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(input.Password))

	if err != nil {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := cont.createJWT(user)

	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": "Failed to create JWT"})
		return
	}

	c.JSON(200, gin.H{"token": token})
}
