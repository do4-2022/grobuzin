package function

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"time"
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
	BuilderEndpoint    string
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
	err := cont.buildImage(id.String(), dto.Language, dto.Files)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unable to build image!"})
		return
	}

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
	err := cont.buildImage(uuid.String(), json.Language, json.Files)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unable to build image!"})
		return
	}

	c.JSON(http.StatusOK, function)
}

func (cont *Controller) DeleteFunction(c *gin.Context) {
	var uuid uuid.UUID = uuid.MustParse(c.Param("id"))

	cont.DB.Delete(&database.Function{ID: uuid})
	err := cont.CodeStorageService.DeleteCode(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unable to delete code!"})
		return
	}
	err = cont.CodeStorageService.DeleteRootFs(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unable to delete rootfs!"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (cont *Controller) buildImage(id string, variant string, files map[string]string) error {
	request := BuilderRequest{
		Id: id,

		Variant: variant,
		Files:   files,
	}
	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	url := cont.BuilderEndpoint + "/build"
	response, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return nil
	}
	defer response.Body.Close()

	contentBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		fmt.Printf("The HTTP request failed with status code %d\n error: %s", response.StatusCode, contentBody)
		return errors.New("failed to build image")
	}
	return nil
}

type BuilderRequest struct {
	Id      string            `json:"id"`
	Files   map[string]string `json:"files"`
	Variant string            `json:"variant"`
}

func (c *Controller) RunFunction(ctx *gin.Context) {
    fnID, err := uuid.Parse(ctx.Param("id"))
	
	//var fnBody any; // because the body is entirely defined by the user, we just forward it to the function without checking it

	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid function ID"})
		return
	}

	var fn database.FunctionState
	err = c.DB.Where(&database.FunctionState{FunctionID: fnID}).First(&fn).Error

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.AbortWithStatusJSON(404, gin.H{"error": "Function not found"})
		return
	}

	if fn.Status != "Ready" {
		log.Println("Waiting for function", fn.ID, "to be ready")
		time.Sleep(100 * time.Millisecond)

		for attempts := 0; attempts < 5; attempts++ {
			err = c.DB.Where(&database.FunctionState{FunctionID: fnID, Status: "Ready"}).First(&fn).Error

			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Println(err)
				ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
				return
			} else {
				log.Println("Function", fn.FunctionID, "is not ready yet... Retrying")

				time.Sleep(100 * time.Millisecond)
			}

			if fn.Status == "Ready" { 
				log.Println("Function", fn.FunctionID, "is ready")

				break 
			};

		}

		if fn.Status != "Ready" {
			ctx.AbortWithStatusJSON(500, gin.H{"error": "Function is not ready"})
			return
		}
	}

	// TODO: forward the body to fn.Address:fn.Port
	ctx.JSON(200, fn)
}

