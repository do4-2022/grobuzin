package database

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FunctionState struct {
	gorm.Model
	ID      uuid.UUID `json:"id" ,gorm:"primarykey;type:uuid;default:gen_random_uuid()"`
	Status  string    `json:"status"`
	Address string    `json:"address"`
	Port    uint16    `json:"port"`

	FunctionID uuid.UUID `json:"function_id"`
}

func (FunctionState) TableName() string {
	return "function_states"
}
