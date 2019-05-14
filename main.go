package main

import (
    "os"
    "time"
    "fmt"
    "log"
    "net/http"
    "html/template"
    "crypto/rand"
    "golang.org/x/crypto/bcrypt"
    "encoding/base64"
    "github.com/gorilla/mux"
    "database/sql"
    "github.com/go-sql-driver/mysql"
)

var templates *template.Template

type User struct {
    ID          int
    Username    string
    Password    string
    Token       string
}

type AccountDb struct {
    Id                  int             `json:"id"`
    Region              string          `json:"region"`
    Tag                 string          `json:"tag"`
    Ign                 string          `json:"ign"`
    Username            string          `json:"username"`
    Password            string          `json:"password"`
    User                string          `json:"user"`
    Leaverbuster        int             `json:"leaverbuster"`
    Ban                 mysql.NullTime  `json:"ban"`
    Perma               bool            `json:"ban"`
    PasswordChanged     bool            `json:"password_changed"`
    Pre30               bool            `json:"pre_30"`
    Elo                 string          `json:"pre_30"`
}

type AccountData struct {
    Banned  bool
    Link    string
    Elo     string
    Account AccountDb
}

type AccountsPage struct {
    Username    string
    Accounts    []AccountData
}

type EditPage struct {
    Users       []string
    Username    string
    Account     AccountDb
}

var db *sql.DB

func main() {
    templates = template.Must(template.ParseGlob(os.Getenv("LEAGUE_ACCS_TEMPLATE_DIR")))

    var err error
    db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/lol_accs",
        os.Getenv("LEAGUE_ACCS_DB_USER"),
        os.Getenv("LEAGUE_ACCS_DB_PASSWORD"),
        os.Getenv("LEAGUE_ACCS_DB_ADDRESS")))
    if err != nil {
        log.Fatal(err)
    }

    err = db.Ping()
    if err != nil {
        log.Fatal(err)
    }

    defer db.Close()

    router := mux.NewRouter()
    router.HandleFunc("/login", login)
    router.HandleFunc("/", accounts)
    // router.HandleFunc("/edit/{id:[0-9]+}", edit)
    log.Fatal(http.ListenAndServe(":8080", router))
}

func allAccounts() ([]*AccountDb, error) {
    rows, err := db.Query("SELECT * FROM accounts")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    accs := make([]*AccountDb, 0)
    for rows.Next() {
        acc := new(AccountDb)
        err := rows.Scan(&acc.Id, &acc.Region, &acc.Tag, &acc.Ign,
            &acc.Username, &acc.Password, &acc.User, &acc.Leaverbuster,
            &acc.Ban, &acc.Perma, &acc.PasswordChanged, &acc.Pre30, &acc.Elo)
        if err != nil {
            return nil, err
        }
        accs = append(accs, acc)
    }

    if err = rows.Err(); err != nil {
        return nil, err
    }
    return accs, nil
}

func accounts(w http.ResponseWriter, r *http.Request) {
    curUser, err := checkAuth(w, r)
    if err != nil {
        return
    }

    accountsParsed, err := allAccounts()
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintln(w, "Internal Server Error")
        return
    }

    var accountsComputed []AccountData
    var link, elo string
    banned := false

    for _, account := range accountsParsed {
        link = fmt.Sprintf("https://www.leagueofgraphs.com/de/summoner/%s/%s", account.Region, account.Ign)

        elo = "Not implemented"

        if account.Perma {
            banned = true
        } else if account.Ban.Valid {
            if account.Ban.Time.Unix() - time.Now().Unix() > 0 {
                banned = true
            }
        } else {
            banned = false
        }

        accountsComputed = append(accountsComputed, AccountData{Banned: banned, Link: link, Elo: elo, Account: *account})
    }

    data := AccountsPage{Username: curUser.Username, Accounts: accountsComputed}

    templates.ExecuteTemplate(w, "accounts.html", data)
}

