package main

type Credentials struct {
    User string `json:"user"`
    Pass string `json:"pass"`
    Cert string `json:"cert"`
}
type CredentialsMap map[string]Credentials

type TargetGroup struct {
    Name string `json:"name"`
    Addr []string `json:"addr"`
    Cred string `json:"cred"`
    Prot string `json:"prot"`
    Sys []string `json:"sys"`
}

type Target struct {
    Name string `json:"name"`
    Addr string `json:"addr"`
    Cred string `json:"cred"`
    Prot string `json:"prot"`
    Sys []string `json:"sys"`
}
type TargetMap map[string]TargetGroup

type TargetEntry struct {
    Target Target
    Credentials Credentials
}
type TargetArr []TargetEntry
// define task as operation instead
type Task struct {
    Reqs []string `json:"reqs"`
    Vars []string `json:"vars"`
    Type string `json:"type"`
    Task string `json:"task"`
}
type TaskMap map[string]Task

type TaskEntry struct {
    Exec Execution
    Cmd string
    Desc string
    Ret string
    Scale []float64
    Units []string
}
type TaskArr []TaskEntry

type Execution struct {
    Active bool `json:"active"`
    Sys string `json:"sys"`
    Desc []string `json:"desc"`
    Task string `json:"task"`
    Vars [][]string `json:"vars"`
    Reports []string `json:"reports"`
    Scale []float64 `json:"scale"`
    Units []string `json:"units"`
}
type ExecutionMap map[string]Execution

type Report1 struct {
    Op string `json:"op"`
}
type ReportMap map[string]Report1

type Execution1 struct {
    Tasks ExecutionMap `json:"tasks"`
    Definitions TaskMap `json:"definitions"`
}

type Situation struct {
    Targets []string `json:"targets"`
    Definitions TargetMap `json:"definitions"`
    Credentials CredentialsMap `json:"credentials"`
}

type Protocol struct {
   Protocol []string `json:"protocol"`
}

type Control1 struct {
    Frequency string `json:"frequency"`
    Duration string `json:"duration"`
    Reports ReportMap `json:"reports"`
}

type Order struct {
    Mission string `json:"mission"`
    Situation Situation `json:"situation"`
    Execution Execution1 `json:"execution"`
    Sustainment Protocol `json:"sustainment"`
    Control Control1 `json:"control"`
}

type Report struct {
    Timestamp uint64
    Value     uint64
    Rate      uint64
}

func (slice TaskArr) Len() int {
    return len(slice)
}

func (slice TaskArr) Less(i, j int) bool {
    return slice[i].Exec.Desc[0] < slice[j].Exec.Desc[0];
}

func (slice TaskArr) Swap(i, j int) {
    slice[i], slice[j] = slice[j], slice[i]
}
