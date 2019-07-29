package db

import (
	"context"
	"database/sql/driver"
	"fmt"
	"time"
)

type DB interface {
	User(ctx context.Context, username string) (*User, error)
	Usernames(ctx context.Context) ([]string, error)
	UserByToken(ctx context.Context, token string) (*User, error)
	AddUser(ctx context.Context, username, password string) error
	RemoveUser(ctx context.Context, username string) error
	EditToken(ctx context.Context, id int, token string) error

	Account(ctx context.Context, id int) (*Account, error)
	Accounts(ctx context.Context) ([]*Account, error)
	AddAccount(ctx context.Context, acc *Account) error
	RemoveAccount(ctx context.Context, id int) error
	EditAccount(ctx context.Context, id int, acc *Account) error
	EditElo(ctx context.Context, id int, elo string) error

	Close() error
}

type User struct {
	ID       int
	Username string
	Password string
	Token    string
}

type Account struct {
	ID              int
	Region          string
	Tag             string
	IGN             string
	Username        string
	Password        string
	User            string
	Leaverbuster    int
	Ban             NullTime
	Perma           bool
	PasswordChanged bool
	Pre30           bool
	Elo             string
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
