package sqlite3

import (
	"context"

	"github.com/erikfastermann/lam/db"
)

func (sqlDB DB) User(ctx context.Context, username string) (*db.User, error) {
	u := new(db.User)
	err := sqlDB.stmts[stmtUser].stmt.QueryRowContext(ctx, username).
		Scan(&u.ID, &u.Username, &u.Password, &u.Token)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (sqlDB DB) Usernames(ctx context.Context) ([]string, error) {
	rows, err := sqlDB.stmts[stmtUsernames].stmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	usernames := make([]string, 0)
	username := new(string)
	for rows.Next() {
		err := rows.Scan(username)
		if err != nil {
			return nil, err
		}
		usernames = append(usernames, *username)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return usernames, err
}

func (sqlDB DB) UserByToken(ctx context.Context, token string) (*db.User, error) {
	u := new(db.User)
	err := sqlDB.stmts[stmtUserByToken].stmt.QueryRowContext(ctx, token).
		Scan(&u.ID, &u.Username, &u.Password, &u.Token)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (sqlDB DB) AddUser(ctx context.Context, username, password string) error {
	return sqlDB.txExec(ctx, stmtAddUser, username, password)
}

func (sqlDB DB) RemoveUser(ctx context.Context, username string) error {
	return sqlDB.txExec(ctx, stmtRemoveUser, username)
}

func (sqlDB DB) EditToken(ctx context.Context, id int, token string) error {
	return sqlDB.txExec(ctx, stmtEditToken, token, id)
}
