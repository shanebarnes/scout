package control

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/shanebarnes/goto/logger"
	"github.com/shanebarnes/scout/global"
)

func GroupReportsHandler(writer http.ResponseWriter, request *http.Request) {
	defer logger.PrintlnTime(logger.Info, time.Time{}, "Prepared response to "+request.Method+" "+request.URL.Path)()
	encoder := json.NewEncoder(writer)

	if query, err := queryHandler(writer, request, encoder); err == nil {
		if reports, err := GroupReportHandler(writer, query); err == nil {
			encoder.Encode(reports)
		}
	}
}

func GroupReportHandler(writer http.ResponseWriter, query url.Values) (*[]AggregateReport, error) {
	var err error = nil
	var reports []AggregateReport

	db := global.GetDb()
	err = db.Select("SELECT * FROM AggregateReport ar GROUP BY ar.group_id, ar.task_id ORDER BY ar.group_id, ar.task_id, ar.report_id DESC ", &reports)

	return &reports, err
}
