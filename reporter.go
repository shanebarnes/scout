package main

func ReportThread(t []target) {
    for {
        for i := range t {
            impl := target.GetImpl(t[i])
            for range impl.task {
                if _, err := target.Report(t[i]); err == nil {
                }
            }
            for j := range impl.db {
                _stdscr.MovePrintf(i+j, 0, "%2d: [%-64s] %.3f", i+j, impl.db[j].task, impl.db[j].rate)
                _stdscr.Refresh()
            }
        }
    }
}
