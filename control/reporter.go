package control

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/shanebarnes/goto/logger"
	"github.com/shanebarnes/scout/execution"
	"github.com/shanebarnes/scout/global"
)

type ScoutReport struct {
	ReportId int64                  `db:"report_id"   json:"report_id"`
	TargetId int64                  `db:"target_id"   json:"target_id"`
	Address  string                 `db:"address"     json:"target_location"`
	Name     string                 `db:"name"        json:"target_name"`
	TaskId   int64                  `db:"task_id"     json:"task_id"`
	Task     string                 `db:"description" json:"task_name"`
	Protocol string                 `db:"protocol"    json:"task_protocol"`
	Xdiff    float64                `db:"x_diff"      json:"x_diff"`
	Xval     float64                `db:"x_val"       json:"x_val"`
	Ydiff    float64                `db:"y_diff"      json:"y_diff"`
	Ymax     float64                `db:"y_max"       json:"y_max"`
	Ymin     float64                `db:"y_min"       json:"y_min"`
	Yrate    float64                `db:"y_rate"      json:"y_rate"`
	Yval     float64                `db:"y_val"       json:"y_val"`
	Reports  []execution.TaskReport `db:"-"           json:"reports"`
}

func HandleRequests(ctl *Control) {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", HomeHandler)
	router.HandleFunc("/dashboard", dashboardHandler)
	router.PathPrefix("/freeboard").Handler(http.StripPrefix("/freeboard", http.FileServer(http.Dir(ctl.Root))))
	router.HandleFunc("/group_reports", GroupReportsHandler)
	router.HandleFunc("/reports", reportsHandler)
	router.HandleFunc("/statistics", StatisticsHandler)
	router.HandleFunc("/tasks", TasksHandler)

	// @bug Fix the startup problem here
	time.Sleep(3000 * time.Millisecond)

	loadDashboard(ctl)
	RunAggregator()
	http.ListenAndServe(":8080", router)
}

func loadDashboard(ctl *Control) {
	reports := getReports()
	dashboard := NewDashboard(*reports)
	if b, err := json.MarshalIndent(dashboard, "", "    "); err == nil {
		os.Truncate(ctl.Root+"/"+ScoutFreeboard, 0)

		if file, err := os.OpenFile(ctl.Root+"/"+ScoutFreeboard, os.O_CREATE|os.O_RDWR, 0644); err == nil {
			if _, err := file.Write(b); err != nil {
				logger.PrintlnError(err.Error())
			}

			file.Sync()
			file.Close()
		} else {
			logger.PrintlnError(err.Error())
		}
	} else {
		logger.PrintlnError(err.Error())
	}
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/freeboard/#source="+ScoutFreeboard, http.StatusSeeOther)
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

func getReports() *[]ScoutReport {
	var scoutReports []ScoutReport
	db := global.GetDb()

	sql := "SELECT tr.report_id, tr.target_id, td.address, tg.name, tr.task_id, te.description, tg.protocol, tr.x_diff, tr.x_val, tr.y_diff, tr.y_max, tr.y_min, tr.y_rate, tr.y_val FROM TargetReport tr"
	sql = sql + " LEFT JOIN TargetDef td ON td.id = tr.target_id"
	sql = sql + " LEFT JOIN TargetGroup tg ON tg.id = td.group_id"
	sql = sql + " LEFT JOIN TaskEntry te ON te.id = tr.task_id"
	sqlGroup := " GROUP BY tr.target_id, tr.task_id"
	sqlOrder := " ORDER BY tr.target_id, tr.task_id, tr.x_val DESC"

	sql = sql + sqlGroup + sqlOrder
	db.Select(sql, &scoutReports)

	var taskReports []execution.TaskReport
	db.Select("SELECT * FROM TaskReport", &taskReports)

	for i := range scoutReports {
		for j := range taskReports {
			if taskReports[j].TaskId == scoutReports[i].TaskId {
				scoutReports[i].Reports = append(scoutReports[i].Reports, taskReports[j])
			}
		}
	}

	return &scoutReports
}

func reportHandler(writer http.ResponseWriter, query url.Values) (*[]ScoutReport, error) {
	var err error = nil
	var scoutReports []ScoutReport
	var taskReports []execution.TaskReport

	db := global.GetDb()
	err = db.Select("SELECT * FROM TaskReport", &taskReports)

	sql := "SELECT tr.report_id, tr.target_id, td.address, tg.name, tr.task_id, te.description, tg.protocol, tr.x_diff, tr.x_val, tr.y_diff, tr.y_max, tr.y_min, tr.y_rate, tr.y_val FROM TargetReport tr"
	sql = sql + " LEFT JOIN TargetDef td ON td.id = tr.target_id"
	sql = sql + " LEFT JOIN TargetGroup tg ON tg.id = td.group_id"
	sql = sql + " LEFT JOIN TaskEntry te ON te.id = tr.task_id"
	sqlGroup := " GROUP BY tr.target_id, tr.task_id"
	sqlOrder := " ORDER BY tr.target_id, tr.task_id, tr.x_val DESC"
	sqlClause := " WHERE "

	for key, val := range query {
		switch strings.ToLower(key) {
		case "task_protocol":
			sql = sql + sqlClause + "tg.protocol LIKE '%" + strings.Join(val, "%' AND tg.protocol LIKE '%") + "%'"
		case "target_id":
			sql = sql + sqlClause + "tr.target_id IN (" + strings.Join(val, ",") + ")"
		case "task_id":
			sql = sql + sqlClause + "tr.task_id IN (" + strings.Join(val, ",") + ")"
		default:
			err = errors.New("Key '" + key + "' not supported")
			break
		}

		sqlClause = " AND "
	}

	if err == nil {
		sql = sql + sqlGroup + sqlOrder
		err = db.Select(sql, &scoutReports)
	} else {
		errorHandler(writer, http.StatusBadRequest)
	}

	for i := range scoutReports {
		for j := range taskReports {
			if taskReports[j].TaskId == scoutReports[i].TaskId {
				scoutReports[i].Reports = append(scoutReports[i].Reports, taskReports[j])
			}
		}
	}

	return &scoutReports, err
}

func reportsHandler(writer http.ResponseWriter, request *http.Request) {
	encoder := json.NewEncoder(writer)

	if query, err := queryHandler(writer, request, encoder); err == nil {
		if reports, err := reportHandler(writer, query); err == nil {
			encoder.Encode(reports)
		}
	}
}
