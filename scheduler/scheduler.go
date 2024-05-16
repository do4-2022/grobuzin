package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"
	"github.com/do4-2022/grobuzin/database"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var ErrRecordNotFound = errors.New("could not find an available function")

type Scheduler struct {
	Redis *redis.Client
	Context *context.Context
	Lambdo *LambdoService
}

// Get all keys of a state by its ID 
func (s *Scheduler) GetStateByID(id string) (fnState database.FunctionState, err error) {
	err = s.Redis.HGetAll(
		*s.Context,
		id,
	).Scan(&fnState)

	return
}

// Uses SCAN to look for the first instance marked as ready (https://redis.io/docs/latest/commands/scan/)
func (s *Scheduler) LookForReadyInstance(functionId uuid.UUID, cursor uint64) (fnState database.FunctionState, returnedCursor uint64, err error) {
	// match rule to get all instances of a function 
	stateMatch := fmt.Sprintf(functionId.String(), ":*")

	keys, returnedCursor, err := s.Redis.Scan(*s.Context, cursor, stateMatch, 10).Result()

	if err != nil {
		return
	}

	for _, id := range keys {
		err := s.Redis.HGetAll(*s.Context, id).Scan(&fnState);

		if err == nil && database.FnStatusCode(fnState.Status) == database.FnReady {
			// we found a ready instance so we return it
			return fnState, 0, nil
		}
	}

	if returnedCursor != 0 {
		// the current sweep did not find anything, we try again 
		return s.LookForReadyInstance(functionId, returnedCursor)
	}

	return fnState, 0, ErrRecordNotFound  // we did not find anything thus, id is empty
} 

func (s *Scheduler) SpawnVM(functionId uuid.UUID) (LambdoRunResponse LambdoSpawnResponse, err error) {
	res, err := s.Lambdo.SpawnVM(functionId)

	if (err != nil) {
		return
	}

	stateID := fmt.Sprintf(functionId.String(), ":", res.ID)
	
	err = s.Redis.HSet(*s.Context, stateID, &database.FunctionState{ 
		ID: res.ID,
		Address: res.Address, 
		Port: res.Port,
		Status: database.FnCreating,
		LastUsed: "never",
	}).Err()
	
	return
}

func (s *Scheduler) GetFunctionStates(functionId uuid.UUID) (states []database.FunctionState) {
	stateQuery := fmt.Sprintf(functionId.String(), ":*")
	firstsweep, cursor := true, uint64(0)
	
	// because do-while does not exists in these lands
	for !firstsweep && cursor != 0 {
		firstsweep = false

		var keys []string 
		keys, cursor = s.Redis.Scan(*s.Context, cursor, stateQuery, 10).Val()
		
		// for each key we got, we retrieve all of it's information
		for _, ID := range keys {
			var state database.FunctionState
			
			if s.Redis.HGetAll(*s.Context, ID).Scan(&state) != nil {
				return
			}

			states = append(states, state)
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
