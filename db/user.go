package db

type user struct {
	ID       int
	Username string
	Password string
	Token    string
}

func (db DB) User(username string) (*user, error) {
	u := new(user)
	err := db.QueryRow("SELECT _rowid_, username, password, token FROM users WHERE username=?", username).
		Scan(&u.ID, &u.Username, &u.Password, &u.Token)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (db DB) Usernames() ([]string, error) {
	rows, err := db.Query(`SELECT username FROM users`)
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

func (db DB) UsernameByToken(token string) (string, error) {
	username := new(string)
	err := db.QueryRow(`SELECT username FROM users WHERE token=?`, token).Scan(username)
	if err != nil {
		return "", err
	}
	return *username, nil
}

func (db DB) EditToken(id int, token string) error {
	_, err := db.Exec("UPDATE users SET token=? WHERE _rowid_=?", token, id)
	if err != nil {
		return err
	}
	return nil
}
