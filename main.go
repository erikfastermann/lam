package main

import (
    "os"
    "time"
    "io/ioutil"
    "fmt"
    "strconv"
    "log"
    "net/http"
    "html/template"
    "crypto/rand"
    "encoding/base64"
    "encoding/json"
    "github.com/gorilla/mux"
)

type User struct {
    Password    string
    Token       *string
}

var u1Token string = ""
var u2Token string = ""

var loginUsernames = []string { os.Getenv("LEAGUE_ACCS_USER1"), os.Getenv("LEAGUE_ACCS_USER2"), }

var users = map[string]User {
    loginUsernames[0]: {Password: os.Getenv("LEAGUE_ACCS_USER1_PW"), Token: &u1Token},
    loginUsernames[1]: {Password: os.Getenv("LEAGUE_ACCS_USER2_PW"), Token: &u2Token},
}

var templates = template.Must(template.ParseGlob(os.Getenv("LEAGUE_ACCS_TEMPLATE_DIR")))

var accountsJsonFile = os.Getenv("LEAGUE_ACCS_JSON")

type LoginPage struct {
    Username string
    Password string
}

type AccountsPage struct {
    Username    string
    Accounts    AccountsTable
}

type AccountsTable []struct {
    Id                  int         `json:"id"`
    Region              string      `json:"region"`
    Tag                 string    `json:"tag"`
    Ign                 string      `json:"ign"`
    Username            string      `json:"username"`
    Password            string      `json:"password"`
    User                string      `json:"user"`
    Users               []string
    Leaverbuster        int         `json:"leaverbuster"`
    Ban                 string      `json:"ban"`
    Banned              bool
    Password_changed    bool        `json:"password_changed"`
    Pre_30              bool        `json:"pre_30"`
    Link                string
    Elo                 string      `json:"elo"`
}


type EditPage struct {
    Username    string
    Account     EditTable
}

type EditTable struct {
    Id                  int
    Region              string
    Tag                 string
    Ign                 string
    Username            string
    Password            string
    User                string
    Users               []string
    Leaverbuster        int
    Ban                 string
    Banned              bool
    Password_changed    bool
    Pre_30              bool
    Link                string
    Elo                 string
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

    router := mux.NewRouter()
    router.HandleFunc("/login", login)
    router.HandleFunc("/", accounts)
    router.HandleFunc("/edit/{id:[0-9]+}", edit)
    log.Fatal(http.ListenAndServe(":8080", router))
}

func accounts(w http.ResponseWriter, r *http.Request) {
    // currentUsername, err := checkAuth(w, r)
    // if err != nil {
    //     return
    // }
    currentUsername := "erik"

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

    var accountsParsed AccountsTable
    json.Unmarshal(accountsContent, &accountsParsed)

    for i, account := range accountsParsed {
        accountsParsed[i].Link = fmt.Sprintf("https://www.leagueofgraphs.com/de/summoner/%s/%s", account.Region, account.Ign)

        accountsParsed[i].Elo = "Not implemented"

        if account.Ban == "permanent" {
            accountsParsed[i].Banned = true
            continue
        }
        if account.Ban != "" {
            ban, err := time.Parse(time.RFC3339, account.Ban)
            if err != nil {
                log.Println("ERROR: Couldn't parse date:", account.Ban)
                continue
            }
            if ban.Unix() - time.Now().Unix() > 0 {
                accountsParsed[i].Banned = true
            }
        }
    }

    data := AccountsPage{Username: currentUsername, Accounts: accountsParsed}

    templates.ExecuteTemplate(w, "accounts.html", data)
}

func edit(w http.ResponseWriter, r *http.Request) {
    // currentUsername, err := checkAuth(w, r)
    // if err != nil {
    //     return
    // }
    if r.Method != http.MethodGet {
        http.Redirect(w, r, "/", http.StatusSeeOther)
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

    var accountsParsed AccountsTable
    json.Unmarshal(accountsContent, &accountsParsed)

    urlVars := mux.Vars(r)
    id, err := strconv.Atoi(urlVars["id"])
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintln(w, "Internal Server Error")
        return
    }
    if id > len(accountsParsed) - 1 {
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprintln(w, "Bad Request")
        return
    }

    currentAccount := EditTable(accountsParsed[id])
    currentAccount.Users = loginUsernames
    currentUsername := "erik"
    data := EditPage{Username: currentUsername, Account: currentAccount}

    templates.ExecuteTemplate(w, "edit.html", data)
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
    http.Redirect(w, r, "/", http.StatusSeeOther)
}
