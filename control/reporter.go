package control

import (
    "encoding/csv"
    "log"
    "os"
    "strconv"
    "strings"
    "time"

    gc "github.com/rthornton128/goncurses"
    "github.com/shanebarnes/scout/situation"
)

var _stdscr *gc.Window = nil

func Stop() {
    gc.End()
}

func initGui() {
    var err error
    _stdscr, err = gc.Init()
    if err != nil {
        gc.End()
        log.Fatal(err)
    }
    defer gc.End()

    gc.Echo(false)
    gc.CBreak(true)
    gc.Cursor(0)
    gc.StartColor()

    gc.InitPair(gc.C_WHITE, gc.C_WHITE, gc.C_BLACK)
    gc.InitPair(gc.C_YELLOW, gc.C_YELLOW, gc.C_BLACK)
    gc.InitPair(gc.C_RED, gc.C_RED, gc.C_BLACK)
    gc.InitPair(gc.C_CYAN, gc.C_CYAN, gc.C_BLACK)
}

func ReportThread(t []situation.Target) {
    initGui()

    tdb := make([][]Database, len(t))
    //tdb []database
    //tdb = make([]database, len(t.task))
    for i := range t {
        impl := situation.Target.GetImpl(t[i])
        tdb[i] = make([]Database, len(impl.Task))
        for j := range impl.Task {
            //t.db[i][j] = NewDataBase(conf.Target.Name, t.task[i].Cmd, t.task[i].Scale, t.task[i].Units)
            tdb[i][j] = NewDataBase(impl.Conf.Target.Name, impl.Task[j].Desc, /*impl.Task[j].Form,*/ impl.Task[j].Scale, impl.Task[j].Units)
        }
    }

    file, _ := os.Create("report.csv")
    defer file.Close()

    writer := csv.NewWriter(file)
    writer.Comma = '\t'
    defer writer.Flush()

    //writer.Write([]string{"target", "operation", "date", "diff", "rate", "raw"})
    //writer.Flush()

    for {
        groupReports := 0
        row := 0
        for i := range t {
            impl := situation.Target.GetImpl(t[i])
            addr := impl.Conf.Target.Addr
            if impl.Conf.Target.Prot == "EXEC" {
                addr = "127.0.0.1"
            }
            _stdscr.ColorOn(gc.C_CYAN)
            _stdscr.MovePrintf(row,
                               0,
                               "%2d: Target Name: %s, Addr: %s, Sys: %s\n",
                               i,
                               impl.Conf.Target.Name,
                               addr,
                               impl.Conf.Target.Sys)
            _stdscr.ColorOff(gc.C_CYAN)
            _stdscr.ClearToEOL()
            row++

            taskReports := 0
            for range impl.Task {
                if obs, err := situation.Target.Report(t[i]); err == nil {
                    dp, _ := NewDataPoint(obs.Tv, obs.Dur, obs.Val)
                    Evaluate(&dp, &tdb[i][obs.Idx])
                    //val = strconv.FormatInt(int64(obs.idx), 16)
                    //val = t.db[0].rate
                    //db = &tdb[i][obs.idx]

                    taskReports = taskReports + 1
                }
            }

            if taskReports > 0 {
                groupReports = groupReports + 1
            } //else {
            //    continue
            //}

            var data = [][]string{{}}
            for j := range tdb[i]/*impl.db*/ {
                record := []string{addr,
                                   impl.Task[j].Desc,
                                   strconv.FormatFloat(/*impl.*/tdb[i][j].DpN.X / 1000., 'f', 3, 64),
                                   "0.000",
                                   "0.000",
                                   "0.000"}

                for k := range impl.Task[j].Exec.Reports {
                    val := 0.
                    prefix := ""
                    report := strings.ToLower(impl.Task[j].Exec.Reports[k])
                    switch report {
                        case "diff":
                            val = /*impl.*/tdb[i][j].Diff
                            record[3] = strconv.FormatFloat(val, 'f', 3, 64)
                        case "rate":
                            val = /*impl.*/tdb[i][j].Rate
                            record[4] = strconv.FormatFloat(val, 'f', 3, 64)
                        case "raw":
                            val = /*impl.*/tdb[i][j].DpN.Y
                            record[5] = strconv.FormatFloat(val, 'f', 3, 64)
                        default:
                            val = 0.
                    }

                    val = val * /*impl.*/tdb[i][j].Scale[k]
                    val, prefix = ToUnits(val, 10)
                    _stdscr.MovePrintf(row,
                                       0,
                                       "    %4d: %-32s [%-4s] %7.3f %-6s (dur: %-12s)",
                                       /*impl.*/tdb[i][j].N,
                                       impl.Task[j].Desc,
                                       strings.ToLower(impl.Task[j].Exec.Reports[k]),
                                       val,
                                       prefix + /*impl.*/tdb[i][j].Units[k],
                                       /*impl.*/tdb[i][j].DpN.d)
                    _stdscr.ClearToEOL()
                    row++
                }
                data = append(data, record)
            }

            // This is not thread-safe
            (*REPORTS)[i] = tdb[i]
            //writer.WriteAll(data)
            _stdscr.Refresh()
        }

        if groupReports == 0 {
            time.Sleep(time.Millisecond * 1000)
        }
    }

}
