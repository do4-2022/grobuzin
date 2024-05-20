package function

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/do4-2022/grobuzin/database"
	"github.com/do4-2022/grobuzin/objectStorage"
	"github.com/do4-2022/grobuzin/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	NotOwner = "You are not the owner of this function"
	NotFound = "Function not found!"
	DBError  = "Database error!"
)

type Controller struct {
	CodeStorageService 	*objectStorage.CodeStorageService
	DB                 	*gorm.DB
	BuilderEndpoint    	string
	Scheduler          	*scheduler.Scheduler
}

// Checks if said user owns the function in order to perform CRUD operations
func (c *Controller) IsOwner(userId uint, fnId string) (isOwner bool, err error) {
	var res database.Function
	isOwner = false

	tx := c.DB.Where(
		"ID = ?", 
		fnId, 
	).Select(
		"OwnerID", // because we don't need the whole object
	).First(&res)

	err = tx.Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println(err.Error())

		return 
	}

	log.Println("OwnerID: ", res.OwnerID, "UserID: ", userId)

	isOwner = tx.RowsAffected > 0 && res.OwnerID == userId 

	return
}

func (cont *Controller) GetAllFunction(c *gin.Context) {
	var functions []database.Function
	cont.DB.Find(&functions)

	c.JSON(http.StatusOK, functions)
}

func (cont *Controller) GetOneFunction(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetUint("userID")

	var function database.Function
	result := cont.DB.Find(&function, "ID = ?", id)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": DBError})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": NotFound})
		return
	}

	isOwner, err := cont.IsOwner(userID, id)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": DBError})
	} else if err != nil {
		// if the function does not exist, we return a 404
		c.JSON(http.StatusNotFound, gin.H{"error": NotFound})
		return
	}

	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": NotOwner})
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
	userID := c.GetUint("userID")

	// does the user exist?
	var user database.User
	err := cont.DB.First(&user, userID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found!"})
		return
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": DBError})
		return
	}

	var dto FunctionDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var function = database.Function{
		ID:          	id,
		Name:        	dto.Name,
		Description: 	dto.Description,
		Language:    	dto.Language,
		OwnerID: 		userID,
		Built: 			false,	
	}
	cont.DB.Create(&function)

	cont.CodeStorageService.PutCode(id, dto.Files)
	err = cont.buildImage(id.String(), dto.Language, dto.Files)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unable to build image!"})
		return
	}

	c.JSON(http.StatusOK, function)
}

func (cont *Controller) PutFunction(c *gin.Context) {
	userID := c.GetUint("userID")

	var json FunctionDTO
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var uuid uuid.UUID = uuid.MustParse(c.Param("id"))

	isOwner, err := cont.IsOwner(userID, uuid.String())

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": DBError})
	} else if err != nil {
		// if the function does not exist, we return a 404
		c.JSON(http.StatusNotFound, gin.H{"error": NotFound})
		return
	}

	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": NotOwner})
		return
	}

	var function = database.Function{
		ID:          uuid,
		Name:        json.Name,
		Description: json.Description,
		Language:    json.Language,
	}

	cont.DB.Save(function)

	cont.CodeStorageService.PutCode(uuid, json.Files)
	err = cont.buildImage(uuid.String(), json.Language, json.Files)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unable to build image!"})
		return
	}

	c.JSON(http.StatusOK, function)
}

func (cont *Controller) DeleteFunction(c *gin.Context) {
	var uuid uuid.UUID = uuid.MustParse(c.Param("id"))
	userID := c.GetUint("userID")

	isOwner, err := cont.IsOwner(userID, uuid.String())

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": DBError})
	} else if err != nil {
		// if the function does not exist, we return a 404
		c.JSON(http.StatusNotFound, gin.H{"error": NotFound})
		return
	}

	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": NotOwner})
		return
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": DBError})
	}

	cont.DB.Delete(&database.Function{ID: uuid})
	err = cont.CodeStorageService.DeleteCode(uuid)
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

	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid function ID"})
		return
	}

	var fn database.Function
	var fnState database.FunctionState

	// does the function exist?
	err = c.DB.Where(&database.Function{ID: fnID}).First(&fn).Error

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.AbortWithStatusJSON(404, gin.H{"error": NotFound})
		return
	}

	// does the function have an instance?
	fnState, _, err = c.Scheduler.LookForReadyInstance(fnID, 0)

	// if the function does not have an instance, we create ask the scheduler to create one
	if errors.Is(err, scheduler.ErrRecordNotFound) {
		res, err := c.Scheduler.SpawnVM(fnID)

		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}

		// retrieving freshly created function state
		fnState, err = c.Scheduler.GetStateByID(
			fmt.Sprintf(fnID.String(), ":", res.ID),
		)

		if err != nil {
			return
		} 
	} else if err != nil { // else if the error is not a record not found, we return an error
		log.Println(err.Error())
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Could not cold start the function"})
		return
	}

	stateID := fmt.Sprintf(fnID.String(), ":", fnState.ID)

	if fnState.Status != database.FnReady {
		log.Println("Waiting for function", fn.ID, "to be ready")
		time.Sleep(100 * time.Millisecond)


		// we will try 5 times to check if the instance is ready
		for attempts := 0; attempts < 5; attempts++ {
			// we check if it is ready
			fnState, err := c.Scheduler.GetStateByID(stateID)

			if err != nil { 
				log.Println(err.Error())
				ctx.AbortWithStatusJSON(500, gin.H{"Could not cold start the function": err.Error()})
				return
			}

			if fnState.Status == database.FnReady { 
				log.Println("Function", fnID, "is ready")
				break 
			};
		}

		// if even after 5 attempts the function is not ready, we return an error
		if fnState.Status != database.FnReady {
			ctx.AbortWithStatusJSON(503, gin.H{"error": "Function is not ready"})
			return
		}
	}

	// we notify everyone that the function is running
	err = c.Scheduler.SetStatus(stateID, database.FnRunning)

	if err != nil {
		log.Println(fmt.Sprint("Could not update state of VM", fnState.ID ,": ", err.Error()))
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Cannot update function's status"})
		return
	}
	
	_, err = http.Post(
		fmt.Sprint("http://", string(fnState.Address), ":", fnState.Port, "/execute"),
		"application/json",
		ctx.Request.Body,
	)

	// if the function had trouble running, we update the status to unknown
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		_ = c.Scheduler.SetStatus(stateID, database.FnUnknownState)
		return
	} else {
		ctx.Status(204)
	}

	err = c.Scheduler.SetStatus(stateID, database.FnReady)

	if err != nil {
		log.Println(
			fmt.Sprint("Could not update state of VM", fnState.ID ,": ", err.Error()),
		)
		log.Println(err.Error())
	}
}

