package main

import (
    "encoding/csv"
    "os"
    "strconv"
    "strings"

    gc "github.com/rthornton128/goncurses"
)

func ReportThread(t []target) {
    file, _ := os.Create("report.csv")
    defer file.Close()

    writer := csv.NewWriter(file)
    writer.Comma = '\t'
    defer writer.Flush()

    //writer.Write([]string{"target", "operation", "date", "diff", "rate", "raw"})
    writer.Write([]string{"target", "operation", "type", "date", "value"})
    writer.Flush()

    for {
        m := 0
        for i := range t {
            impl := target.GetImpl(t[i])
            addr := impl.conf.Target.Addr
            if impl.conf.Target.Prot == "EXEC" {
                addr = "127.0.0.1"
            }
            _stdscr.ColorOn(gc.C_CYAN)
            _stdscr.MovePrintf(m, 0, "%2d: Target Name: %s, Addr: %s, Sys: %s\n", i, impl.conf.Target.Name, addr, impl.conf.Target.Sys)
            _stdscr.ColorOff(gc.C_CYAN)
            _stdscr.ClearToEOL()
            m++

            reportsReady := 0
            for reportsReady == 0 {
                for range impl.task {
                    if _, err := target.Report(t[i]); err == nil {
                        reportsReady = reportsReady + 1
                    }
                }
            }

            var data = [][]string{{}}
            for j := range impl.db {
                //var record = []string{addr, impl.task[j].Desc, strconv.FormatFloat(impl.db[j].dpN.x, 'f', 0, 64)}

                for k := range impl.task[j].Exec.Reports {
                    val := 0.
                    prefix := ""
                    report := strings.ToLower(impl.task[j].Exec.Reports[k])
                    switch report {
                        case "diff":
                            val = impl.db[j].diff
                        case "rate":
                            val = impl.db[j].rate
                        case "raw":
                            val = impl.db[j].dpN.y
                        default:
                            val = 0.
                    }

                    data = append(data, []string{addr, impl.task[j].Desc, report, strconv.FormatFloat(impl.db[j].dpN.x, 'f', 0, 64), strconv.FormatFloat(val, 'f', 3, 64)})
                    //record = append(record, strconv.FormatFloat(val, 'f', 3, 64))
                    val = val * impl.db[j].scale[k]
                    val, prefix = ToUnits(val, 10)
                    _stdscr.MovePrintf(m, 0, "    %4d: %-32s [%-4s] %7.3f %s%s", impl.db[j].N, impl.task[j].Desc, strings.ToLower(impl.task[j].Exec.Reports[k]), val, prefix, impl.db[j].units[k])
                    _stdscr.ClearToEOL()
                    m++
                }
                //data = append(data, record)
            }
            writer.WriteAll(data)
            _stdscr.Refresh()
        }
    }
}
