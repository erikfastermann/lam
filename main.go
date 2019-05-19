package main

import (
    "os"
    "time"
    "fmt"
    "errors"
    "strconv"
    "log"
    "sort"
    "net/http"
    "net/url"
    "html/template"
    "crypto/rand"
    "golang.org/x/crypto/bcrypt"
    "encoding/base64"
    "github.com/gorilla/mux"
    "database/sql"
    "github.com/go-sql-driver/mysql"
    "github.com/PuerkitoBio/goquery"
)

var templates *template.Template

type User struct {
    ID          int
    Username    string
    Password    string
    Token       string
}

type AccountDb struct {
    ID                  int
    Region              string
    Tag                 string
    Ign                 string
    Username            string
    Password            string
    User                string
    Leaverbuster        int
    Ban                 mysql.NullTime
    Perma               bool
    PasswordChanged     bool
    Pre30               bool
    Elo                 string
}

type AccountData struct {
    Color   string
    Banned  bool
    Link    string
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
    envVar := func(env string) (string) {
        str := os.Getenv(env)
        if str == "" {
            log.Fatalln(env, "is empty or doesn't exist")
        }
        return str
    }

    templateDir := envVar("LEAGUE_ACCS_TEMPLATE_DIR")
    templates = template.Must(template.ParseGlob(templateDir))

    dbUser := envVar("LEAGUE_ACCS_DB_USER")
    dbPassword := envVar("LEAGUE_ACCS_DB_PASSWORD")
    dbAddress := envVar("LEAGUE_ACCS_DB_ADDRESS")
    var err error
    db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/lol_accs", dbUser, dbPassword, dbAddress))
    if err != nil {
        log.Fatal(err)
    }

    err = db.Ping()
    if err != nil {
        log.Fatal(err)
    }

    defer db.Close()

    go webParser()
    router := mux.NewRouter()
    router.HandleFunc("/login", loginGet).Methods("GET")
    router.HandleFunc("/login", loginPost).Methods("POST")
    router.HandleFunc("/", accounts).Methods("GET")
    router.HandleFunc("/edit/{id:[0-9]+}", editAccGet).Methods("GET")
    router.HandleFunc("/edit/{id:[0-9]+}", editAccPost).Methods("POST")
    router.HandleFunc("/new", createAccGet).Methods("GET")
    router.HandleFunc("/new", createAccPost).Methods("POST")
    log.Fatal(http.ListenAndServe(":8080", router))
}

func webParser() {
    client := &http.Client{
        Timeout: 10 * time.Second,
    }

    for range time.NewTicker(time.Hour).C {
        accs, err := allAccounts()
        if err != nil {
            log.Println("WEB-PARSER: ERROR reading from database.", err)
            return
        }
        for _, acc := range accs {
            url, err := url.Parse(fmt.Sprintf("https://www.leagueofgraphs.com/en/summoner/%s/%s", acc.Region, acc.Ign))
            if err != nil {
                log.Println("WEB-PARSER: ERROR escaping", url, err)
                continue
            }
            link := url.String()

            res, err := client.Get(link)
            if err != nil {
                log.Println("WEB-PARSER: ERROR opening", link, err)
                continue
            }
            defer res.Body.Close()

            doc, err := goquery.NewDocumentFromReader(res.Body)
            if err != nil {
                log.Println("WEB-PARSER: ERROR parsing", link, err)
                continue
            }
            leagueTier := doc.Find(".leagueTier").Text()
            if leagueTier == "" {
                log.Println("WEB-PARSER: ERROR finding .leagueTier", link)
                continue
            }

            tokenPrep, err := db.Prepare("UPDATE accounts SET Elo=? WHERE ID=?")
            if err != nil {
                log.Println("WEB-PARSER: FAILED preparing Elo", leagueTier, "for Account", acc.Ign, err)
                continue
            }
            _, err = tokenPrep.Exec(leagueTier, acc.ID)
            if err != nil {
                log.Println("WEB-PARSER: FAILED storing Elo", leagueTier, "for Account", acc.Ign, err)
                continue
            }
            log.Println("WEB-PARSER: SUCCESS storing Elo:", leagueTier, "for Account", acc.Ign)
        }
    }
}

