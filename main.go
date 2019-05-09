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
    "io/ioutil"
    "fmt"
    "log"
    "net/http"
    "html/template"
    "crypto/rand"
    "encoding/base64"
    "encoding/json"
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

var accountsJsonFile = os.Getenv("LEAGUE_ACCS_JSON")

type LoginPage struct {
    Username string
    Password string
}

type AccountsPage []struct {
    Region              string      `json:"region"`
    Tags                []string    `json:"tags"`
    Ign                 string      `json:"ign"`
    Username            string      `json:"username"`
    Password            string      `json:"password"`
    User                string      `json:"user"`
    Leaverbuster        string      `json:"leaverbuster"`
    Ban                 string      `json:"ban"`
    Ban_recently        string      `json:"ban_recently"`
    Owner_active        bool        `json:"owner_active"`
    Password_changed    bool        `json:"password_changed"`
    Pre_30              bool        `json:"pre_3p"`
    Elo                 string      `json:"elo"`
}

func main() {
    accountsJsonFileStat, err := os.Stat(accountsJsonFile)
    log.Println("Using", accountsJsonFile)
    if err != nil {
        log.Fatal("ERROR: Error loading Json file. LEAGUE_ACCS_JSON set correctly?")
    }
    if accountsJsonFileStat.Mode().IsRegular() == false {
        log.Fatal("ERROR: LEAGUE_ACCS_JSON is not a regular file!")
    }

    http.HandleFunc("/login", login)
    http.HandleFunc("/", accounts)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func accounts(w http.ResponseWriter, r *http.Request) {
    current_username, err := checkAuth(w, r)
    if err != nil {
        return
    }

    accountsFile, err := os.Open(accountsJsonFile)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintln(w, "Internal Server Error")
        return
    }
    defer accountsFile.Close()

    accountsContent, err := ioutil.ReadAll(accountsFile)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintln(w, "Internal Server Error")
        return
    }

    var accountsParsed AccountsPage
    json.Unmarshal(accountsContent, &accountsParsed)
    log.Println(current_username)
    log.Println(accountsParsed)

    templates.ExecuteTemplate(w, "accounts.html", nil)
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
