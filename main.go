package main

import (
	"github.com/do4-2022/grobuzin/database"
	"github.com/do4-2022/grobuzin/routes"
)

func main() {
	db := database.Init()

	r := routes.GetRoutes(db)
	r.Run()
}
