package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/erikfastermann/lam/db"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	if len(os.Args) > 3 {
		usage()
	}

	dbPath := "/var/lib/docker/volumes/lam_db/_data/lam.db"
	if len(os.Args) == 3 {
		dbPath = os.Args[2]
	}
	db, err := db.Init(dbPath)
	checkErr(err)

	switch os.Args[1] {
	case "-l":
		usernames, err := db.Usernames()
		checkErr(err)
		for _, i := range usernames {
			u, err := db.User(i)
			checkErr(err)
			fmt.Printf("%d - %s - %s - %s\n", u.ID, u.Username, u.Password, u.Token)
		}
	case "-a":
		username := getUsername()
		password := getPassword()
		checkErr(err)
		_, err := db.User(username)
		if err != nil {
			checkErr(db.AddUser(username, password))
			return
		}
		checkErr(errors.New("username already exists"))
	case "-r":
		username := getUsername()
		_, err := db.User(username)
		if err != nil {
			checkErr(fmt.Errorf("%v, username doesn't exist", err))
		}
		checkErr(db.RemoveUser(username))
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
