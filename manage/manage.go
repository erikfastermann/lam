package main

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/erikfastermann/lam/db"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	dbPath := "/var/lib/docker/volumes/lam_db/_data/lam.db"
	switch len(os.Args) {
	default:
		usage()
	case 2:
	case 3:
		dbPath = os.Args[2]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := db.Init(ctx, dbPath)
	checkErr(err)
	defer db.Close()

	switch os.Args[1] {
	case "-l":
		usernames, err := db.Usernames(ctx)
		checkErr(err)
		for _, i := range usernames {
			u, err := db.User(ctx, i)
			checkErr(err)
			fmt.Printf("%d - %s - %s - %s\n", u.ID, u.Username, u.Password, u.Token)
		}
	case "-a":
		username := getUsername()
		password := getPassword()
		checkErr(db.AddUser(ctx, username, password))
	case "-r":
		username := getUsername()
		err := db.RemoveUser(ctx, username)
		if err == sql.ErrNoRows {
			checkErr(errors.New("username doesn't exist"))
		}
		checkErr(err)
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `USAGE: `+os.Args[0]+` COMMAND DATABASE-PATH
	-l	list all users with id, username, password-hash and session-cookie
	-a	add a user, username and password are read from STDIN
	-r	remove users by username (read from STDIN)`)
	os.Exit(1)
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getUsername() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Username: ")
	username, err := reader.ReadString('\n')
	checkErr(err)
	return strings.TrimSpace(username)
}

func getPassword() string {
	fmt.Print("Enter password: ")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Print("\n")
	checkErr(err)
	hash, err := bcrypt.GenerateFromPassword(password, 14)
	checkErr(err)
	return string(hash)
}
