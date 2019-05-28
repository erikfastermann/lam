package db

type user struct {
	ID       int
	Username string
	Password string
	Token    string
}

func (s Service) User(username string) (*user, error) {
	u := new(user)
	err := s.DB.QueryRow("SELECT ID, Username, Password, Token FROM users WHERE Username=?", username).
		Scan(&u.ID, &u.Username, &u.Password, &u.Token)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s Service) Usernames() ([]string, error) {
	rows, err := s.DB.Query(`SELECT Username FROM users`)
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

func (s Service) EditToken(id int, token string) error {
	_, err := s.DB.Exec("UPDATE users SET Token=? WHERE ID=?", token, id)
	if err != nil {
		return err
	}
	return nil
}
