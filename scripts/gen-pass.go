package main

import (
    "os"
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    if len(os.Args) != 2 {
        fmt.Fprintln(os.Stderr, "USAGE:", os.Args[0], "PASSWORD")
        return
    }
    password := []byte(os.Args[1])
    hash, err := bcrypt.GenerateFromPassword(password, bcrypt.MinCost)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        return
    }
    fmt.Println(string(hash))
}

