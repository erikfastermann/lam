package sqlite3

import (
	"context"
	"database/sql"
	"errors"

	"github.com/erikfastermann/lam/db"
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

type DB struct {
	handle *sql.DB
	stmts  map[stmtQuery]*sql.Stmt
}

func Init(ctx context.Context, path string) (db.DB, error) {
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

	queries := []struct {
		sq  stmtQuery
		str string
	}{
		{stmtAccount, `SELECT _rowid_, region, tag, ign, username, password, user,
			leaverbuster, ban, perma, password_changed, pre_30, elo
			FROM accounts
			WHERE _rowid_=?`},
		{stmtAccounts, `SELECT _rowid_, region, tag, ign, username, password, user,
			leaverbuster, ban, perma, password_changed, pre_30, elo
			FROM accounts
			ORDER BY password_changed ASC, perma ASC, region ASC, tag ASC`},
		{stmtAddAccount, `INSERT INTO accounts(region, tag, ign, username, password, user,
			leaverbuster, ban, perma, password_changed, pre_30)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`},
		{stmtRemoveAccount, `DELETE FROM accounts WHERE _rowid_=?`},
		{stmtEditAccount, `UPDATE accounts SET region=?, tag=?, ign=?, username=?, password=?, user=?,
			leaverbuster=?, ban=?, perma=?, password_changed=?, pre_30=?
			WHERE _rowid_=?`},
		{stmtEditElo, `UPDATE accounts SET elo=? WHERE _rowid_=?`},
		{stmtUser, `SELECT _rowid_, username, password, token FROM users WHERE username=?`},
		{stmtUsernames, `SELECT username FROM users`},
		{stmtUserByToken, `SELECT _rowid_, username, password, token FROM users WHERE token=?`},
		{stmtAddUser, `INSERT INTO users(username, password, token) VALUES(?, ?, '')`},
		{stmtRemoveUser, `DELETE FROM users WHERE username=?`},
		{stmtEditToken, `UPDATE users SET token=? WHERE _rowid_=?`},
	}

	stmts := make(map[stmtQuery]*sql.Stmt, 0)
	for _, query := range queries {
		stmt, err := db.PrepareContext(ctx, query.str)
		if err != nil {
			return nil, err
		}
		stmts[query.sq] = stmt
	}

	return &DB{db, stmts}, nil
}

func (db DB) Close() error {
	var errStmt error
	for _, stmt := range db.stmts {
		err := stmt.Close()
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
	return db.asTx(ctx, sq, func(stmt *sql.Stmt) error {
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
	})
}

func (db DB) asTx(ctx context.Context, sq stmtQuery, f func(*sql.Stmt) error) error {
	tx, err := db.handle.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt := tx.StmtContext(ctx, db.stmts[sq])
	if err := f(stmt); err != nil {
		errRb := tx.Rollback()
		if errRb != nil {
			return errRb
		}
		return err
	}
	return tx.Commit()
}
