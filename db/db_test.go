package db

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDB(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "lam-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// d, err := Init("users.csv", "accs.csv", "ctr.csv")
	path := func(path string) string {
		return filepath.Join(dir, path)
	}
	d, err := Init(path("accs.csv"), path("ctr.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()

	if _, err := d.Accounts(); err != nil {
		t.Fatal(err)
	}

	accounts := []*Account{
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
		if err := d.AddAccount(a); err != nil {
			t.Fatal(err)
		}
	}
	for id := 2; id >= 0; id-- {
		if a, err := d.Account(id); err != nil || !reflect.DeepEqual(a, accounts[id]) {
			t.Fatalf("expected acc %+v, got %+v (err: %v)", accounts[id], a, err)
		}
	}

	accs, err := d.Accounts()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(accounts, accs) {
		t.Fatal("accounts don't match")
	}

	if err := d.EditAccount(1, accounts[0]); err != nil {
		t.Fatal(err)
	}
	acc, err := d.Account(1)
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
	if err := d.EditElo(1, elo); err != nil {
		t.Fatal(err)
	}
	acc, err = d.Account(1)
	if err != nil {
		t.Fatal(err)
	}
	if acc.Elo != elo {
		t.Fatalf("EditElo: expected %s, got %s", elo, acc.Elo)
	}

	if err := d.RemoveAccount(1); err != nil {
		t.Fatal(err)
	}
	accs, err = d.Accounts()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(append(make([]*Account, 0), accounts[0], accounts[2]), accs) {
		t.Fatal("accounts don't match after remove")
	}

	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
}
