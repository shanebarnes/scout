package main

import (
    "os/exec"
    "strconv"
    "strings"
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

    //_stdscr.MovePrintf(0, 0, "Finding target %s...\n", t.impl.conf.Target.Addr)
    //_stdscr.Refresh()

    return nil
}

func (t TargetExec) Watch() error {
    //var buffer []byte
    var err error = nil
    defer t.impl.wait.Done()

    for {
        for i := range t.impl.task {
            if buffer, err1 := exec.Command("bash", "-c", t.impl.task[i].Cmd).Output(); err1 == nil {
                value := strings.Trim(string(buffer[:]), " \r\n")

                select {
                    case *t.impl.ch <- strconv.Itoa(i):
                    default:
                }
                select {
                    case *t.impl.ch <- value:
                    default:
                }
                /*select {
                    case *t.impl.ch <- value:
                    default:
                }*/
            } else {
                // Command failed
            }
        }

        time.Sleep(time.Millisecond * 500)
    }

    return err
}

func (t *TargetExec) Report() (*database, error) {
    return ReportImpl(&t.impl)
}

func (t *TargetExec) GetImpl() *TargetImpl {
    return &t.impl
}

func (t *TargetExec) IsLost() bool {
    return true
}
