package main

import (
    gc "github.com/rthornton128/goncurses"
)

func ReportThread(t []target) {
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
            for range impl.task {
                if _, err := target.Report(t[i]); err == nil {
                }
            }

            for j := range impl.db {
                for k := range impl.task[j].Exec.Reports {
                    val := 0.
                    prefix := ""
                    switch impl.task[j].Exec.Reports[k] {
                        case "RATE":
                            val = impl.db[j].rate
                        case "RAW":
                            val = impl.db[j].dpN.y
                        default:
                            val = -1.
                    }
                    val, prefix = ToUnits(val, 10)
                    _stdscr.MovePrintf(m, 0, "    %4d: [%-32s] %7.3f %s%s", impl.db[j].N, impl.task[j].Exec.Desc[k], val, prefix, impl.db[j].units)
                    _stdscr.ClearToEOL()
                    m++
                }
            }
            _stdscr.Refresh()
        }
    }
}
