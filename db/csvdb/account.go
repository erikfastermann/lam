package csvdb

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/erikfastermann/lam/db"
)

const (
	aID              = 0
	aRegion          = 1
	aTag             = 2
	aIGN             = 3
	aUsername        = 4
	aPassword        = 5
	aUser            = 6
	aLeaverbuster    = 7
	aBan             = 8
	aPerma           = 9
	aPasswordChanged = 10
	aPre30           = 11
	aElo             = 12
	aLen             = 13
)

func (d *DB) Account(_ context.Context, id int) (*db.Account, error) {
	accs, err := d.accs.all()
	if err != nil {
		return nil, err
	}
	idStr := strconv.Itoa(id)
	for _, a := range accs {
		if a[aID] == idStr {
			return recordToAcc(a)
		}
	}
	return nil, sql.ErrNoRows
}

// TODO: ORDER BY password_changed ASC, perma ASC, region ASC, tag ASC
func (d *DB) Accounts(_ context.Context) ([]*db.Account, error) {
	records, err := d.accs.all()
	if err != nil {
		return nil, err
	}
	accs := make([]*db.Account, 0)
	for _, acc := range records {
		a, err := recordToAcc(acc)
		if err != nil {
			return nil, err
		}
		accs = append(accs, a)
	}
	return accs, nil
}

func (d *DB) AddAccount(_ context.Context, acc *db.Account) error {
	id, err := bumpCtr(d.ctr, ctrPosAcc)
	if err != nil {
		return err
	}

	acc.ID = id
	return d.accs.insert(accToRecord(acc))
}

func (d *DB) RemoveAccount(_ context.Context, id int) error {
	idStr := strconv.Itoa(id)
	return d.accs.update(func(accs [][]string) ([][]string, error) {
		for i, a := range accs {
			if a[aID] == idStr {
				accs[i] = accs[len(accs)-1]
				accs = accs[:len(accs)-1]
				return accs, nil
			}
		}
		return nil, sql.ErrNoRows
	})
}

func (d *DB) EditAccount(_ context.Context, id int, acc *db.Account) error {
	idStr := strconv.Itoa(id)
	return d.accs.update(func(accs [][]string) ([][]string, error) {
		for i, a := range accs {
			if a[aID] == idStr {
				elo := accs[i][aElo]
				accs[i] = accToRecord(acc)
				accs[i][aID] = idStr
				accs[i][aElo] = elo
				return accs, nil
			}
		}
		return nil, sql.ErrNoRows
	})
}

func (d *DB) EditElo(_ context.Context, id int, elo string) error {
	idStr := strconv.Itoa(id)
	return d.accs.update(func(accs [][]string) ([][]string, error) {
		for i, a := range accs {
			if a[aID] == idStr {
				accs[i][aElo] = elo
				return accs, nil
			}
		}
		return nil, sql.ErrNoRows
	})
}

const (
	nullTime   = "null"
	timeFormat = time.RFC3339
)

func accToRecord(a *db.Account) []string {
	ban := nullTime
	if a.Ban.Valid {
		ban = a.Ban.Time.Format(timeFormat)
	}

	s := make([]string, aLen)
	s[aID] = strconv.Itoa(a.ID)
	s[aRegion] = a.Region
	s[aTag] = a.Tag
	s[aIGN] = a.IGN
	s[aUsername] = a.Username
	s[aPassword] = a.Password
	s[aUser] = a.User
	s[aLeaverbuster] = strconv.Itoa(a.Leaverbuster)
	s[aBan] = ban
	s[aPerma] = strconv.FormatBool(a.Perma)
	s[aPasswordChanged] = strconv.FormatBool(a.PasswordChanged)
	s[aPre30] = strconv.FormatBool(a.Pre30)
	s[aElo] = a.Elo
	return s
}

func recordToAcc(r []string) (*db.Account, error) {
	id, err := strconv.Atoi(r[aID])
	if err != nil {
		return nil, err
	}
	leaverbuster, err := strconv.Atoi(r[aLeaverbuster])
	if err != nil {
		return nil, err
	}
	perma, err := strconv.ParseBool(r[aPerma])
	if err != nil {
		return nil, err
	}
	passwordChanged, err := strconv.ParseBool(r[aPasswordChanged])
	if err != nil {
		return nil, err
	}
	pre30, err := strconv.ParseBool(r[aPre30])
	if err != nil {
		return nil, err
	}

	ban := db.NullTime{}
	banStr := r[aBan]
	if strings.ToLower(banStr) != nullTime {
		t, err := time.Parse(timeFormat, banStr)
		if err != nil {
			return nil, err
		}
		ban = db.NullTime{
			Time:  t,
			Valid: true,
		}
	}

	return &db.Account{
		ID:              id,
		Region:          r[aRegion],
		Tag:             r[aTag],
		IGN:             r[aIGN],
		Username:        r[aUsername],
		Password:        r[aPassword],
		User:            r[aUser],
		Leaverbuster:    leaverbuster,
		Ban:             ban,
		Perma:           perma,
		PasswordChanged: passwordChanged,
		Pre30:           pre30,
		Elo:             r[aElo],
	}, nil
}
