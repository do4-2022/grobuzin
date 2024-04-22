package function

import (
	"net/http"

	"github.com/do4-2022/grobuzin/database"
	"github.com/do4-2022/grobuzin/objectStorage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Controller struct {
	CodeStorageService *objectStorage.CodeStorageService
	DB                 *gorm.DB
}

func (cont *Controller) GetAllFunction(c *gin.Context) {
	var functions []database.Function
	cont.DB.Find(&functions)

	c.JSON(http.StatusOK, functions)
}

func (cont *Controller) GetOneFunction(c *gin.Context) {
	id := c.Param("id")

	var function database.Function
	result := cont.DB.Find(&function, "ID = ?", id)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Database error!"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Function not found!"})
		return
	}

	files, err := cont.CodeStorageService.GetCode(uuid.MustParse(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Code not found!"})
		return
	}

	dto := FunctionDTO{
		Name:        function.Name,
		Description: function.Description,
		Language:    function.Language,
		Files:       files,
	}

	c.JSON(http.StatusOK, dto)
}

func (cont *Controller) PostFunction(c *gin.Context) {
	id := uuid.New()
	var dto FunctionDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var function = database.Function{
		ID:          id,
		Name:        dto.Name,
		Description: dto.Description,
		Language:    dto.Language,
	}
	cont.DB.Create(&function)

	cont.CodeStorageService.PutCode(id, dto.Files)

	c.JSON(http.StatusOK, function)
}

func (cont *Controller) PutFunction(c *gin.Context) {
	var json FunctionDTO
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var uuid uuid.UUID = uuid.MustParse(c.Param("id"))
	var function = database.Function{
		ID:          uuid,
		Name:        json.Name,
		Description: json.Description,
		Language:    json.Language,
	}

	cont.DB.Save(function)

	cont.CodeStorageService.PutCode(uuid, json.Files)

	c.JSON(http.StatusOK, function)
}

func (cont *Controller) DeleteFunction(c *gin.Context) {
	var uuid uuid.UUID = uuid.MustParse(c.Param("id"))

	cont.DB.Delete(&database.Function{ID: uuid})
	err := cont.CodeStorageService.DeleteCode(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unable not delete code!"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
