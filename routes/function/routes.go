package function

import "github.com/gin-gonic/gin"

func ConfigureRoutes(router *gin.Engine) {
	group := router.Group("/function")

	group.POST("/", PostFunction)
	group.GET("/", GetAllFunction)
	group.GET("/:id", GetOneFunction)
	group.PUT("/:id", PutFunction)
	group.DELETE("/:id", DeleteFunction)
}
