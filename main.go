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

    "github.com/shanebarnes/goto/logger"
    "github.com/shanebarnes/scout/control"
    "github.com/shanebarnes/scout/execution"
    "github.com/shanebarnes/scout/global"
    "github.com/shanebarnes/scout/mission"
    "github.com/shanebarnes/scout/situation"
)

func sigHandler(ch *chan os.Signal) {
    sig := <-*ch
    control.Stop()
    logger.PrintlnInfo("Captured sig " + sig.String())
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

    file, _ := os.OpenFile("scout.log", os.O_APPEND | os.O_CREATE | os.O_RDWR, 0644)
    defer file.Close()

    logger.Init(log.Ldate | log.Ltime | log.Lmicroseconds, logger.Info, file)

    logger.PrintlnInfo("Starting scout", mission.GetVersion())

    orderFile := flag.String("order", "order.json", "file containing scouting operations order")
    /*reportFile := */flag.String("report", "report.db", "file containing scouting report")
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "version %s\n", mission.GetVersion())
        fmt.Fprintln(os.Stderr, "usage:")
        flag.PrintDefaults()
    }
    flag.Parse()

    db := global.GetDb()
    db.Open("scout.db")
    defer db.Close()

    order := loadOrder(orderFile)

    arr, err := situation.Parse(&order.Situation)
    if (err != nil ) {
        logger.PrintlnError(err.Error())
        os.Exit(1)
    }

    tasks, err2 := execution.Parse(&order.Execution)
    //control.TASKS = &tasks
    if (err2 != nil ) {
        logger.PrintlnError(err2.Error())
        os.Exit(1)
    }

    if err = control.Parse(&order.Control); err != nil {
        logger.PrintlnError(err.Error())
        os.Exit(1)
    }

    targets := make([]situation.Target, len(arr))
    channels := make([]chan string, len(arr))

    var wg sync.WaitGroup
    for i := range arr {
        channels[i] = make(chan string, 1000)
        if arr[i].Target.Prot == "SSH" {
            ssh := new(situation.TargetSsh)
            ssh.New(i, arr[i], tasks)
            ssh.Impl.Ch = &channels[i]
            ssh.Impl.Wait = &wg
            targets[i] = ssh
        } else {
            exec := new(situation.TargetExec)
            exec.New(i, arr[i], tasks)
            exec.Impl.Ch = &channels[i]
            exec.Impl.Wait = &wg
            targets[i] = exec
        }
    }

    wg.Add(len(targets))
    for i := range targets {
        go situation.Target.Watch(targets[i])
    }

    control.Init(targets)
    go control.HandleRequests(&order.Control)
    //control.ReportThread()
    wg.Wait()
}

func loadOrder(fileName *string) Order {

    file, _ := os.Open(*fileName)
    decoder := json.NewDecoder(file)
    order := Order{}
    err := decoder.Decode(&order)
    if err != nil {
        logger.PrintlnError(err.Error())
    }

    return order
}
