package main

import (
    "bytes"
    "encoding/json"
    "errors"
    "flag"
    "fmt"
    "golang.org/x/crypto/ssh"
    "io/ioutil"
    "log"
    "os"
    "os/signal"
    "strings"
    "syscall"
    "sync"
    "time"

    gc "github.com/rthornton128/goncurses"
)

const _VERSION string = "0.1.0"
var _stdscr *gc.Window = nil

func sigHandler(ch *chan os.Signal) {
    sig := <-*ch
    gc.End()
    fmt.Println("Captured sig", sig)
    os.Exit(3)
}

func main() {
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs,
                  syscall.SIGHUP,
                  syscall.SIGINT,
                  syscall.SIGQUIT,
                  syscall.SIGABRT,
                  syscall.SIGKILL,
                  syscall.SIGSEGV,
                  syscall.SIGTERM,
                  syscall.SIGSTOP)
    go sigHandler(&sigs)

    initLog()
    initGui()
    order := loadOrder()

    arr, err := parseSituation(&order.Situation)
    if (err != nil ) {
        fmt.Println(err)
        os.Exit(1)
    }

    tasks, err2 := parseExecution(&order.Execution)
    if (err2 != nil ) {
        fmt.Println(err2)
        os.Exit(1)
    }

    targets := findTargets(&arr)

    reports :=  make([]database, len(targets) * len(tasks))
    for i := range reports {
        reports[i] = NewDataBase(arr[i].Target.Name, tasks[i].Cmd, tasks[i].Scale, tasks[i].Units)
    }

    for {
        watchTargets(targets, &tasks, &reports)
        time.Sleep(time.Millisecond * 500)
    }
}

func initLog() {
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func initGui() {
    var err error
    _stdscr, err = gc.Init()
    if err != nil {
        gc.End()
        log.Fatal(err)
    }
    defer gc.End()

    gc.Echo(false)
    gc.CBreak(true)
    gc.Cursor(0)
    gc.StartColor()

    gc.InitPair(gc.C_WHITE, gc.C_WHITE, gc.C_BLACK)
    gc.InitPair(gc.C_YELLOW, gc.C_YELLOW, gc.C_BLACK)
    gc.InitPair(gc.C_RED, gc.C_RED, gc.C_BLACK)
    gc.InitPair(gc.C_CYAN, gc.C_CYAN, gc.C_BLACK)
}

func getPrettyJson(v interface{}) string {
    buffer, err := json.MarshalIndent(v, "", "    ")
    if (err != nil) {

    }
    return string(buffer)
}

func loadOrder() Order {
    option := flag.String("order", "order.json", "file containing scouting operations order")
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "version %s\n", _VERSION)
        fmt.Fprintln(os.Stderr, "usage:")
        flag.PrintDefaults()
    }
    flag.Parse()

    file, _ := os.Open(*option)
    decoder := json.NewDecoder(file)
    order := Order{}
    err := decoder.Decode(&order)
    if err != nil {
        fmt.Println("error: ", err)
    } else {
        //log.Println(getPrettyJson(order))
    }
    //log.Println("Targets:", len(order.Situation.Targets))
    //log.Println("Tasks  :", len(order.Execution.Tasks))
    return order
}

func parseSituation(situation *Situation) (TargetArr, error) {
    size := len(situation.Targets)
    ret := make(TargetArr, size)
    var err error = nil
    definitions := situation.Definitions
    credentials := situation.Credentials

    for i, id := range situation.Targets {
        var exists bool
        var target Target
        if target, exists = definitions[id]; exists {
            ret[i].Target = target
        } else {
            err = errors.New("Target '" + id + "' is not found in definitions")
            break
        }

        var cred Credentials
        if cred, exists = credentials[target.Cred]; exists {
            ret[i].Credentials = cred
        } else {
            err = errors.New("Target '" + id + "' credentials '" + target.Cred + "' not found")
            break
        }
    }

    if size == 0 && err == nil {
        err = errors.New("No targets found")
    }

    return ret, err
}

