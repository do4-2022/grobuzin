package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetRoutes(db *gorm.DB) *gin.Engine {

	router := gin.Default()

	controller := Controller{}

	router.POST("/user", controller.createUser)

	return router

}
