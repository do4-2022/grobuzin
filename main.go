package main

import (
	"os"

	"github.com/do4-2022/grobuzin/database"
	"github.com/do4-2022/grobuzin/routes"
)

func main() {
	db := database.Init()

	// get from env
	JWTSecret := os.Getenv("JWT_SECRET")
	if JWTSecret == "" {
		panic("JWT_SECRET is not set")
	}

	r := routes.GetRoutes(db, JWTSecret)
	err := r.Run()

	if err != nil {
		panic(err)
	}

}
