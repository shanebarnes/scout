package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "sync"

    "github.com/shanebarnes/scout/control"
    "github.com/shanebarnes/scout/execution"
    "github.com/shanebarnes/scout/mission"
    "github.com/shanebarnes/scout/situation"
)

func sigHandler(ch *chan os.Signal) {
    sig := <-*ch
    control.Stop()
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

    orderFile := flag.String("order", "order.json", "file containing scouting operations order")
    /*reportFile := */flag.String("report", "report.csv", "file containing scouting report")
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "version %s\n", mission.GetVersion())
        fmt.Fprintln(os.Stderr, "usage:")
        flag.PrintDefaults()
    }
    flag.Parse()

    order := loadOrder(orderFile)

    arr, err := situation.Parse(&order.Situation)
    if (err != nil ) {
        fmt.Println(err)
        os.Exit(1)
    }

    tasks, err2 := execution.Parse(&order.Execution)
    control.TASKS = &tasks
    if (err2 != nil ) {
        fmt.Println(err2)
        os.Exit(1)
    }

    targets := make([]situation.Target, len(arr))
    channels := make([]chan string, len(arr))

    db := make([][]control.Database, len(arr))
    var wg sync.WaitGroup
    for i := range arr {
        db[i] = make([]control.Database, len(tasks))
        channels[i] = make(chan string, 1000)
        if arr[i].Target.Prot == "SSH" {
            test := new(situation.TargetSsh)
            situation.Target.New(test, arr[i], tasks)
            test.Impl.Ch = &channels[i]
            test.Impl.Wait = &wg
            targets[i] = test
        } else {
            test := new(situation.TargetExec)
            situation.Target.New(test, arr[i], tasks)
            test.Impl.Ch = &channels[i]
            test.Impl.Wait = &wg
            targets[i] = test
        }
    }

    control.DATABASE = &db
    wg.Add(len(targets))
    for i := range targets {
        go situation.Target.Watch(targets[i])
    }
    go control.HandleRequests()
    control.ReportThread(targets)
    wg.Wait()
}

func initLog() {
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func getPrettyJson(v interface{}) string {
    buffer, err := json.MarshalIndent(v, "", "    ")
    if (err != nil) {

    }
    return string(buffer)
}

func loadOrder(fileName *string) Order {

    file, _ := os.Open(*fileName)
    decoder := json.NewDecoder(file)
    order := Order{}
    err := decoder.Decode(&order)
    if err != nil {
        fmt.Println("error: ", err)
    }

    return order
}
