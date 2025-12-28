package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

func InitDB() error {
	var err error
	db, err = pgxpool.New(context.Background(),
		"postgres://cex:cexpass@localhost:5432/cexdb")
	return err
}
