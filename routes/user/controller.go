package user

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Controller struct {
	DB        *gorm.DB
	JWTSecret string
}

func ConfigureRoutes(router *gin.Engine, db *gorm.DB, jwtSecret string) {
	controller := Controller{DB: db, JWTSecret: jwtSecret}
	group := router.Group("/user")

	group.POST("/", controller.createUser)
	group.POST("/login", controller.login)
}
