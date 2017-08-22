package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "net/url"
    "strconv"

    "github.com/gorilla/mux"
)

func handleRequests() {
    router := mux.NewRouter().StrictSlash(true)
    router.HandleFunc("/", homeHandler)
    router.HandleFunc("/reports", reportsHandler)
    router.HandleFunc("/tasks", tasksHandler)
    http.ListenAndServe(":8080", router)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "scout %s\r\n", _VERSION)
}

func errorHandler(writer http.ResponseWriter, status int) {
    writer.WriteHeader(status)

    switch status {
    case http.StatusBadRequest:
        fmt.Fprint(writer, "400 bad request\r\n")
    case http.StatusNotFound:
        fmt.Fprint(writer, "404 page not found\r\n")
    default:
    }
}

func queryHandler(writer http.ResponseWriter, request *http.Request, encoder *json.Encoder) (url.Values, error) {
    vals := url.Values{}
    var err error = nil

    if vals, err = url.ParseQuery(request.URL.RawQuery); err == nil {
        for key, _ := range vals {
            if key == "pretty" {
                encoder.SetIndent("", "   ")
                delete(vals, key)
                break
            }
        }
    } else {
        errorHandler(writer, http.StatusNotFound)
    }

    return vals, err
}

func reportsHandler(writer http.ResponseWriter, request *http.Request) {
    //db: = [][]database
    encoder := json.NewEncoder(writer)

    if _, err := queryHandler(writer, request, encoder); err == nil {
        encoder.Encode(*_database)
    }
}

func taskHandler(writer http.ResponseWriter, query url.Values) (map[int]TaskEntry, error) {
    tasks := make(map[int]TaskEntry)
    var err error = nil

    for key, val := range query {
        if key == "id" {
            for _, i := range val {
                if id, e := strconv.Atoi(i); e == nil {
                    if (id < len(*_tasks)) && (id >= 0) {
                        tasks[id] = (*_tasks)[id]
                    } else {
                        err = errors.New("Task '" + i + "' not found")
                        break
                    }
                } else {
                    err = errors.New("'" + i + "' is not a number")
                    break
                }
            }
        } else {
            // @todo Search by task system (e.g., MacOs)
            err = errors.New("Key '" + key + "' not supported")
        }

        if err != nil {
            errorHandler(writer, http.StatusBadRequest)
            break
        }
    }

    return tasks, err
}

func tasksHandler(writer http.ResponseWriter, request *http.Request) {
    encoder := json.NewEncoder(writer)

    if query, err := queryHandler(writer, request, encoder); err == nil {
        if tasks, err := taskHandler(writer, query); err == nil {
            if len(tasks) == 0 {
                for i, v := range *_tasks {
                    tasks[i] = v
                }
            }
            encoder.Encode(tasks)
        }
    }
}
