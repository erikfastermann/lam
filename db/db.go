package db

import (
	"database/sql"
	"fmt"
)

type Service struct {
	DB *sql.DB
}

func New(user, password, address, name string) (Service, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, address, name))
	if err != nil {
		return Service{}, err
	}
	err = db.Ping()
	if err != nil {
		return Service{}, err
	}
	return Service{DB: db}, err
}
