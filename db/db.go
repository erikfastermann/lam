package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type stmtQuery int

const (
	stmtAccount stmtQuery = iota
	stmtAccounts
	stmtAddAccount
	stmtRemoveAccount
	stmtEditAccount
	stmtEditElo
	stmtUser
	stmtUsernames
	stmtUserByToken
	stmtAddUser
	stmtRemoveUser
	stmtEditToken
)

type prepStmt struct {
	query string
	stmt  *sql.Stmt
}

type DB struct {
	handle *sql.DB
	stmts  map[stmtQuery]prepStmt
}

func Init(ctx context.Context, path string) (*DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS users (
		username VARCHAR(24) NOT NULL UNIQUE,
		password VARCHAR(63) NOT NULL,
		token VARCHAR(60) NOT NULL
	)`)
	if err != nil {
		return nil, err
	}
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS accounts (
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

	stmts := map[stmtQuery]prepStmt{
		stmtAccount: prepStmt{
			query: `SELECT _rowid_, region, tag, ign, username, password, user,
				leaverbuster, ban, perma, password_changed, pre_30, elo
				FROM accounts
				WHERE _rowid_=?`,
		},
		stmtAccounts: prepStmt{
			query: `SELECT _rowid_, region, tag, ign, username, password, user,
				leaverbuster, ban, perma, password_changed, pre_30, elo
				FROM accounts
				ORDER BY password_changed ASC, perma ASC, region ASC, tag ASC`,
		},
		stmtAddAccount: prepStmt{
			query: `INSERT INTO accounts(region, tag, ign, username, password, user,
				leaverbuster, ban, perma, password_changed, pre_30)
				VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		},
		stmtRemoveAccount: prepStmt{
			query: `DELETE FROM accounts WHERE _rowid_=?`,
		},
		stmtEditAccount: prepStmt{
			query: `UPDATE accounts SET region=?, tag=?, ign=?, username=?, password=?, user=?,
				leaverbuster=?, ban=?, perma=?, password_changed=?, pre_30=?
				WHERE _rowid_=?`,
		},
		stmtEditElo: prepStmt{
			query: `UPDATE accounts SET elo=? WHERE _rowid_=?`,
		},
		stmtUser: prepStmt{
			query: `SELECT _rowid_, username, password, token FROM users WHERE username=?`,
		},
		stmtUsernames: prepStmt{
			query: `SELECT username FROM users`,
		},
		stmtUserByToken: prepStmt{
			query: `SELECT _rowid_, username, password, token FROM users WHERE token=?`,
		},
		stmtAddUser: prepStmt{
			query: `INSERT INTO users(username, password, token) VALUES(?, ?, '')`,
		},
		stmtRemoveUser: prepStmt{
			query: `DELETE FROM users WHERE username=?`,
		},
		stmtEditToken: prepStmt{
			query: `UPDATE users SET token=? WHERE _rowid_=?`,
		},
	}

	for key, val := range stmts {
		stmt, err := db.PrepareContext(ctx, val.query)
		if err != nil {
			return nil, err
		}
		stmts[key] = prepStmt{val.query, stmt}
	}

	return &DB{db, stmts}, nil
}

func (db DB) Close() error {
	var errStmt error
	for _, val := range db.stmts {
		err := val.stmt.Close()
		if err != nil {
			errStmt = err
		}
	}
	errHandle := db.handle.Close()
	if errHandle != nil {
		return errHandle
	}
	return errStmt
}

func (db DB) txExec(ctx context.Context, sq stmtQuery, args ...interface{}) error {
	tx, err := db.handle.Begin()
	if err != nil {
		return err
	}

	err = func() error {
		stmt := tx.StmtContext(ctx, db.stmts[sq].stmt)
		if err != nil {
			return err
		}
		result, err := stmt.ExecContext(ctx, args...)
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

type NullTime struct {
	Time  time.Time
	Valid bool
}

func (nt *NullTime) Scan(value interface{}) error {
	if value == nil {
		nt.Time = time.Time{}
		nt.Valid = false
		return nil
	}
	t, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("%T is not nil or time.Time", value)
	}
	nt.Time = t
	nt.Valid = true
	return nil
}

func (nt NullTime) Value() (driver.Value, error) {
	if nt.Valid {
		return nt.Time, nil
	}
	return nil, nil
}
