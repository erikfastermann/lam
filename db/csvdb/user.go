package csvdb

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/erikfastermann/lam/db"
)

const (
	uID       = 0
	uUsername = 1
	uPassword = 2
	uToken    = 3
	uLen      = 4
)

func (d *DB) User(ctx context.Context, username string) (*db.User, error) {
	users, err := d.users.all(ctx)
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		if u[uUsername] == username {
			return recordToUser(u)
		}
	}
	return nil, sql.ErrNoRows
}

func (d *DB) Usernames(ctx context.Context) ([]string, error) {
	users, err := d.users.all(ctx)
	if err != nil {
		return nil, err
	}
	usernames := make([]string, 0)
	for _, u := range users {
		usernames = append(usernames, u[uUsername])
	}
	return usernames, nil
}

func (d *DB) UserByToken(ctx context.Context, token string) (*db.User, error) {
	users, err := d.users.all(ctx)
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		if u[uToken] == token {
			return recordToUser(u)
		}
	}
	return nil, sql.ErrNoRows
}

func (d *DB) AddUser(ctx context.Context, username, password string) error {
	id, err := bumpCtr(ctx, d.ctr, ctrPosUser)
	if err != nil {
		return err
	}

	return d.users.insert(ctx, userToRecord(&db.User{
		ID:       id,
		Username: username,
		Password: password,
	}))
}

func (d *DB) RemoveUser(ctx context.Context, username string) error {
	return d.users.update(ctx, func(users [][]string) ([][]string, error) {
		for i, u := range users {
			if u[uUsername] == username {
				users[i] = users[len(users)-1]
				users = users[:len(users)-1]
				return users, nil
			}
		}
		return nil, sql.ErrNoRows
	})
}

func (d *DB) EditToken(ctx context.Context, id int, token string) error {
	idStr := strconv.Itoa(id)
	return d.users.update(ctx, func(users [][]string) ([][]string, error) {
		for i, u := range users {
			if u[uID] == idStr {
				users[i][uToken] = token
				return users, nil
			}
		}
		return nil, sql.ErrNoRows
	})
}

func userToRecord(u *db.User) []string {
	s := make([]string, uLen)
	s[uID] = strconv.Itoa(u.ID)
	s[uUsername] = u.Username
	s[uPassword] = u.Password
	s[uToken] = u.Token
	return s
}

func recordToUser(r []string) (*db.User, error) {
	id, err := strconv.Atoi(r[uID])
	if err != nil {
		return nil, err
	}
	return &db.User{
		ID:       id,
		Username: r[uUsername],
		Password: r[uPassword],
		Token:    r[uToken],
	}, nil
}
