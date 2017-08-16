package main

import (
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/gorilla/mux"
)

func all(w http.ResponseWriter, r *http.Request) {
    enc := json.NewEncoder(w)

    if len(r.URL.RawQuery) == 0 {
        enc.Encode(*_database)
    } else if r.URL.RawQuery == "pretty" {
        enc.SetIndent("", "   ")
        enc.Encode(*_database)
    } else {
        w.WriteHeader(http.StatusNotFound)
        fmt.Fprint(w, "404 page not found\r\n")
    }
}

func home(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Welcome to scout!") // Add version info
}

func handleRequests() {
    router := mux.NewRouter().StrictSlash(true)
    router.HandleFunc("/", home)
    router.HandleFunc("/all", all)
    http.ListenAndServe(":8080", router)
}
