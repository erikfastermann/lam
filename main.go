package main

import (
    "fmt"
    "log"
    "net/http"
    "html/template"
    "crypto/rand"
    "encoding/base64"
)

type User struct {
    Password string
    Token *string
}

var u1Token string = ""
var u2Token string = ""

var users = map[string]User {
    "felix": {Password: "BALENCIAGAJA69", Token: &u1Token},
    "erik": {Password: "VVSJA88", Token: &u2Token},
}

type LoginPage struct {
    Username string
    Password string
}

type AccountsPage struct {
    Username string
}

func main() {
    http.HandleFunc("/login", login)
    http.HandleFunc("/", accounts)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func accounts(w http.ResponseWriter, r *http.Request) {
    username, err := checkAuth(w, r)
    if err != nil {
        return
    }
    data := AccountsPage{Username: username}
    template, err := template.ParseFiles("template/accounts.html")
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintln(w, "Internal Server Error")
        return
    }
    template.Execute(w, data)
}

func checkAuth(w http.ResponseWriter, r *http.Request) (string, error) {
    c, err := r.Cookie("session_token")
    if err != nil {
        if err == http.ErrNoCookie {
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return "", err
        }
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprintln(w, "Bad Request")
        return "", err
    }

    sessionToken := c.Value
    log.Println("TESTING-TOKEN:", sessionToken)
    for username, user := range users {
        if *user.Token == sessionToken {
            log.Println("AUTH:", username)
            return username, nil
        }
    }

    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return "", err
}

func login(w http.ResponseWriter, r *http.Request) {
    template, err := template.ParseFiles("template/login.html")
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintln(w, "Internal Server Error")
        return
    }

    if r.Method != http.MethodPost {
        template.Execute(w, nil)
        return
    }

    creds := LoginPage {
        Username: r.FormValue("username"),
        Password: r.FormValue("password"),
    }

    expectedPassword := users[creds.Username].Password
    if expectedPassword != creds.Password {
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

    *users[creds.Username].Token = sessionToken
    log.Println("LOGIN:", creds.Username, *users[creds.Username].Token)

    http.SetCookie(w, &http.Cookie{
        Name:    "session_token",
        Value:   sessionToken,
    })
    http.Redirect(w, r, "/accs", http.StatusSeeOther)
}