func allAccounts() ([]*AccountDb, error) {
    rows, err := db.Query(`SELECT ID, Region, Tag, Ign, Username, Password, User,
        Leaverbuster, Ban, Perma, PasswordChanged, Pre30, Elo FROM accounts`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    accs := make([]*AccountDb, 0)
    for rows.Next() {
        acc := new(AccountDb)
        err := rows.Scan(&acc.ID, &acc.Region, &acc.Tag, &acc.Ign,
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

    accsParsed, err := allAccounts()
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintln(w, "Internal Server Error")
        return
    }

    var accsComputed []AccountData
    var link string

    for _, acc := range accsParsed {
        banned := false
        if acc.Region == "" || acc.Ign == "" {
            link = ""
        } else {
            link = fmt.Sprintf("https://www.leagueofgraphs.com/de/summoner/%s/%s", acc.Region, acc.Ign)
        }

        if acc.Perma {
            banned = true
        } else if acc.Ban.Valid {
            if acc.Ban.Time.Unix() - time.Now().Unix() > 0 {
                banned = true
            }
        } else {
            banned = false
        }

        accsComputed = append(accsComputed, AccountData{Banned: banned, Link: link, Account: *acc})
    }

    sort.SliceStable(accsComputed, func(i, j int) bool {
        return accsComputed[i].Account.Tag < accsComputed[j].Account.Tag
    })

    var accsFinal []AccountData
    for i := 0; i < 3; i++ {
        for _, acc := range accsComputed {
            switch i {
            case 0:
                if !acc.Banned && !acc.Account.PasswordChanged {
                    accsFinal = append(accsFinal, acc)
                }
            case 1:
                if acc.Banned && !acc.Account.Perma {
                    acc.Color = "table-warning"
                    accsFinal = append(accsFinal, acc)
                }
            case 2:
                if acc.Account.Perma || acc.Account.PasswordChanged {
                    acc.Color = "table-danger"
                    accsFinal = append(accsFinal, acc)
                }
            }
        }
    }

    data := AccountsPage{Username: curUser.Username, Accounts: accsFinal}
    templates.ExecuteTemplate(w, "accounts.html", data)
}

func createAccPost(w http.ResponseWriter, r *http.Request) {
    if _, err := checkAuth(w, r); err != nil {
       return
    }

    if err := r.ParseForm(); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprintln(w, "Bad Request")
        return
    }

    acc, err := parseFormAcc(r.PostForm)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprintln(w, "Bad Request")
        return
    }

    accPrep, err := db.Prepare(`INSERT INTO accounts(Region, Tag, Ign, Username, Password,
        User, Leaverbuster, Ban, Perma, PasswordChanged, Pre30) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
    if err != nil {
        log.Println("NEW: FAILED preparing db for new account.", err)
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintln(w, "Internal Server Error")
        return
    }
    _, err = accPrep.Exec(acc.Region, acc.Tag, acc.Ign, acc.Username, acc.Password,
        acc.User, acc.Leaverbuster, acc.Ban, acc.Perma, acc.PasswordChanged, acc.Pre30)
    if err != nil {
        log.Println("NEW: FAILED writing to db for new account.", err)
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintln(w, "Internal Server Error")
        return
    }

    log.Println("NEW: SUCCESS creating account")
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func createAccGet(w http.ResponseWriter, r *http.Request) {
    curUser, err := checkAuth(w, r)
    if err != nil {
       return
    }

    loginUsers, err := queryUsernames()
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintln(w, "Internal Server Error")
        return
    }

    acc := AccountDb{Region: "euw", User: curUser.Username}
    data := EditPage{Users: loginUsers, Username: curUser.Username, Account: acc}
    templates.ExecuteTemplate(w, "edit.html", data)
}

func editAccGet(w http.ResponseWriter, r *http.Request) {
    curUser, err := checkAuth(w, r)
    if err != nil {
       return
    }

    urlVars := mux.Vars(r)
    id, err := strconv.Atoi(urlVars["id"])
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintln(w, "Internal Server Error")
        return
    }

    var acc AccountDb
    err = db.QueryRow(`SELECT ID, Region, Tag, Ign, Username, Password, User,
        Leaverbuster, Ban, Perma, PasswordChanged, Pre30, Elo FROM accounts WHERE ID=?`, id).
        Scan(&acc.ID, &acc.Region, &acc.Tag, &acc.Ign,
        &acc.Username, &acc.Password, &acc.User, &acc.Leaverbuster,
        &acc.Ban, &acc.Perma, &acc.PasswordChanged, &acc.Pre30, &acc.Elo)
    if err != nil {
        fmt.Println("EDIT: ERROR reading from database", err)
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintln(w, "Bad Request")
        return
    }

    loginUsers, err := queryUsernames()
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintln(w, "Internal Server Error")
        return
    }

    data := EditPage{Users: loginUsers, Username: curUser.Username, Account: acc}
    templates.ExecuteTemplate(w, "edit.html", data)
}

func editAccPost(w http.ResponseWriter, r *http.Request) {
    if _, err := checkAuth(w, r); err != nil {
        return
    }

    urlVars := mux.Vars(r)
    id, err := strconv.Atoi(urlVars["id"])
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintln(w, "Internal Server Error")
        return
    }

    if err := r.ParseForm(); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprintln(w, "Bad Request")
        return
    }

    acc, err := parseFormAcc(r.PostForm)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintln(w, "Bad Request")
        return
    }

    accPrep, err := db.Prepare(`UPDATE accounts SET Region=?, Tag=?, Ign=?, Username=?, Password=?,
        User=?, Leaverbuster=?, Ban=?, Perma=?, PasswordChanged=?, Pre30=? WHERE ID=?`)
    if err != nil {
        log.Println("EDIT: FAILED preparing db for account-id", id, err)
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintln(w, "Internal Server Error")
        return
    }
    _, err = accPrep.Exec(acc.Region, acc.Tag, acc.Ign, acc.Username, acc.Password,
        acc.User, acc.Leaverbuster, acc.Ban, acc.Perma, acc.PasswordChanged, acc.Pre30, id)
    if err != nil {
        log.Println("EDIT: FAILED writing to db for account-id", id, err)
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintln(w, "Internal Server Error")
        return
    }

    log.Println("EDIT: SUCCESS editing account", id, acc.Ign)
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func parseFormAcc(form url.Values) (*AccountDb, error) {
    acc := new(AccountDb)

    formVal := func(key string) string {
        val, ok := form[key]
        if !ok {
            return ""
        }
        return val[0]
    }

    acc.Region = formVal("region")
    acc.Tag = formVal("tag")
    acc.Ign = formVal("ign")
    acc.Username = formVal("username")
    acc.Password = formVal("password")
    acc.User = formVal("user")

    leaverbuster := formVal("leaverbuster")
    leaverbusterInt, err := strconv.Atoi(leaverbuster)
    if err != nil {
        return nil, err
    }
    acc.Leaverbuster = leaverbusterInt

    banForm := formVal("ban")
    var ban mysql.NullTime
    if banForm == "" {
        ban = mysql.NullTime{Valid: false}
    } else {
        banTime, err := time.Parse("2006-01-02 15:04", banForm)
        if err != nil {
            return nil, err
        }
        ban = mysql.NullTime{Time: banTime, Valid: true}
    }
    acc.Ban = ban

    parseBool := func(str string) (bool, error) {
        if str == "true" {
            return true, nil
        }
        if str == "false" || str == "" {
            return false, nil
        }
        return false, errors.New("BAD REQUEST")
    }

    perma := formVal("perma")
    acc.Perma, err = parseBool(perma)
    if err != nil {
        return nil, err
    }

    pwChanged := formVal("password_changed")
    acc.PasswordChanged, err = parseBool(pwChanged)
    if err != nil {
        return nil, err
    }

    pre30 := formVal("pre_30")
    acc.Pre30, err = parseBool(pre30)
    if err != nil {
        return nil, err
    }
    return acc, nil
}

func queryUsernames() ([]string, error) {
    loginUsers := make([]string, 0)

    rows, err := db.Query("SELECT Username FROM users")
    if err != nil {
        return loginUsers, err
    }
    defer rows.Close()

    for rows.Next() {
        var u string
        err := rows.Scan(&u)
        if err != nil {
            return loginUsers, err
        }
        loginUsers = append(loginUsers, u)
    }
    if err = rows.Err(); err != nil {
        return loginUsers, err
    }

    return loginUsers, nil
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
        Name:    "session_token",
        Value:   sessionToken,
    })
    http.Redirect(w, r, "/", http.StatusSeeOther)
}
