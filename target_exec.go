package main

import (
    "os/exec"
    "time"
)

type TargetExec struct {
    impl TargetImpl
}

func (t *TargetExec) New(conf TargetEntry, tasks TaskArr) error {
    return NewImpl(&t.impl, conf, tasks)
}

func (t TargetExec) Find() error {
    defer t.impl.wait.Done()
    return nil
}

func (t TargetExec) Watch() error {
    var err error = nil
    defer t.impl.wait.Done()

    for {
        start := time.Now()
        if buffer, err := exec.Command("bash", "-c", t.impl.cmds).Output(); err == nil {
            RecordImpl(&t.impl, buffer, time.Since(start))
        } else {
            // Command failed
        }

        t.impl.nextWatch = t.impl.nextWatch.Add(1000 * time.Millisecond)
        for time.Since(t.impl.nextWatch).Nanoseconds() / int64(time.Millisecond) >= 1000 {
            t.impl.nextWatch = t.impl.nextWatch.Add(1000 * time.Millisecond)
        }

        time.Sleep(t.impl.nextWatch.Sub(time.Now()))
    }

    return err
}

func (t *TargetExec) Report() (*database, error) {
    return ReportImpl(&t.impl)
}

func (t *TargetExec) GetImpl() *TargetImpl {
    return &t.impl
}
