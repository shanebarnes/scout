package execution

import (
    "errors"
    "sort"
    "strings"

//    "github.com/shanebarnes/goto/logger"
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
    Exec ExecutionGroup
    Cmd string
    Desc string
    Ret string
}

type TaskReport struct {
    Type   string  `json:"type"`
    Scale  float64 `json:"scale"`
    Units  string  `json:"units"`
    Widget string  `json:"widget"`
}

type ExecutionGroup struct {
    Active  bool         `json:"active"`
    Sys     string       `json:"sys"`
    Desc    []string     `json:"desc"`
    Task    string       `json:"task"`
    Vars    [][]string   `json:"vars"`
    Reports []TaskReport `json:"reports"`
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
                    var entry TaskEntry
                    entry.Exec = task
                    entry.Cmd = cmd
                    entry.Desc = task.Desc[j]
                    entry.Ret = def.Type
                    //entry.Scale = task.Scale
                    //entry.Widget = task.Widget // Validate widget type
                    //if len(task.Units) == len(task.Reports) {
                    //    entry.Units = task.Units
                    //} else {
                    //    err = errors.New("Task '" + task.Task + "' reports and units lengths do not match")
                    //    break
                    //}
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
