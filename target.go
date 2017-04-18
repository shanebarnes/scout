package main

import (
    "bytes"
    "encoding/json"
    "errors"
    "strconv"
    "strings"
    "sync"
    "time"
)

type TargetImpl struct {
    conf TargetEntry
    task TaskArr
    cmds string
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
}

type TargetInfo struct {
    Info []string `json:"info"`
}

func NewImpl(t *TargetImpl, conf TargetEntry, tasks TaskArr) error {
    var err error = nil
    var cmdBuffer, valBuffer bytes.Buffer

    t.conf = conf
    // Todo: move to scout parsing
    for i := range tasks {
        for j := range conf.Target.Sys {
            if conf.Target.Sys[j] == tasks[i].Exec.Sys {
                t.task = append(t.task, tasks[i])
            }
        }
    }

    t.db =  make([]database, len(t.task))
    // @todo Run all commands in parallel and do not assume bash target
    // environment.
    for i := range t.db {
        t.db[i] = NewDataBase(conf.Target.Name, t.task[i].Cmd, t.task[i].Scale, t.task[i].Units)

        cmdBuffer.WriteString("val" + strconv.Itoa(i) + "=$(" + t.task[i].Cmd + ");")
        if i > 0 {
            valBuffer.WriteString("printf \",\\\"$val" + strconv.Itoa(i) + "\\\"\";")
        } else {
            valBuffer.WriteString("printf \"\\\"$val" + strconv.Itoa(i) + "\\\"\";")
        }
    }

    t.cmds = cmdBuffer.String() + "echo -n '{\"info\":[';" + valBuffer.String() + "echo ']}'"

    return err
}

func RecvFrom(ch *chan string) (string, error) {
    var err error = nil
    var val string

    select {
        case val = <-*ch:
        case <-time.After(time.Millisecond * 0):
            err = errors.New("Recv channel timeout")
    }

    return val, err
}

func RecordImpl(t *TargetImpl, obsData []byte, obsTime time.Duration) (error) {
    var err error = nil
    var info TargetInfo

    if err = json.Unmarshal(obsData, &info); err == nil {
        for j := range info.Info {
            value := strings.Trim(string(info.Info[j]), " \r\n")
            select {
                case *t.ch <- strconv.Itoa(j):
                default:
            }
            select {
                case *t.ch <- value:
                default:
            }
            select {
                case *t.ch <- obsTime.String():
                default:
            }
        }
    } else {
        // Observation data parsing failed
    }

    return err
}

func ReportImpl(t *TargetImpl) (*database, error) {
    var db *database = nil
    var err error = nil
    var idx int = -1
    var dur, val string
    tv := uint64(time.Now().UnixNano()) / uint64(time.Millisecond)

    if val, err = RecvFrom(t.ch); err == nil {
        if idx, err = strconv.Atoi(val); err == nil {
            val, err = RecvFrom(t.ch)
            dur, err = RecvFrom(t.ch)
        }
    }

    if err == nil {
        dp, _ := NewDataPoint(tv, dur, val)
        Evaluate(&dp, &t.db[idx])
        //val = strconv.FormatInt(int64(idx), 16)
        //val = t.db[0].rate
        db = &t.db[idx]
    }

    return db, err
}
