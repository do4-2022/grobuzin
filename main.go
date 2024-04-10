package main

import (
	"log"

	"github.com/do4-2022/grobuzin/database"
	"github.com/do4-2022/grobuzin/routes"

	"github.com/caarlos0/env/v10"
)

import _ "github.com/joho/godotenv/autoload"

type Config struct {
	// rootFsStorageDSN string `env:"ROOT_FS_STORAGE_DSN,notEmpty"`
	// VMStorageDSN string `env:"VM_STORAGE_DSN,notEmpty"`
	FuntionStateStorageDSN string `env:"FUNCTION_STATE_STORAGE_DSN,notEmpty" envDefault:"host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=Asia/Shanghai"`
	JWTSecret string `env:"JWT_SECRET,notEmpty"`
}

func main() {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("%+v\n", err)
	}

	db := database.Init(cfg.FuntionStateStorageDSN)
	r := routes.GetRoutes(db, cfg.JWTSecret)

	err := r.Run()

	if err != nil {
		panic(err)
	}
}
