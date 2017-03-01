package main

import (
    "errors"
    "strconv"
    "sync"
    "time"
)

type TargetImpl struct {
    conf TargetEntry
    task TaskArr
    db []database
    ch *chan string
    wait *sync.WaitGroup
}

type target interface {
    New(conf TargetEntry, t TaskArr) error
    Find() error
    Watch() error
    Report() (*database, error)
    GetImpl() *TargetImpl
    IsLost() bool
}

func NewImpl(t *TargetImpl, conf TargetEntry, tasks TaskArr) error {
    var err error = nil

    t.conf = conf

    // Todo: move to scout parsing
    for i := range tasks {
        for j := range conf.Target.Sys {
            if conf.Target.Sys[j] == tasks[i].Sys {
                t.task = append(t.task, tasks[i])
            }
        }
    }

    t.db =  make([]database, len(t.task))
    for i := range t.db {
        t.db[i] = NewDataBase(conf.Target.Name, t.task[i].Cmd, t.task[i].Scale, t.task[i].Units)
    }

    return err
}

func RecvFrom(ch *chan string) (string, error) {
    var err error = nil
    var val string

    select {
        case val = <-*ch:
        case <-time.After(time.Millisecond * 10):
            err = errors.New("Recv channel timeout")
    }

    return val, err
}

func ReportImpl(t *TargetImpl) (*database, error) {
    var db *database = nil
    var err error = nil
    var idx int = -1
    var val string
    tv := uint64(time.Now().UnixNano()) / uint64(time.Millisecond)

    if val, err = RecvFrom(t.ch); err == nil {
        if idx, err = strconv.Atoi(val); err == nil {
            val, err = RecvFrom(t.ch)
        }
    }

    if err == nil {
        dp, _ := NewDataPoint(tv, val)
        Evaluate(&dp, &t.db[idx])
        //val = strconv.FormatInt(int64(idx), 16)
        //val = t.db[0].rate
        db = &t.db[idx]
    }

    return db, err
}
