package main

import(
	"fmt"
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

var db *sql.DB

const (
	DB_USER = "postgres"
	DB_NAME = "go_db_test"
)

func GetDB() *sql.DB {
	dbinfo := fmt.Sprintf("user=%s dbname=%s sslmode=disable", DB_USER, DB_NAME)

	var err error
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		log.Fatal(err)
	}

	return db
}