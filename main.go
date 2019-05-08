/* Environment vars:
User names:                     LEAGUE_ACCS_USER1       LEAGUE_ACCS_USER2
User passwords:                 LEAGUE_ACCS_USER1_PW    LEAGUE_ACCS_USER2_PW
CSRF Token (32 Byte):           LEAGUE_ACCS_CSRF
Template Dir (e.g.: /tmp/*):    LEAGUE_ACCS_TEMPLATE_DIR
Setting all of them is mandatory for a usable experience.
*/

package main

import (
    "os"
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
    os.Getenv("LEAGUE_ACCS_USER1"): {Password: os.Getenv("LEAGUE_ACCS_USER1_PW"), Token: &u1Token},
    os.Getenv("LEAGUE_ACCS_USER2"): {Password: os.Getenv("LEAGUE_ACCS_USER2_PW"), Token: &u2Token},
}

var templates = template.Must(template.ParseGlob(os.Getenv("LEAGUE_ACCS_TEMPLATE_DIR")))

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
    templates.ExecuteTemplate(w, "accounts.html", data)
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
    if r.Method != http.MethodPost {
        templates.ExecuteTemplate(w, "login.html", nil)
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
    _, err := rand.Read(randBytes)
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
