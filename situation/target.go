package situation

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron"
	"github.com/shanebarnes/goto/logger"
	"github.com/shanebarnes/scout/execution"
	"github.com/shanebarnes/scout/global"
)

type TargetReport struct {
	ReportId int64   `db:"report_id" sql:"report_id INTEGER NOT NULL"`
	TargetId int64   `db:"target_id" sql:"target_id INTEGER NOT NULL"`
	TaskId   int64   `db:"task_id"   sql:"task_id INTEGER NOT NULL"`
	X_Diff   float64 `db:"x_diff"    sql:"x_diff FLOAT NOT NULL"`
	Y_Diff   float64 `db:"y_diff"    sql:"y_diff FLOAT NOT NULL"`
	Y_Max    float64 `db:"y_max"     sql:"y_max FLOAT NOT NULL"`
	Y_Min    float64 `db:"y_min"     sql:"y_min FLOAT NOT NULL"`
	Y_Rate   float64 `db:"y_rate"    sql:"y_rate FLOAT NOT NULL"`
	X_Val    float64 `db:"x_val"     sql:"x_val FLOAT NOT NULL"`
	Y_Val    float64 `db:"y_val"     sql:"y_val FLOAT NOT NULL"`
}

type TargetImpl struct {
	Db          *global.DbImpl
	Ch          *chan string
	cmds        string
	Conf        TargetEntry
	CronJob     *cron.Cron
	CronSpec    string
	Id          int
	RecordCache []TargetReport
	Task        execution.TaskArray
	Wait        *sync.WaitGroup
	WatchCount  int64
	WatchLimit  int64
}

type Target interface {
	New(id int, conf TargetEntry, t execution.TaskArray) error
	Find() error
	Watch() error
	Report() (*TargetObs, error)
	GetImpl() *TargetImpl
}

type TargetInfo struct {
	Info []string `json:"info"`
}

type TargetObs struct {
	Idx int
	Tv  uint64
	Dur string
	Val string
}

func NewImpl(t *TargetImpl, id int, conf TargetEntry, tasks execution.TaskArray) error {
	var err error = nil
	var cmdBuffer, valBuffer bytes.Buffer

	t.CronJob = cron.New()
	t.RecordCache = make([]TargetReport, len(tasks))

	t.Id = id
	t.Conf = conf
	// Todo: move to scout parsing
	for i := range tasks {
		for j := range conf.Target.Sys {
			if conf.Target.Sys[j] == tasks[i].Exec.Sys {
				t.Task = append(t.Task, tasks[i])
			}
		}
	}

	// @todo Run all commands in parallel and do not assume bash target
	// environment.
	for i := range t.Task {
		cmdBuffer.WriteString("val" + strconv.Itoa(i) + "=$(" + t.Task[i].Cmd + ");")
		if i > 0 {
			valBuffer.WriteString("printf \",\\\"$val" + strconv.Itoa(i) + "\\\"\";")
		} else {
			valBuffer.WriteString("printf \"\\\"$val" + strconv.Itoa(i) + "\\\"\";")
		}
	}

	t.cmds = cmdBuffer.String() + "echo -n '{\"info\":[';" + valBuffer.String() + "echo ']}'"

	t.Db = global.GetDb()
	t.Db.CreateTable(&TargetReport{})

	t.WatchCount = 0
	t.WatchLimit = 0

	return err
}

func RecvFrom(ch *chan string) (string, error) {
	var err error = nil
	var val string

	select {
	case val = <-*ch:
	case <-time.After(time.Millisecond * 0):
		err = errors.New("Recv channel timeout")
	}

	return val, err
}

func RecordImpl(t *TargetImpl, obsData []byte, obsTime time.Duration) error {
	var err error = nil
	var info TargetInfo

	if err = json.Unmarshal(obsData, &info); err == nil {
		var report TargetReport
		//var obs *TargetObs = nil
		for i := range info.Info {
			// @todo batch insert

			value := strings.Trim(string(info.Info[i]), " \r\n")
			report.ReportId = t.RecordCache[i].ReportId + 1
			report.TargetId = int64(t.Id)
			report.TaskId = t.Task[i].Id
			report.X_Val = float64(uint64(time.Now().UnixNano())/uint64(time.Millisecond)) / 1000
			report.Y_Max = t.RecordCache[i].Y_Max
			report.Y_Min = t.RecordCache[i].Y_Min
			report.Y_Val, err = strconv.ParseFloat(value, 64)

			if t.RecordCache[i].ReportId > 0 {
				report.X_Diff = report.X_Val - t.RecordCache[i].X_Val
				report.Y_Diff = report.Y_Val - t.RecordCache[i].Y_Val

				if report.Y_Val > report.Y_Max {
					report.Y_Max = report.Y_Val
				}

				if report.Y_Val < report.Y_Min {
					report.Y_Min = report.Y_Val
				}

				if report.X_Diff > 0 {
					report.Y_Rate = report.Y_Diff / report.X_Diff
				}
			} else {
				report.Y_Max = report.Y_Val
				report.Y_Min = report.Y_Val
			}

			// @bug Preparing the same sql statement repeatedly doesn't make sense
			t.Db.InsertInto(&report)
			t.RecordCache[i] = report
		}
	} else {
		// Observation data parsing failed
	}

	return err
}

func ReportImpl(t *TargetImpl) (*TargetObs, error) {
	var err error = nil
	obs := TargetObs{Idx: -1, Tv: uint64(time.Now().UnixNano()) / uint64(time.Millisecond)}

	// @todo Use a single struct for value and duration
	if obs.Val, err = RecvFrom(t.Ch); err == nil {
		if obs.Idx, err = strconv.Atoi(obs.Val); err == nil {
			obs.Val, err = RecvFrom(t.Ch)
			obs.Dur, err = RecvFrom(t.Ch)
		}
	}

	return &obs, err
}

func CheckWatchImpl(t *TargetImpl) {
	t.WatchCount = t.WatchCount + 1

	logger.PrintlnDebug("Completed watch", t.WatchCount, "of target", t.Id)

	if t.WatchLimit > -1 && t.WatchCount >= t.WatchLimit {
		StopWatchImpl(t)
	}
}

func StartWatchImpl(t *TargetImpl, fn func()) {
	t.Wait.Add(1)
	logger.PrintlnInfo("Started watching target", t.Id)

	t.CronJob.AddFunc(t.CronSpec, fn)
	t.CronJob.Start()
}

func StopWatchImpl(t *TargetImpl) {
	logger.PrintlnInfo("Stopped watching target", t.Id)
	t.CronJob.Stop()
	t.Wait.Done()
}
