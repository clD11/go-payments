package app

import (
	"github.com/go-pg/pg"
)

type Config struct {
	DB *pg.Options
}
