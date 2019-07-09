package db

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

func (db DB) Account(id int) (*Account, error) {
	acc := new(Account)
	err := db.stmts[stmtAccount].stmt.QueryRow(id).
		Scan(&acc.ID, &acc.Region, &acc.Tag, &acc.IGN,
			&acc.Username, &acc.Password, &acc.User, &acc.Leaverbuster,
			&acc.Ban, &acc.Perma, &acc.PasswordChanged, &acc.Pre30, &acc.Elo)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (db DB) Accounts() ([]*Account, error) {
	rows, err := db.stmts[stmtAccounts].stmt.Query()
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

func (db DB) AddAccount(acc *Account) error {
	return db.txExec(stmtAddAccount, acc.Region, acc.Tag, acc.IGN, acc.Username,
		acc.Password, acc.User, acc.Leaverbuster, acc.Ban, acc.Perma, acc.PasswordChanged, acc.Pre30)
}

func (db DB) RemoveAccount(id int) error {
	return db.txExec(stmtRemoveAccount, id)
}

func (db DB) EditAccount(id int, acc *Account) error {
	return db.txExec(stmtEditAccount, acc.Region, acc.Tag, acc.IGN, acc.Username, acc.Password,
		acc.User, acc.Leaverbuster, acc.Ban, acc.Perma, acc.PasswordChanged, acc.Pre30, id)
}

func (db DB) EditElo(id int, elo string) error {
	return db.txExec(stmtEditElo, elo, id)
}
