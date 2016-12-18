package utils

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"fmt"
)

const (
	//Docker host ip in my mac.
	DB_HOST = "192.168.99.100"
	DB_PORT = "3306"
	DB_USER = "root"
	DB_PASSWORD = "123456"
	DB_NAME = "VisAwesome"
)

func CheckErrorPanic(err error) {
	if err != nil {
		log.Panic(err.Error())
	} else {
		return
	}
}
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
func GetDBHandler() *sql.DB{
	var handler *sql.DB
	dbinfo := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME)
	handler, err := sql.Open("mysql", dbinfo)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("SQL Connection Success(%s:%s)\n", DB_HOST, DB_PORT)
	}
	return handler
}