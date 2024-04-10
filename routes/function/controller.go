package function

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetAllFunction(c *gin.Context) {
	c.JSON(http.StatusOK, GetAllFunctions())
}

func GetOneFunction(c *gin.Context) {
	id := c.Param("id")

	c.JSON(http.StatusOK, functions[id])
}

func PostFunction(c *gin.Context) {
	id := uuid.New().String()
	var json FunctionDTO
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	function := dtoToFunction(id, json)
	functions[id] = function

	c.JSON(http.StatusOK, function)
}

func PutFunction(c *gin.Context) {
	id := c.Param("id")
	var json FunctionDTO
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	function := dtoToFunction(id, json)
	functions[id] = function

	c.JSON(http.StatusOK, function)
}

func DeleteFunction(c *gin.Context) {
	id := c.Param("id")

	delete(functions, id)

	c.JSON(http.StatusNoContent, nil)
}
