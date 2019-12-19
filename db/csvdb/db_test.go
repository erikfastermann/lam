package csvdb

import (
	"context"
	"database/sql"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/erikfastermann/lam/db"
)

func TestDB(t *testing.T) {
	var _ db.DB = &DB{}
	ctx := context.TODO()

	dir, err := ioutil.TempDir(os.TempDir(), "lam-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// d, err := Init("users.csv", "accs.csv", "ctr.csv")
	path := func(path string) string {
		return filepath.Join(dir, path)
	}
	d, err := Init(path("users.csv"), path("accs.csv"), path("ctr.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()

	for i := 0; i < 3; i++ {
		s := strconv.Itoa(i)
		username := "user" + s
		pass := "pass" + s
		if err := d.AddUser(ctx, username, pass); err != nil {
			t.Fatal(err)
		}
		u, err := d.User(ctx, username)
		if err != nil {
			t.Fatal(err)
		}
		if u.ID != i {
			t.Fatalf("id %d != %d", u.ID, i)
		}
		if u.Username != username {
			t.Fatalf("username %s != %s", u.Username, username)
		}
		if u.Password != pass {
			t.Fatalf("password %s != %s", u.Password, pass)
		}
		if u.Token != "" {
			t.Fatalf("token not empty, got %s", u.Token)
		}
	}

	names, err := d.Usernames(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(names); l != 3 {
		t.Fatalf("expected 3 usernames, got %d %q", l, names)
	}
	for i, n := range names {
		if name := "user" + strconv.Itoa(i); n != name {
			t.Fatalf("expected name %s, got %s", name, n)
		}
	}

	tok := "wjrhgfkiawjldhiajk\njsegkdh"
	if err := d.EditToken(ctx, 1, tok); err != nil {
		t.Fatal(err)
	}
	if u, err := d.UserByToken(ctx, tok); err != nil || u.Token != tok {
		t.Fatalf("expected token %s, got %s (err: %v)", tok, u.Token, err)
	}

	if err := d.RemoveUser(ctx, "user1"); err != nil {
		t.Fatal(err)
	}
	if u, err := d.User(ctx, "user1"); err != sql.ErrNoRows {
		t.Fatalf("found deleted user: %+v (err: %v)", u, err)
	}

	if _, err := d.Accounts(ctx); err != nil {
		t.Fatal(err)
	}

	accounts := []*db.Account{
		{
			ID:       0,
			Region:   "euw",
			Tag:      "blub",
			IGN:      "player0",
			Username: "p0",
			Password: "pass",
			User:     "me",
			Elo:      "Wood IV",
		},
		{
			ID:           1,
			Region:       "na",
			Tag:          "blub",
			IGN:          "player1",
			Username:     "p1",
			Password:     "pass",
			User:         "me",
			Leaverbuster: 10,
			Perma:        true,
		},
		{
			ID:              2,
			Region:          "ru",
			Tag:             "hah",
			IGN:             "player2",
			Username:        "p2",
			Password:        "pass",
			User:            "me",
			PasswordChanged: true,
			Pre30:           true,
		},
	}
	for _, a := range accounts {
		if err := d.AddAccount(ctx, a); err != nil {
			t.Fatal(err)
		}
	}
	for id := 2; id >= 0; id-- {
		if a, err := d.Account(ctx, id); err != nil || !reflect.DeepEqual(a, accounts[id]) {
			t.Fatalf("expected acc %+v, got %+v (err: %v)", accounts[id], a, err)
		}
	}

	accs, err := d.Accounts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(accounts, accs) {
		t.Fatal("accounts don't match")
	}

	if err := d.EditAccount(ctx, 1, accounts[0]); err != nil {
		t.Fatal(err)
	}
	acc, err := d.Account(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if acc.ID != accounts[1].ID {
		t.Fatal("updated id on edit")
	}
	if acc.Elo != accounts[1].Elo {
		t.Fatal("updated elo on edit")
	}
	acc.Elo = accounts[0].Elo
	acc.ID = accounts[0].ID
	if !reflect.DeepEqual(acc, accounts[0]) {
		t.Fatal("account doesn't match after edit")
	}

	elo := "Challenger"
	if err := d.EditElo(ctx, 1, elo); err != nil {
		t.Fatal(err)
	}
	acc, err = d.Account(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if acc.Elo != elo {
		t.Fatalf("EditElo: expected %s, got %s", elo, acc.Elo)
	}

	if err := d.RemoveAccount(ctx, 1); err != nil {
		t.Fatal(err)
	}
	accs, err = d.Accounts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(append(make([]*db.Account, 0), accounts[0], accounts[2]), accs) {
		t.Fatal("accounts don't match after remove")
	}

	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
}
