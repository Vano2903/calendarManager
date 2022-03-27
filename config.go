package main

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func connectToDB() (db *sql.DB, err error) {
	return sql.Open("mysql", "root:root@tcp(localhost:3306)/calendar?parseTime=true&charset=utf8mb4")
}
