package control

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/shanebarnes/goto/logger"
	"github.com/shanebarnes/scout/global"
)

type statistics struct {
	//Groups  uint64 `db:"groups"`
	Reports uint64 `db:"reports"`
	//Targets uint64 `db:"targets"`
	//Tasks   uint64 `db:"tasks"`
}

func StatisticsHandler(w http.ResponseWriter, r *http.Request) {
	defer logger.PrintlnTime(logger.Info, time.Time{}, "Prepared response to "+r.Method+" "+r.URL.Path)()
	sql := "SELECT COUNT(*) AS reports FROM TargetReport"
	db := global.GetDb()

	var stats []statistics
	if err := db.Select(sql, &stats); err != nil {
		logger.PrintlnError(err)
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(stats)
}
