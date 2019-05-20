package main

import (
    "fmt"
    "log"
    "time"
    "errors"
    "net/http"
    "github.com/gorilla/mux"
    "net/url"
    "strconv"
    "github.com/go-sql-driver/mysql"
)

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

type EditPage struct {
    Users       []string
    Username    string
    Account     AccountDb
}

func createGet(w http.ResponseWriter, r *http.Request) {
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


func createPost(w http.ResponseWriter, r *http.Request) {
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

func editGet(w http.ResponseWriter, r *http.Request) {
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

func editPost(w http.ResponseWriter, r *http.Request) {
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

