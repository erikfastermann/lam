package db

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func Init(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		username VARCHAR(24) NOT NULL UNIQUE,
		password VARCHAR(63) NOT NULL,
		token VARCHAR(60) NOT NULL
	)`)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS accounts (
		region varchar(4) NOT NULL,
		tag varchar(24) NOT NULL,
		ign varchar(16) NOT NULL,
		username varchar(24) NOT NULL,
		password varchar(63) NOT NULL,
		user varchar(24) NOT NULL,
		leaverbuster int(2) NOT NULL,
		ban datetime DEFAULT NULL,
		perma boolean NOT NULL,
		password_changed boolean NOT NULL,
		pre_30 boolean NOT NULL,
		elo varchar(24) NOT NULL DEFAULT 'Not parsed'
	)`)
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db DB) txExec(query string, args ...interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	err = func() error {
		result, err := tx.Exec(query, args...)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		switch rows {
		case 0:
			return sql.ErrNoRows
		case 1:
			return nil
		default:
			return errors.New("more than 1 row would be affected in a query")
		}
	}()

	if err != nil {
		errRb := tx.Rollback()
		if errRb != nil {
			return errRb
		}
		return err
	}
	return tx.Commit()
}
