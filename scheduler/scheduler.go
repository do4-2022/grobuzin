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

func (s *Scheduler) SetStatus(id string, status database.FnStatusCode) error {
	return s.Redis.HSet(
		*s.Context,
		id,
		"status",
		int(status),
	).Err()
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

func (s *Scheduler) SpawnVM(function database.Function) (fnState database.FunctionState, err error) {
	res, err := s.Lambdo.SpawnVM(function)

	if (err != nil) {
		return
	}

	stateID := fmt.Sprintf(function.ID.String(), ":", res.ID)

	fnState = database.FunctionState{ 
		ID: res.ID,
		Address: s.Lambdo.URL, 
		Port: res.Ports[0][0],
		Status: int(database.FnReady),
		LastUsed: "never",
	}

	err = s.Redis.HSet(*s.Context, stateID, fnState).Err()
	
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
