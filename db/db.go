package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type Service struct {
	DB *sql.DB
}

func New(user, password, address, name string) (*Service, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, address, name))
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
	    ID INT NOT NULL AUTO_INCREMENT,
	    Username VARCHAR(24) CHARACTER SET utf8 NOT NULL,
	    Password VARCHAR(255) CHARACTER SET utf8 NOT NULL,
	    Token VARCHAR(255) CHARACTER SET utf8 NOT NULL,
	    PRIMARY KEY (ID)
	)`)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS accounts (
	    ID INT NOT NULL AUTO_INCREMENT,
	    Region VARCHAR(4) CHARACTER SET utf8 NOT NULL,
	    Tag VARCHAR(255) CHARACTER SET utf8 NOT NULL,
	    Ign VARCHAR(16) CHARACTER SET utf8 NOT NULL,
	    Username VARCHAR(24) CHARACTER SET utf8 NOT NULL,
	    Password VARCHAR(255) CHARACTER SET utf8 NOT NULL,
	    User VARCHAR(24) CHARACTER SET utf8 NOT NULL,
	    Leaverbuster INT NOT NULL,
	    Ban DATETIME,
	    Perma BOOLEAN NOT NULL,
	    PasswordChanged BOOLEAN NOT NULL,
	    Pre30 BOOLEAN NOT NULL,
	    Elo VARCHAR(24) CHARACTER SET utf8 NOT NULL DEFAULT "Not parsed",
	    PRIMARY KEY (ID)
	)`)
	if err != nil {
		return nil, err
	}

	return &Service{DB: db}, nil
}
