package db

type User struct {
	ID       int
	Username string
	Password string
	Token    string
}

func (s Service) User(username string) (*User, error) {
	u := new(User)
	err := s.DB.QueryRow("SELECT ID, Username, Password, Token FROM users WHERE Username=?", username).
		Scan(&u.ID, &u.Username, &u.Password, &u.Token)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s Service) EditToken(id int, token string) error {
	_, err := s.DB.Exec("UPDATE users SET Token=? WHERE ID=?", token, id)
	if err != nil {
		return err
	}
	return nil
}
