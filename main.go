package main

import (
    "os"
    "fmt"
    "log"
    "net/http"
    "html/template"
    "github.com/gorilla/mux"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

var templates *template.Template

var db *sql.DB

func main() {
    envVar := func(env string) (string) {
        str := os.Getenv(env)
        if str == "" {
            log.Fatalln("Environment variable", env, "is empty or doesn't exist")
        }
        return str
    }

    templateDir := envVar("LEAGUE_ACCS_TEMPLATE_DIR")
    templates = template.Must(template.ParseGlob(templateDir))

    dbUser := envVar("LEAGUE_ACCS_DB_USER")
    dbPassword := envVar("LEAGUE_ACCS_DB_PASSWORD")
    dbAddress := envVar("LEAGUE_ACCS_DB_ADDRESS")
    dbName := envVar("LEAGUE_ACCS_DB_NAME")
    var err error
    db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", dbUser, dbPassword, dbAddress, dbName))
    if err != nil {
        log.Fatal(err)
    }
    err = db.Ping()
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    go elo()
    router := mux.NewRouter()
    router.HandleFunc("/login", loginGet).Methods("GET")
    router.HandleFunc("/login", loginPost).Methods("POST")
    router.HandleFunc("/", accounts).Methods("GET")
    router.HandleFunc("/edit/{id:[0-9]+}", editGet).Methods("GET")
    router.HandleFunc("/edit/{id:[0-9]+}", editPost).Methods("POST")
    router.HandleFunc("/new", createGet).Methods("GET")
    router.HandleFunc("/new", createPost).Methods("POST")
    log.Println("Starting server on port 8080")
    log.Fatal(http.ListenAndServe(":8080", router))
}

