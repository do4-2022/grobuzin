package scheduler

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type FunctionLocation struct {
	Address string	`redis:"address"`
	Port    uint16	`redis:"port"`
}


type Scheduler struct {
	Redis *redis.Client
	Context *context.Context
	Lambdo *LambdoService
}


func (s *Scheduler) SpawnVM(functionId uuid.UUID) (LambdoRunResponse, err error) {
	res, err := s.Lambdo.RunFunction("lorm ipsum dolor sit amet") // TODO change this when available

	if (err != nil) {
		return
	}

	locationID := fmt.Sprintf(functionId.String(), ":", uuid.New().String())
	
	err = s.Redis.HSet(*s.Context, locationID, &FunctionLocation{ 
		Address: res.Address, 
		Port: res.Port,
	}).Err()
	
	return
}

func (s *Scheduler) GetFunctionLocations(functionId uuid.UUID) (locations []FunctionLocation, err error) {
	locationQuery := fmt.Sprintf(functionId.String(), ":*")

	IDs := s.Redis.Keys(*s.Context, locationQuery).Val()

	for _, ID := range IDs {
		var location FunctionLocation
		
		if s.Redis.HGetAll(*s.Context, ID).Scan(&location) != nil {
			return
		}

		locations = append(locations, location)
	}

	return 
}