package function

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

func ConfigureRoutes(router *gin.Engine, db *gorm.DB, minioClient *minio.Client) {

	group := router.Group("/function")

	bucketName := "functions"
	location := "eu-west-1"
	ctx := context.Background()
	err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}

	controller := Controller{minioClient, db}

	group.POST("/", controller.PostFunction)
	group.GET("/", controller.GetAllFunction)
	group.GET("/:id", controller.GetOneFunction)
	group.PUT("/:id", controller.PutFunction)
	group.DELETE("/:id", controller.DeleteFunction)
}