func parseExecution(execution *Execution1) (TaskArr, error) {
    size := len(execution.Tasks)
    ret := make(TaskArr, size)
    var err error = nil
    tasks := execution.Tasks
    definitions := execution.Definitions

    var i int = 0
    for _, task := range tasks {
        var def Task
        var exists bool
        if def, exists = definitions[task.Task]; exists {
            if len(task.Vars) == len(def.Vars) {
                cmd := def.Task
                for j, param := range task.Vars {
                    cmd = strings.Replace(cmd, def.Vars[j], param, 1)
                }
                ret[i].Exec = task
                ret[i].Cmd = cmd
                ret[i].Ret = def.Type
                ret[i].Scale = task.Scale
                ret[i].Units = task.Units
            } else {
                err = errors.New("Task '" + task.Task + "' vars do not match definitions")
                break
            }
        } else {
            err = errors.New("Task '" + task.Task + "' is not found in definitions")
            break
        }

        i++
    }

    if size == 0 && err == nil {
        err = errors.New("No tasks found")
    } else {
        i = 0
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

    return ret, err
}

func findTargets(arr *TargetArr) []*ssh.Client {
    size := len(*arr)
    clients := make([]*ssh.Client, size)
    channels := make([]chan *ssh.Client, size)
    //var wg sync.WaitGroup
    //wg.Add(count)

    for i := range *arr {
        _stdscr.MovePrintf(0, 0, "Connecting to %s...\n", (*arr)[i].Target.Addr)
        _stdscr.Refresh()

        channels[i] = make(chan *ssh.Client)
        go connectTarget(&(*arr)[i], /*&wg,*/ &channels[i])
    }

    //wg.Wait()
    for i := range channels {
        //select {
            /*case*/ clients[i] = <-channels[i]/*:*/
        //    default:
        //}
    }

    return clients
}

func connectTarget(target *TargetEntry, /*wg *sync.WaitGroup,*/ ch *chan *ssh.Client) {
    //defer wg.Done()
    var client *ssh.Client = nil
    var config *ssh.ClientConfig = nil

    if len(target.Credentials.Pass) > 0 {
        config = &ssh.ClientConfig {
            User: target.Credentials.User,
            Auth: []ssh.AuthMethod {
                ssh.Password(target.Credentials.Pass),
            },
            Timeout: 2000 * time.Millisecond,
        }
    } else if len(target.Credentials.Cert) > 0 {
        file, _ := ioutil.ReadFile(target.Credentials.Cert)
        signer, err := ssh.ParsePrivateKey(file)
        if err != nil {
            log.Fatal(err)
        }
        config = &ssh.ClientConfig {
            User: target.Credentials.User,
            Auth: []ssh.AuthMethod {
                ssh.PublicKeys(signer),
            },
            Timeout: 2000 * time.Millisecond,
        }
    }

    client, err := ssh.Dial("tcp", target.Target.Addr + ":22", config)
    if err != nil {
        log.Fatal("Failed to dial: ", err)
    } else {
        //select {
        /*    case*/ *ch <- client/*:*/
        //    default:
        //}
    }
}

func watchTargets(targets []*ssh.Client, tasks *TaskArr, reports *[]database) {
    var wg sync.WaitGroup
    channels := make([]chan string, len(targets))

    wg.Add(len(targets))
    for i, target := range targets {
        channels[i] = make(chan string, 10)
        go observeTarget(target, (*tasks)[0].Cmd, &wg, &channels[i])
    }
    wg.Wait()

    for i := range channels {
        value := <-channels[i]
        timeval := uint64(time.Now().UnixNano()) / uint64(time.Millisecond)

        dp, _ := NewDataPoint(timeval, value)
        Evaluate(&dp, &(*reports)[i])
        //switch (*tasks)[0].Ret {
        //    case "bool":
        //    case "float64":
        //    case "int64":
        //    case "uint64":
        //    default:
        //        _stdscr.MovePrintf(10, 0, "Unknown return type: %s, %s", (*tasks)[0].Ret, value)
        //        _stdscr.Refresh()
        //}
    }
    reportTargets(reports)
}

func observeTarget(client *ssh.Client, cmd string, wg *sync.WaitGroup, ch *chan string) {
    defer wg.Done()
    session, err := client.NewSession()
    if err != nil {
        log.Fatal("Failed to create session: ", err)
    }
    defer session.Close()

    var b bytes.Buffer
    session.Stdout = &b
    if err := session.Run(cmd); err != nil {
        log.Fatal("Failed to run: " + err.Error())
    }

    val := strings.Trim(b.String(), " \n")
    select {
        case *ch <- val:
        default:
    }
}

func reportTargets(reports *[]database) {
    var maxTicks int = 20
    var yScale uint64 = 50
    var total uint64 = 0
    _stdscr.Move(0,0)

    for i := range *reports {
        x := 0
        _stdscr.MovePrintf(i, x, "%-15s", (*reports)[i].target)

        x += 15
        _stdscr.MovePrintln(i, x, "[")

        x += 1
        ticks := uint64((*reports)[i].rate) / yScale

        for j := 0; j < maxTicks; j++ {
            if (j * 100 / maxTicks >= 66) {
                _stdscr.ColorOn(gc.C_RED)
            } else if (j * 100 / maxTicks >= 33) {
                _stdscr.ColorOn(gc.C_YELLOW)
            } else {
                _stdscr.ColorOn(gc.C_CYAN)
            }

            if int(ticks) > j {
                _stdscr.MovePrintln(i, x, "|")
            } else {
                _stdscr.MovePrintln(i, x, " ")
            }

            x++
        }

        _stdscr.ColorOff(gc.C_RED)
        _stdscr.ColorOff(gc.C_YELLOW)
        _stdscr.ColorOff(gc.C_CYAN)
        _stdscr.MovePrintf(i, x, "%6d %s] %s\n", uint64((*reports)[i].rate), (*reports)[i].units, (*reports)[i].task)

        total += uint64((*reports)[i].rate)
    }

    _stdscr.MovePrintf(len(*reports), 0,  "%-15s", "Total")
    _stdscr.MovePrintf(len(*reports), 36, "%6d Mbps\n", total)
    _stdscr.Refresh()
}
