package execution

import (
    "errors"
    "sort"
    "strings"

    "github.com/shanebarnes/scout/global"
)

// define task as operation instead
type Task struct {
    Reqs []string `json:"reqs"`
    Vars []string `json:"vars"`
    Type string `json:"type"`
    Task string `json:"task"`
}
type TaskMap map[string]Task

type TaskEntry struct {
    Id   int64          `db:"id"                   sql:"id INTEGER NOT NULL"`
    GroupId int         `db:"group_id"             sql:"group_id INTEGER NOT NULL"`
    Name string         `db:"name"                 sql:"name TEXT NOT NULL"`
    Active int          `db:"active"               sql:"active INTEGER NOT NULL"`
    Exec ExecutionGroup `db:"-"           json:"-" sql:"-"`
    Cmd  string         `db:"command"              sql:"command TEXT NOT NULL"`
    Desc string         `db:"description"          sql:"description TEXT NOT NULL"`
    Ret  string         `db:"-"           json:"-" sql:"-"`
}

type TaskGroup struct {
    Id   int    `sql:"id INTEGER NOT NULL PRIMARY KEY"`
    Name string `sql:"name TEXT NOT NULL"`
}
type TaskGroupMap map[string]TaskGroup

type TaskReport struct {
    TaskId   int64  `sql:"task_id INTEGER NOT NULL"`
    ReportId int64  `sql:"report_id INTEGER NOT NULL"`
    Type     string `json:"type"   sql:"type TEXT NOT NULL"`
    Xform    string `json:"xform"  sql:"xform TEXT NOT NULL"`
    Units    string `json:"units"  sql:"units TEXT NOT NULL"`
    Widget   string `json:"widget" sql:"widget TEXT NOT NULL"`
}

type ExecutionGroup struct {
    Active  bool         `json:"active"  sql:"active INTEGER NOT NULL"`
    Sys     string       `json:"sys"     sql:"system TEXT NOT NULL"`
    Desc    []string     `json:"desc"    sql:"description TEXT NOT NULL"`
    Task    string       `json:"task"    sql:"task TEXT NOT NULL"`
    Vars    [][]string   `json:"vars"    sql:"-"`
    Reports []TaskReport `json:"reports" sql:"-"`
}
type ExecutionMap map[string]ExecutionGroup

type Execution struct {
    Tasks ExecutionMap `json:"tasks"`
    Definitions TaskMap `json:"definitions"`
}

type TaskArray []TaskEntry

func (slice TaskArray) Len() int {
    return len(slice)
}

func (slice TaskArray) Less(i, j int) bool {
    return slice[i].Exec.Desc[0] < slice[j].Exec.Desc[0];
}

func (slice TaskArray) Swap(i, j int) {
    slice[i], slice[j] = slice[j], slice[i]
}

func Parse(exec *Execution) (TaskArray, error) {
    // @todo Return a reference?

    db := global.GetDb()
    var entry TaskEntry
    db.CreateTable(&entry)
    var group TaskGroup
    var taskGroup TaskGroupMap
    db.CreateTable(&group)
    var tr TaskReport
    db.CreateTable(&tr)
    taskGroup = make(map[string]TaskGroup)

    size := 0
    ret := make(TaskArray, 0)
    var err error = nil
    tasks := exec.Tasks
    definitions := exec.Definitions

    for _, task := range tasks {
        var def Task
        var exists bool
        if def, exists = definitions[task.Task]; exists {
            for j, vars := range task.Vars {
                if len(vars) == len(def.Vars) {
                    cmd := def.Task
                    for k, param := range vars {
                        cmd = strings.Replace(cmd, def.Vars[k], param, 1)
                    }

                    var gid TaskGroup
                    var ok bool
                    if gid, ok = taskGroup[task.Sys]; !ok {
                        gid.Id = len(taskGroup)
                        gid.Name = task.Sys
                        taskGroup[task.Sys] = gid
                    }

                    var entry TaskEntry
                    entry.Id = int64(size)
                    entry.GroupId = gid.Id
                    entry.Name = task.Task
                    if task.Active {
                        entry.Active = 1
                    } else {
                        entry.Active = 0
                    }
                    entry.Exec = task
                    entry.Cmd = cmd
                    entry.Desc = task.Desc[j]
                    entry.Ret = def.Type


                    db.InsertInto(&entry)
                    db.InsertInto(&gid)
                    //db.InsertInto(&entry.Exec, "task_group_definitions")
                    for m := range entry.Exec.Reports {
                        entry.Exec.Reports[m].TaskId = int64(size)
                        entry.Exec.Reports[m].ReportId = int64(m)
                        db.InsertInto(&entry.Exec.Reports[m])
                    }

                    ret = append(ret, entry)
                    size = size + 1
                } else {
                    err = errors.New("Task '" + task.Task + "' vars do not match definitions")
                    break
                }
            }
        } else {
            err = errors.New("Task '" + task.Task + "' is not found in definitions")
            break
        }
    }

    if size == 0 && err == nil {
        err = errors.New("No tasks found")
    } else {
        i := 0
        for _, task := range ret {
            if task.Exec.Active {
                ret[i] = task
                i++
            }
        }
        if i == 0 {
            err = errors.New("No active tasks found")
        } else {
            ret = ret[:i]
        }
    }

    sort.Sort(ret)

    return ret, err
}
