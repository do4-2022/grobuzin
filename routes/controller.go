package routes

import "gorm.io/gorm"

// Controller for the whole API, all endpoints should be implemented on this struct.
// This allows to have a single point of access to the database and other services.
type Controller struct {
	DB *gorm.DB
}
