package control

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/shanebarnes/goto/logger"
	"github.com/shanebarnes/scout/execution"
	"github.com/shanebarnes/scout/global"
)

func TaskHandler(writer http.ResponseWriter, query url.Values) (*[]execution.TaskEntry, error) {
	var err error = nil
	var tasks []execution.TaskEntry
	db := global.GetDb()
	fields := db.GetFields(&execution.TaskEntry{})
	sql := "SELECT " + fields + " FROM TaskEntry"
	sqlClause := " WHERE "

	for key, val := range query {
		switch strings.ToLower(key) {
		case "active":
			sql = sql + sqlClause + "active IN (" + strings.Join(val, ",") + ")"
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

	if err == nil {
		err = db.Select(sql, &tasks)
	} else {
		errorHandler(writer, http.StatusBadRequest)
	}

	return &tasks, err
}

func TasksHandler(writer http.ResponseWriter, request *http.Request) {
	defer logger.PrintlnTime(logger.Info, time.Time{}, "Prepared response to "+request.Method+" "+request.URL.Path)()
	encoder := json.NewEncoder(writer)

	if query, err := queryHandler(writer, request, encoder); err == nil {
		if tasks, err := TaskHandler(writer, query); err == nil {
			encoder.Encode(tasks)
		}
	}
}
