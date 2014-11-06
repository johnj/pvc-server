package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

func GetDB() *sql.DB {
	db, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		config.Postgresql.User, config.Postgresql.Password,
		config.Postgresql.Host, config.Postgresql.Dbname))
	if err != nil {
		panic(err)
	}
	return db
}
