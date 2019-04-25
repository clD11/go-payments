package main

import (
	"github.com/clD11/form3-payments/app"
	"github.com/go-pg/pg"
)

func main() {
	config := app.Config{
		DB: &pg.Options{
			Addr:     "postgres:5432",
			Database: "postgres",
			User:     "postgres",
			Password: "postgres",
		},
	}
	a := &app.App{}
	a.Initialize(&config)
	a.Run(":8080")
}
