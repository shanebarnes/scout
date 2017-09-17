package main

import (
    "encoding/csv"
    "os"
    "strconv"
    "strings"
    "time"

    gc "github.com/rthornton128/goncurses"
)

func ReportThread(t []target) {
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
            impl := target.GetImpl(t[i])
            addr := impl.conf.Target.Addr
            if impl.conf.Target.Prot == "EXEC" {
                addr = "127.0.0.1"
            }
            _stdscr.ColorOn(gc.C_CYAN)
            _stdscr.MovePrintf(row,
                               0,
                               "%2d: Target Name: %s, Addr: %s, Sys: %s\n",
                               i,
                               impl.conf.Target.Name,
                               addr,
                               impl.conf.Target.Sys)
            _stdscr.ColorOff(gc.C_CYAN)
            _stdscr.ClearToEOL()
            row++

            taskReports := 0
            for range impl.task {
                if _, err := target.Report(t[i]); err == nil {
                    taskReports = taskReports + 1
                }
            }

            if taskReports > 0 {
                groupReports = groupReports + 1
            } //else {
            //    continue
            //}

            var data = [][]string{{}}
            for j := range impl.db {
                record := []string{addr,
                                   impl.task[j].Desc,
                                   strconv.FormatFloat(impl.db[j].DpN.X / 1000., 'f', 3, 64),
                                   "0.000",
                                   "0.000",
                                   "0.000"}

                for k := range impl.task[j].Exec.Reports {
                    val := 0.
                    prefix := ""
                    report := strings.ToLower(impl.task[j].Exec.Reports[k])
                    switch report {
                        case "diff":
                            val = impl.db[j].Diff
                            record[3] = strconv.FormatFloat(val, 'f', 3, 64)
                        case "rate":
                            val = impl.db[j].Rate
                            record[4] = strconv.FormatFloat(val, 'f', 3, 64)
                        case "raw":
                            val = impl.db[j].DpN.Y
                            record[5] = strconv.FormatFloat(val, 'f', 3, 64)
                        default:
                            val = 0.
                    }

                    val = val * impl.db[j].Scale[k]
                    val, prefix = ToUnits(val, 10)
                    _stdscr.MovePrintf(row,
                                       0,
                                       "    %4d: %-32s [%-4s] %7.3f %-6s (dur: %-12s)",
                                       impl.db[j].N,
                                       impl.task[j].Desc,
                                       strings.ToLower(impl.task[j].Exec.Reports[k]),
                                       val,
                                       prefix + impl.db[j].Units[k],
                                       impl.db[j].DpN.d)
                    _stdscr.ClearToEOL()
                    row++
                }
                data = append(data, record)
            }

            // This is not thread-safe
            (*_database)[0] = impl.db
            //writer.WriteAll(data)
            _stdscr.Refresh()
        }

        if groupReports == 0 {
            time.Sleep(time.Millisecond * 1000)
        }
    }
}
