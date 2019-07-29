package sqlite3

import (
	"context"

	"github.com/erikfastermann/lam/db"
)

func (sqlDB DB) Account(ctx context.Context, id int) (*db.Account, error) {
	acc := new(db.Account)
	err := sqlDB.stmts[stmtAccount].stmt.QueryRowContext(ctx, id).
		Scan(&acc.ID, &acc.Region, &acc.Tag, &acc.IGN,
			&acc.Username, &acc.Password, &acc.User, &acc.Leaverbuster,
			&acc.Ban, &acc.Perma, &acc.PasswordChanged, &acc.Pre30, &acc.Elo)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (sqlDB DB) Accounts(ctx context.Context) ([]*db.Account, error) {
	rows, err := sqlDB.stmts[stmtAccounts].stmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accs := make([]*db.Account, 0)
	for rows.Next() {
		acc := new(db.Account)
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

func (sqlDB DB) AddAccount(ctx context.Context, acc *db.Account) error {
	return sqlDB.txExec(ctx, stmtAddAccount, acc.Region, acc.Tag, acc.IGN, acc.Username,
		acc.Password, acc.User, acc.Leaverbuster, acc.Ban, acc.Perma, acc.PasswordChanged, acc.Pre30)
}

func (sqlDB DB) RemoveAccount(ctx context.Context, id int) error {
	return sqlDB.txExec(ctx, stmtRemoveAccount, id)
}

func (sqlDB DB) EditAccount(ctx context.Context, id int, acc *db.Account) error {
	return sqlDB.txExec(ctx, stmtEditAccount, acc.Region, acc.Tag, acc.IGN, acc.Username, acc.Password,
		acc.User, acc.Leaverbuster, acc.Ban, acc.Perma, acc.PasswordChanged, acc.Pre30, id)
}

func (sqlDB DB) EditElo(ctx context.Context, id int, elo string) error {
	return sqlDB.txExec(ctx, stmtEditElo, elo, id)
}
