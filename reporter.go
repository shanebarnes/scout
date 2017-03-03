package main

func ReportThread(t []target) {
    for {
        k := 0
        for i := range t {
            impl := target.GetImpl(t[i])
            addr := impl.conf.Target.Addr
            if impl.conf.Target.Prot == "EXEC" {
                addr = "127.0.0.1"
            }
            _stdscr.MovePrintf(k, 0, "%2d: Target Name: %s, Addr: %s, Sys: %s\n", k, impl.conf.Target.Name, addr, impl.conf.Target.Sys)
            _stdscr.ClearToEOL()
            k++
            for range impl.task {
                if _, err := target.Report(t[i]); err == nil {
                }
            }
            for j := range impl.db {
                val := 0.
                switch impl.task[j].Exec.Reports[0] {
                    case "RATE":
                        val = impl.db[j].rate
                    case "RAW":
                        val = impl.db[j].dpN.y
                    default:
                        val = -1.
                }
                _stdscr.MovePrintf(k, 0, "%2d:     [%-96s] %.3f", k, impl.db[j].task, val)
                _stdscr.ClearToEOL()
                k++
            }
            _stdscr.Refresh()
        }
    }
}
