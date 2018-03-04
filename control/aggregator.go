package control

import (
	"strconv"
	"time"

	"github.com/robfig/cron"
	"github.com/shanebarnes/goto/logger"
	"github.com/shanebarnes/scout/global"
)

type AggregateReport struct {
	ReportId int64   `db:"report_id" sql:"report_id INTEGER NOT NULL"`
	GroupId  int64   `db:"group_id"  sql:"group_id INTEGER NOT NULL"`
	TaskId   int64   `db:"task_id"   sql:"task_id INTEGER NOT NULL"`
	Targets  int64   `db:"targets"   sql:"targets INTEGER NOT NULL"`
	Xdiff    float64 `db:"x_diff"    sql:"x_diff FLOAT NOT NULL"`
	Xval     float64 `db:"x_val"     sql:"x_val FLOAT NOT NULL"`
	Ydiff    float64 `db:"y_diff"    sql:"y_diff FLOAT NOT NULL"`
	Ymax     float64 `db:"y_max"     sql:"y_max FLOAT NOT NULL"`
	Ymin     float64 `db:"y_min"     sql:"y_min FLOAT NOT NULL"`
	Yrate    float64 `db:"y_rate"    sql:"y_rate FLOAT NOT NULL"`
	Yval     float64 `db:"y_val"     sql:"y_val FLOAT NOT NULL"`
}

func RunAggregator(ctl *Control) {
	db := global.GetDb()
	db.CreateTable(&AggregateReport{})
	var reportId int64 = 0

	cron := cron.New()
	cron.AddFunc(ctl.Frequency, func() {
		entries := cron.Entries()
		if entries != nil && len(entries) > 0 {
			interval := float64(entries[0].Next.Sub(entries[0].Prev) / time.Second)
			timeval := float64(entries[0].Prev.UnixNano()) / float64(time.Second)
			reportTime := timeval - interval
			reportId = reportId + 1
			createReports(reportId, reportTime)
		}
	})
	cron.Start()
	//defer cron.Stop()
}

func createReports(reportId int64, reportTime float64) {
	db := global.GetDb()

	sql := "SELECT " + strconv.FormatInt(reportId, 10) + " AS report_id, d.group_id, r.task_id, COUNT(*) AS targets, AVG(r.x_diff) AS x_diff, MAX(r.x_val) AS x_val, AVG(r.y_diff) AS y_diff, MAX(r.y_max) AS y_max, MIN(r.y_min) AS y_min, SUM(r.y_rate) AS y_rate, SUM(r.y_val) AS y_val FROM TargetReport r LEFT JOIN TargetDef d ON d.id = r.target_id LEFT JOIN TargetGroup g ON g.id = d.group_id WHERE r.x_val >= " + strconv.FormatFloat(reportTime, 'f', -1, 64) + " GROUP BY d.group_id, r.task_id ORDER BY d.group_id, r.task_id"

	logger.PrintlnDebug(sql)

	var reports []AggregateReport
	db.Select(sql, &reports)
	db.InsertInto(&reports)
}
