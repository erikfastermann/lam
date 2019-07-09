package db

type User struct {
	ID       int
	Username string
	Password string
	Token    string
}

func (db DB) User(username string) (*User, error) {
	u := new(User)
	err := db.stmts[stmtUser].stmt.QueryRow(username).
		Scan(&u.ID, &u.Username, &u.Password, &u.Token)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (db DB) Usernames() ([]string, error) {
	rows, err := db.stmts[stmtUsernames].stmt.Query()
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

func (db DB) UserByToken(token string) (*User, error) {
	u := new(User)
	err := db.stmts[stmtUserByToken].stmt.QueryRow(token).
		Scan(&u.ID, &u.Username, &u.Password, &u.Token)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (db DB) AddUser(username, password string) error {
	return db.txExec(stmtAddUser, username, password)
}

func (db DB) RemoveUser(username string) error {
	return db.txExec(stmtRemoveUser, username)
}

func (db DB) EditToken(id int, token string) error {
	return db.txExec(stmtEditToken, token, id)
}
