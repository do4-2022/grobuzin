package routes

import "github.com/gin-gonic/gin"

type User struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (cont *Controller) createUser(c *gin.Context) {
	// do stuff
}
