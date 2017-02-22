package main


func ReportThread(t []target) {
    for {
        for i := range t {
            if db, err := target.Report(t[i]); err == nil {
                _stdscr.MovePrintf(i, 0, "%d, %s: %.3f", i, db.task, db.rate)
                _stdscr.Refresh()
            }
        }
    }
}
