package situation

import (
    "bytes"
    "encoding/json"
    "errors"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/shanebarnes/scout/execution"
    "github.com/shanebarnes/scout/global"
)

type TargetImpl struct {
    Id int
    Conf TargetEntry
    Task execution.TaskArray
    cmds string
    Ch *chan string
    NextWatch time.Time
    Wait *sync.WaitGroup
    Db *global.DbImpl
}

type Target interface {
    New(id int, conf TargetEntry, t execution.TaskArray) error
    Find() error
    Watch() error
    Report() (*TargetObs, error)
    GetImpl() *TargetImpl
}

type TargetInfo struct {
    Info []string `json:"info"`
}

type TargetObs struct {
    Idx int
    Tv uint64
    Dur string
    Val string
}

func NewImpl(t *TargetImpl, id int, conf TargetEntry, tasks execution.TaskArray) error {
    var err error = nil
    var cmdBuffer, valBuffer bytes.Buffer

    t.Id = id
    t.Conf = conf
    // Todo: move to scout parsing
    for i := range tasks {
        for j := range conf.Target.Sys {
            if conf.Target.Sys[j] == tasks[i].Exec.Sys {
                t.Task = append(t.Task, tasks[i])
            }
        }
    }

    // @todo Run all commands in parallel and do not assume bash target
    // environment.
    for i := range t.Task {
        cmdBuffer.WriteString("val" + strconv.Itoa(i) + "=$(" + t.Task[i].Cmd + ");")
        if i > 0 {
            valBuffer.WriteString("printf \",\\\"$val" + strconv.Itoa(i) + "\\\"\";")
        } else {
            valBuffer.WriteString("printf \"\\\"$val" + strconv.Itoa(i) + "\\\"\";")
        }
    }

    t.cmds = cmdBuffer.String() + "echo -n '{\"info\":[';" + valBuffer.String() + "echo ']}'"
    t.NextWatch = time.Now()

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
        for i := range info.Info {
            value := strings.Trim(string(info.Info[i]), " \r\n")
            select {
                case *t.Ch <- strconv.Itoa(i):
                default:
            }
            select {
                case *t.Ch <- value:
                default:
            }
            select {
                case *t.Ch <- obsTime.String():
                default:
            }
        }
    } else {
        // Observation data parsing failed
    }

    // This is where a SQLite3 transaction should occur

    return err
}

func ReportImpl(t *TargetImpl) (*TargetObs, error) {
    var err error = nil
    obs := TargetObs{ Idx: -1, Tv: uint64(time.Now().UnixNano()) / uint64(time.Millisecond) }

    // @todo Use a single struct for value and duration
    if obs.Val, err = RecvFrom(t.Ch); err == nil {
        if obs.Idx, err = strconv.Atoi(obs.Val); err == nil {
            obs.Val, err = RecvFrom(t.Ch)
            obs.Dur, err = RecvFrom(t.Ch)
        }
    }

    return &obs, err
}
