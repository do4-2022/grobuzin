package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// put
var (
	models = []interface{}{&User{}}
)

func Init() *gorm.DB {

	// Code to initialize database connection

	dsn := "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=Asia/Shanghai"

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
