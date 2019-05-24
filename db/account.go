package db

import "github.com/go-sql-driver/mysql"

type Account struct {
	ID              int
	Region          string
	Tag             string
	IGN             string
	Username        string
	Password        string
	User            string
	Leaverbuster    int
	Ban             mysql.NullTime
	Perma           bool
	PasswordChanged bool
	Pre30           bool
	Elo             string
}

func (s Service) GetAccount(id int) (*Account, error) {
	acc := new(Account)
	err := s.DB.QueryRow(`SELECT ID, Region, Tag, Ign, Username, Password, User,
        Leaverbuster, Ban, Perma, PasswordChanged, Pre30, Elo FROM accounts WHERE ID=?`, id).
		Scan(&acc.ID, &acc.Region, &acc.Tag, &acc.IGN,
			&acc.Username, &acc.Password, &acc.User, &acc.Leaverbuster,
			&acc.Ban, &acc.Perma, &acc.PasswordChanged, &acc.Pre30, &acc.Elo)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (s Service) GetAccounts() ([]*Account, error) {
	rows, err := s.DB.Query(`SELECT ID, Region, Tag, Ign, Username, Password, User,
        Leaverbuster, Ban, Perma, PasswordChanged, Pre30, Elo FROM accounts`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accs := make([]*Account, 0)
	for rows.Next() {
		acc := new(Account)
		err := rows.Scan(&acc.ID, &acc.Region, &acc.Tag, &acc.IGN,
			&acc.Username, &acc.Password, &acc.User, &acc.Leaverbuster,
			&acc.Ban, &acc.Perma, &acc.PasswordChanged, &acc.Pre30, &acc.Elo)
		if err != nil {
			return nil, err
		}
		accs = append(accs, acc)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return accs, nil
}

func (s Service) CreateAccount(acc Account) error {
	_, err := s.DB.Exec(`INSERT INTO accounts(Region, Tag, Ign, Username, 
		Password, User, Leaverbuster, Ban, Perma, PasswordChanged, Pre30)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, acc.Region, acc.Tag, acc.IGN, acc.Username,
		acc.Password, acc.User, acc.Leaverbuster, acc.Ban, acc.Perma, acc.PasswordChanged, acc.Pre30)
	if err != nil {
		return err
	}
	return nil
}

func (s Service) SetAccount(id int, acc Account) error {
	_, err := s.DB.Exec(`UPDATE accounts SET Region=?, Tag=?, Ign=?, Username=?, Password=?,
		User=?, Leaverbuster=?, Ban=?, Perma=?, PasswordChanged=?, Pre30=? WHERE ID=?`,
		acc.Region, acc.Tag, acc.IGN, acc.Username, acc.Password,
		acc.User, acc.Leaverbuster, acc.Ban, acc.Perma, acc.PasswordChanged, acc.Pre30, id)
	if err != nil {
		return err
	}
	return nil
}
