package db

import "context"

type User struct {
	ID       int
	Username string
	Password string
	Token    string
}

func (db DB) User(ctx context.Context, username string) (*User, error) {
	u := new(User)
	err := db.stmts[stmtUser].stmt.QueryRowContext(ctx, username).
		Scan(&u.ID, &u.Username, &u.Password, &u.Token)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (db DB) Usernames(ctx context.Context) ([]string, error) {
	rows, err := db.stmts[stmtUsernames].stmt.QueryContext(ctx)
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

func (db DB) UserByToken(ctx context.Context, token string) (*User, error) {
	u := new(User)
	err := db.stmts[stmtUserByToken].stmt.QueryRowContext(ctx, token).
		Scan(&u.ID, &u.Username, &u.Password, &u.Token)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (db DB) AddUser(ctx context.Context, username, password string) error {
	return db.txExec(ctx, stmtAddUser, username, password)
}

func (db DB) RemoveUser(ctx context.Context, username string) error {
	return db.txExec(ctx, stmtRemoveUser, username)
}

func (db DB) EditToken(ctx context.Context, id int, token string) error {
	return db.txExec(ctx, stmtEditToken, token, id)
}
