package routes

import (
	"log"

	"github.com/do4-2022/grobuzin/routes/user"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetRoutes(db *gorm.DB, JWTSecret string) *gin.Engine {

	router := gin.Default()

	requireAuthMiddleware := user.RequireAuth(JWTSecret)

	log.Println("Setting up routes", requireAuthMiddleware)

	user.ConfigureRoutes(router, db, JWTSecret)

	return router

}