// func edit(w http.ResponseWriter, r *http.Request) {
//     return
//     currentUsername, err := checkAuth(w, r)
//     // _, err := checkAuth(w, r)
//     if err != nil {
//         return
//     }

//     accountsParsed, err := parseAccountsJsonFile()
//     if err != nil {
//         w.WriteHeader(http.StatusInternalServerError)
//         fmt.Fprintln(w, "Internal Server Error")
//         return
//     }

//     urlVars := mux.Vars(r)
//     id, err := strconv.Atoi(urlVars["id"])
//     if err != nil {
//         w.WriteHeader(http.StatusInternalServerError)
//         fmt.Fprintln(w, "Internal Server Error")
//         return
//     }
//     if id > len(accountsParsed) - 1 {
//         w.WriteHeader(http.StatusBadRequest)
//         fmt.Fprintln(w, "Bad Request")
//         return
//     }

//     currentAccount := AccountJson(accountsParsed[id])

//     if r.Method == http.MethodPost {
//         currentAccount.Region = r.FormValue("region")
//         currentAccount.Tag = r.FormValue("tag")
//         currentAccount.Ign = r.FormValue("ign")
//         currentAccount.Username = r.FormValue("username")
//         currentAccount.Password = r.FormValue("password")
//         currentAccount.User = r.FormValue("user")

//         leaverbuster, err := strconv.Atoi(r.FormValue("leaverbuster"))
//         if err != nil {
//             w.WriteHeader(http.StatusBadRequest)
//             fmt.Fprintln(w, "Bad Request")
//             return
//         }
//         currentAccount.Leaverbuster = leaverbuster

//         currentAccount.Ban = r.FormValue("ban")

//         passwordChangedForm := r.FormValue("password_changed")
//         var passwordChanged bool
//         if passwordChangedForm == "true" {
//             passwordChanged = true
//         } else if passwordChangedForm == "false" || passwordChangedForm == "" {
//             passwordChanged = false
//         } else {
//             w.WriteHeader(http.StatusBadRequest)
//             fmt.Fprintln(w, "Bad Request")
//             return
//         }
//         currentAccount.Password_changed = passwordChanged

//         pre30Form := r.FormValue("pre_30")
//         var pre30 bool
//         if pre30Form == "true" {
//             pre30 = true
//         } else if pre30Form == "false" || pre30Form == "" {
//             pre30 = false
//         } else {
//             w.WriteHeader(http.StatusBadRequest)
//             fmt.Fprintln(w, "Bad Request")
//             return
//         }
//         currentAccount.Pre_30 = pre30

//         http.Redirect(w, r, "/", http.StatusSeeOther)
//         return
//     }

//     data := EditPage{Users: loginUsernames, Username: currentUsername, Account: currentAccount}

//     templates.ExecuteTemplate(w, "edit.html", data)
// }

func checkAuth(w http.ResponseWriter, r *http.Request) (User, error) {
    var curUser User

    c, err := r.Cookie("session_token")
    if err != nil {
        if err == http.ErrNoCookie {
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return curUser, err
        }
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprintln(w, "Bad Request")
        return curUser, err
    }
    sessionToken := c.Value

    err = db.QueryRow("SELECT * FROM users WHERE Token=?", sessionToken).
        Scan(&curUser.ID, &curUser.Username, &curUser.Password, &curUser.Token)
    if err != nil {
        log.Println("TOKEN: Failed", sessionToken)
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return curUser, err
    }

    log.Println("TOKEN: AUTHORIZED", sessionToken, "for User", curUser.Username)
    return curUser, nil
}

func login(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        templates.ExecuteTemplate(w, "login.html", nil)
        return
    }

    username := r.FormValue("username")
    passwordHash := r.FormValue("password")

    var curUser User
    err := db.QueryRow("SELECT * FROM users WHERE Username=?", username).
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
        Name:    "session_token",
        Value:   sessionToken,
    })
    http.Redirect(w, r, "/", http.StatusSeeOther)
}
