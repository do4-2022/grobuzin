package function

import (
	"github.com/do4-2022/grobuzin/objectStorage"
	"github.com/do4-2022/grobuzin/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

func ConfigureRoutes(router *gin.Engine, db *gorm.DB, minioClient *minio.Client, builderEndpoint string, scheduler scheduler.Scheduler) {

	group := router.Group("/function")

	codeStorageService := objectStorage.CodeStorageService{MinioClient: minioClient}
	codeStorageService.Init()

	controller := Controller{&codeStorageService, db, builderEndpoint, &scheduler}

	group.POST("/", controller.PostFunction)
	group.GET("/", controller.GetAllFunction)
	group.GET("/:id", controller.GetOneFunction)
	group.PUT("/:id", controller.PutFunction)
	group.DELETE("/:id", controller.DeleteFunction)
	group.POST("/:id/run", controller.RunFunction)
}
