package control

import (
    //"os"
    //"strconv"
    //"strings"
    //"time"

//    "github.com/shanebarnes/goto/logger"
//    "github.com/shanebarnes/scout/global"
    "github.com/shanebarnes/scout/situation"
)

//var _targets []situation.Target

func Stop() {}

func Init(t []situation.Target) {
//    _targets = t

    db := make([][]Database, len(t))

    for i := range t {
        impl := situation.Target.GetImpl(t[i])
        db[i] = make([]Database, len(impl.Task))
        for j := range impl.Task {
            db[i][j] = NewDataBase(impl.Conf.Target.Name, impl.Conf.Target.Addr, impl.Task[j].Desc, impl.Task[j].Exec.Reports)
        }
    }

//    REPORTS = &db
}

func ReportThread() {
    //db := global.GetDb()
    //var report taskReport
    //db.CreateTable(&report, "task_reports")

    //for {
    //    groupReports := 0

    //    for i := range _targets {
    //        impl := situation.Target.GetImpl(_targets[i])

    //        taskReports := 0
    //        for s := range impl.Task {
    //            if obs, err := situation.Target.Report(_targets[i]); err == nil {
    //                //dp, _ := NewDataPoint(obs.Idx, obs.Tv, obs.Dur, obs.Val)

    //                //Evaluate(&dp, &((*REPORTS)[i][s]))//obs.Idx]))

    //                //report.TargetId = int64(i)
    //                //report.TaskId = int64(obs.Idx)
    //                //report.X_Val = dp.X / 1000
    //                //report.Y_Val = dp.Y
    //                //db.InsertInto(&report, "task_reports")

    //                //val = strconv.FormatInt(int64(obs.idx), 16)
    //                //val = t.db[0].rate
    //                //db = &tdb[i][obs.idx]

    //                taskReports = taskReports + 1
    //            }
    //        }

    //        if taskReports > 0 {
    //            logger.PrintlnDebug("Received " + strconv.Itoa(taskReports) + " reports(s) for " + strconv.Itoa(len(_targets)) + " target(s)")
    //            groupReports = groupReports + 1
    //        } //else {
    //        //    continue
    //        //}
    //    }

    //    if groupReports == 0 {
    //        time.Sleep(time.Millisecond * 100)
    //    }
    //}
}
