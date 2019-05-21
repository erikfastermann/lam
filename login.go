package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int
	Username string
	Password string
	Token    string
}

func loginGet(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "login.html", nil)
}

func loginPost(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	passwordHash := r.FormValue("password")

	var curUser User
	err := db.QueryRow("SELECT ID, Username, Password, Token FROM users WHERE Username=?", username).
		Scan(&curUser.ID, &curUser.Username, &curUser.Password, &curUser.Token)
	if err != nil {
		log.Println("LOGIN: Failed.", username, "doesn't exist")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	byteHash := []byte(passwordHash)
	err = bcrypt.CompareHashAndPassword([]byte(curUser.Password), byteHash)
	if err != nil {
		log.Println("LOGIN: ", curUser.Username, err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	randBytes := make([]byte, 24)
	_, err = rand.Read(randBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal Server Error")
		return
	}
	sessionToken := base64.URLEncoding.EncodeToString(randBytes)

	tokenPrep, err := db.Prepare("UPDATE users SET Token=? WHERE ID=?")
	if err != nil {
		log.Println("LOGIN: Failed preparing Token", sessionToken, "for User", curUser.Username, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal Server Error")
		return
	}
	_, err = tokenPrep.Exec(sessionToken, curUser.ID)
	if err != nil {
		log.Println("LOGIN: Failed storing Token", sessionToken, "for User", curUser.Username, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal Server Error")
		return
	}

	log.Println("LOGIN:", username, sessionToken)

	http.SetCookie(w, &http.Cookie{
		Name:  "session_token",
		Value: sessionToken,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func checkAuth(w http.ResponseWriter, r *http.Request) (User, error) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return User{}, err
		}
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Bad Request")
		return User{}, err
	}
	sessionToken := c.Value

	var u User
	err = db.QueryRow("SELECT ID, Username, Password, Token FROM users WHERE Token=?", sessionToken).
		Scan(&u.ID, &u.Username, &u.Password, &u.Token)
	if err != nil {
		log.Println("TOKEN: Failed", sessionToken)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return User{}, err
	}

	log.Println("TOKEN: AUTHORIZED", sessionToken, "for User", u.Username)
	return u, nil
}
