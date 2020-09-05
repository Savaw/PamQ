package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

//TODO use env variable
const (
	DB_USER = "postgres"
	DB_NAME = "go_db_test"
)

func InitDB() {
	dbinfo := fmt.Sprintf("user=%s dbname=%s sslmode=disable", DB_USER, DB_NAME)

	var err error
	DB, err = sql.Open("postgres", dbinfo)
	if err != nil {
		log.Fatal(err)
	}
}
