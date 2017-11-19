package situation

import (
    "os/exec"
    "time"

    "github.com/shanebarnes/scout/execution"
)

type TargetExec struct {
    Impl TargetImpl
}

func (t *TargetExec) New(id int, conf TargetEntry, tasks execution.TaskArray) error {
    return NewImpl(&t.Impl, id, conf, tasks)
}

func (t TargetExec) Find() error {
    defer t.Impl.Wait.Done()
    return nil
}

func (t TargetExec) Watch() error {
    var err error = nil
    defer t.Impl.Wait.Done()

    for {
        start := time.Now()
        if buffer, err := exec.Command("bash", "-c", t.Impl.cmds).Output(); err == nil {
            RecordImpl(&t.Impl, buffer, time.Since(start))
        } else {
            // Command failed
        }

        t.Impl.NextWatch = t.Impl.NextWatch.Add(1000 * time.Millisecond)
        for time.Since(t.Impl.NextWatch).Nanoseconds() / int64(time.Millisecond) >= 1000 {
            t.Impl.NextWatch = t.Impl.NextWatch.Add(1000 * time.Millisecond)
        }

        time.Sleep(t.Impl.NextWatch.Sub(time.Now()))
    }

    return err
}

func (t *TargetExec) Report() (*TargetObs, error) {
    return ReportImpl(&t.Impl)
}

func (t *TargetExec) GetImpl() *TargetImpl {
    return &t.Impl
}
