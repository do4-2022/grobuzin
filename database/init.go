package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// put
var (
	models = []interface{}{&User{}, Function{}, &FunctionState{}}
)

func Init(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(models...)

	if err != nil {
		panic(err)
	}

	return db
}
