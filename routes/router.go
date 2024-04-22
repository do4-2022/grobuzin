package routes

import (
	"github.com/do4-2022/grobuzin/routes/function"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

func GetRoutes(db *gorm.DB, JWTSecret string, BuilderEndpoint string, minioClient *minio.Client) *gin.Engine {
	router := gin.Default()

	// requireAuthMiddleware := user.RequireAuth(JWTSecret)

	// log.Println("Setting up routes", requireAuthMiddleware)

	function.ConfigureRoutes(router, db, minioClient, BuilderEndpoint)
	// user.ConfigureRoutes(router, db, JWTSecret)

	return router

}
