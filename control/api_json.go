package control

import (
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "net/url"
    //"os"
    //"strconv"
    "strings"

    "github.com/gorilla/mux"
    "github.com/shanebarnes/goto/logger"
    "github.com/shanebarnes/scout/execution"
    "github.com/shanebarnes/scout/global"
    "github.com/shanebarnes/scout/mission"
)

//var REPORTS *[][]Database = nil
//var TASKS *execution.TaskArray = nil
// Call this the reporter?

func HandleRequests(ctl *Control) {
    router := mux.NewRouter().StrictSlash(true)
    router.HandleFunc("/", homeHandler)
    router.HandleFunc("/dashboard", dashboardHandler)
    router.PathPrefix("/freeboard").Handler(http.StripPrefix("/freeboard", http.FileServer(http.Dir(ctl.Root))))
    router.HandleFunc("/reports", reportsHandler)
    router.HandleFunc("/tasks", tasksHandler)

    loadDashboard(ctl)
    http.ListenAndServe(":8080", router)
}

func loadDashboard(ctl *Control) {
    //dashboard := NewDashboard(REPORTS)
    //if b, err := json.MarshalIndent(dashboard, "", "    "); err == nil {
    //    if file, err := os.OpenFile(ctl.Root + "/" + ScoutFreeboard, os.O_CREATE | os.O_RDWR, 0644); err == nil {
    //        if _, err := file.Write(b); err != nil {
    //            logger.PrintlnError(err.Error())
    //        }
    //        file.Close()
    //    } else {
    //        logger.PrintlnError(err.Error())
    //    }
    //} else {
    //    logger.PrintlnError(err.Error())
    //}
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/freeboard/#source=" + ScoutFreeboard, http.StatusSeeOther)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "scout %s\r\n", mission.GetVersion())
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

func reportHandler(writer http.ResponseWriter, query url.Values) (map[int][]Database, error) {
    reports := make(map[int][]Database)
    var err error = nil

    // @todo add subid for particular sub-reports
    // @todo add autorefresh using JS or websockets
    //for key, val := range query {
    //    if key == "id" {
    //        for _, i := range val {
    //            if id, e := strconv.Atoi(i); e == nil {
    //                if (id < len(*REPORTS)) && (id >= 0) {
    //                    reports[id] = (*REPORTS)[id]
    //                } else {
    //                    err = errors.New("Report '" + i + "' not found")
    //                    break
    //                }
    //            } else {
    //                err = errors.New("'" + i + "' is not a number")
    //                break
    //            }
    //        }
    //    } else {
    //        err = errors.New("Key '" + key + "' not supported")
    //    }

    //    if err != nil {
    //        errorHandler(writer, http.StatusBadRequest)
    //        break
    //    }
    //}

    return reports, err
}

func reportsHandler(writer http.ResponseWriter, request *http.Request) {
    encoder := json.NewEncoder(writer)

    if query, err := queryHandler(writer, request, encoder); err == nil {
        if reports, err := reportHandler(writer, query); err == nil {
            if len(reports) == 0 {
                //for i, v := range *REPORTS {
                //    reports[i] = v
                //}
            }
            encoder.Encode(reports)
        }
    }
}

func taskHandler(writer http.ResponseWriter, query url.Values) (*[]execution.TaskEntry, error) {
    var err error = nil
    var tasks []execution.TaskEntry
    db := global.GetDb()
    fields := db.GetFields(&execution.TaskEntry{})
    sql := "SELECT " + fields + " FROM TaskEntry"
    sqlClause := " WHERE "

    for key, val := range query {
        switch strings.ToLower(key) {
        case "active":
            sql  = sql + sqlClause + "active IN (" + strings.Join(val, ",") + ")"
        case "cmd":
            sql = sql + sqlClause + "command LIKE '%" + strings.Join(val, "%' AND command LIKE '%") + "%'"
        case "desc":
            sql = sql + sqlClause + "description LIKE '%" + strings.Join(val, "%' AND description LIKE '%") + "%'"
        case "groupid":
            sql = sql + sqlClause + "group_id IN (" + strings.Join(val, ",") + ")"
        case "id":
            sql = sql + sqlClause + "id IN (" + strings.Join(val, ",") + ")"
        case "name":
            sql = sql + sqlClause + "name LIKE '%" + strings.Join(val, "%' AND name LIKE '%") + "%'"
        default:
            err = errors.New("Key '" + key + "' not supported")
            break
        }

        sqlClause = " AND "
    }

logger.PrintlnError("query:", sql)
    if err == nil {
        err = db.Select(sql, &tasks)
    } else {
        errorHandler(writer, http.StatusBadRequest)
    }

    return &tasks, err
}

func tasksHandler(writer http.ResponseWriter, request *http.Request) {
    encoder := json.NewEncoder(writer)

    if query, err := queryHandler(writer, request, encoder); err == nil {
        if tasks, err := taskHandler(writer, query); err == nil {
            encoder.Encode(tasks)
        }
    }
}
