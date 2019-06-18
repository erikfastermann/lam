package db

import (
	"github.com/go-sql-driver/mysql"
)

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

func (db DB) Account(id int) (*Account, error) {
	acc := new(Account)
	err := db.QueryRow(`SELECT _rowid_, region, tag, ign, username, password, user,
        leaverbuster, ban, perma, password_changed, pre_30, elo FROM accounts WHERE _rowid_=?`, id).
		Scan(&acc.ID, &acc.Region, &acc.Tag, &acc.IGN,
			&acc.Username, &acc.Password, &acc.User, &acc.Leaverbuster,
			&acc.Ban, &acc.Perma, &acc.PasswordChanged, &acc.Pre30, &acc.Elo)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (db DB) Accounts() ([]*Account, error) {
	rows, err := db.Query(`SELECT _rowid_, region, tag, ign, username, password, user,
		leaverbuster, ban, perma, password_changed, pre_30, elo FROM accounts
		ORDER BY password_changed ASC, perma ASC, region ASC, tag ASC`)
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
	_, err := db.Exec(`INSERT INTO accounts(region, tag, ign, username,
	 password, user, leaverbuster, ban, perma, password_changed, pre_30)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, acc.Region, acc.Tag, acc.IGN, acc.Username,
		acc.Password, acc.User, acc.Leaverbuster, acc.Ban, acc.Perma, acc.PasswordChanged, acc.Pre30)
	if err != nil {
		return err
	}
	return nil
}

func (db DB) EditAccount(id int, acc *Account) error {
	_, err := db.Exec(`UPDATE accounts SET region=?, tag=?, ign=?, username=?, password=?,
		user=?, leaverbuster=?, ban=?, perma=?, password_changed=?, pre_30=? WHERE _rowid_=?`,
		acc.Region, acc.Tag, acc.IGN, acc.Username, acc.Password,
		acc.User, acc.Leaverbuster, acc.Ban, acc.Perma, acc.PasswordChanged, acc.Pre30, id)
	if err != nil {
		return err
	}
	return nil
}

func (db DB) EditElo(id int, elo string) error {
	_, err := db.Exec("UPDATE accounts SET elo=? WHERE _rowid_=?", elo, id)
	if err != nil {
		return err
	}
	return nil
}
