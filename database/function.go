package database

import (
	"github.com/google/uuid"
)

type Function struct {
	ID          uuid.UUID `json:"id" ,gorm:"primarykey;type:uuid;default:gen_random_uuid()"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Language    string    `json:"language"`
	Built       bool      `json:"built"` // The builder has built the image for this function
}
