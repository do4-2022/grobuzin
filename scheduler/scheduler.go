package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type StatusCode int

const (
	Creating StatusCode = iota
	Ready
	Running
	Unknown
)

// This struct represents an instance of a function, especially it's address and status 
// Each key in the redis is namespaced as it follows:
// <id of the function>:<id of the instance>
type FunctionLocation struct {
	Address 	string		`redis:"address"`
	Port    	uint16		`redis:"port"`
	Status 		StatusCode 	`redis:"status"`
	LastUsed	string		`redis:"lastUsed"`
}


type Scheduler struct {
	Redis *redis.Client
	Context *context.Context
	Lambdo *LambdoService
}

// Uses SCAN to look for the first instance marked as ready (https://redis.io/docs/latest/commands/scan/)
func (s *Scheduler) LookForReadyInstance(functionId uuid.UUID, cursor uint64) (id string, returnedCursor uint64, err error) {
	// match rule to get all instances of a function 
	locationMatch := fmt.Sprintf(functionId.String(), ":*")

	keys, returnedCursor, err := s.Redis.Scan(*s.Context, cursor, locationMatch, 10).Result()

	if err != nil {
		return
	}

	for _, id := range keys {
		code, err := s.Redis.HGet(*s.Context, id, "status").Int();

		if err != nil && StatusCode(code) == Ready {
			// we found a ready instance so we return it
			return id, 0, nil
		}
	}

	if returnedCursor != 0 {
		// the current sweep did not find anything, we try again 
		return s.LookForReadyInstance(functionId, returnedCursor)
	}

	return "", 0, errors.New("could not find an available function")  // we did not find anything thus, id is empty
} 

func (s *Scheduler) SpawnVM(functionId uuid.UUID) (LambdoRunResponse, err error) {
	res, err := s.Lambdo.SpawnVM(functionId)

	if (err != nil) {
		return
	}

	locationID := fmt.Sprintf(functionId.String(), ":", res.ID)
	
	err = s.Redis.HSet(*s.Context, locationID, &FunctionLocation{ 
		Address: res.Address, 
		Port: res.Port,
		Status: Creating,
		LastUsed: "never",
	}).Err()
	
	return
}

func (s *Scheduler) GetFunctionLocations(functionId uuid.UUID) (locations []FunctionLocation) {
	locationQuery := fmt.Sprintf(functionId.String(), ":*")
	firstsweep, cursor := true, uint64(0)
	
	// because do-while does not exists in these lands
	for !firstsweep && cursor != 0 {
		firstsweep = false

		var keys []string 
		keys, cursor = s.Redis.Scan(*s.Context, cursor, locationQuery, 10).Val()
		
		// for each key we got, we retrieve all of it's information
		for _, ID := range keys {
			var location FunctionLocation
			
			if s.Redis.HGetAll(*s.Context, ID).Scan(&location) != nil {
				return
			}

			locations = append(locations, location)
		}
	}

	return 
}

// goes through the whole redis instance and remove that have not been used within the last {hoursTimeout} Hours
func (s *Scheduler) FindAndDestroyUnsused(hoursTimeout float64) {
	now := time.Now()
	keys, cursor := s.Redis.Scan(*s.Context, 0, "*", 10).Val()

	for cursor != 0 {
		for _, ID := range keys {
			val, err := s.Redis.HGet(*s.Context, ID, "lastUsed").Result()

			if err != nil {
				continue
			}

			// if it was never used we delete this
			if val == "never"  {
				if s.Lambdo.DeleteVM(ID) != nil {
					s.Redis.Del(*s.Context, ID)
				}
			} else {
				// if it hasn't been used for {hoursTimeout} we delete it
				lastUsed, err := time.Parse(time.UnixDate, val)
				if err != nil || now.Sub(lastUsed).Abs().Hours() >= hoursTimeout {
					if s.Lambdo.DeleteVM(ID) != nil {
						s.Redis.Del(*s.Context, ID)
					}
				}
			}
		}
		keys, cursor = s.Redis.Scan(*s.Context, cursor, "*", 10).Val()
	}
}
