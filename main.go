package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/shanebarnes/goto/logger"
	"github.com/shanebarnes/scout/control"
	"github.com/shanebarnes/scout/execution"
	"github.com/shanebarnes/scout/global"
	"github.com/shanebarnes/scout/mission"
	"github.com/shanebarnes/scout/situation"
)

func sigHandler(ch *chan os.Signal) {
	sig := <-*ch
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
		syscall.SIGTERM)
	go sigHandler(&sigs)

	file, _ := os.OpenFile("scout.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	defer file.Close()

	logger.Init(log.Ldate|log.Ltime|log.Lmicroseconds, logger.Info, file)

	logger.PrintlnInfo("Starting scout", mission.GetVersion())

	orderFile := flag.String("order", "order.json", "file containing scouting operations order")
	/*reportFile := */ flag.String("report", "report.db", "file containing scouting report")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "version %s\n", mission.GetVersion())
		fmt.Fprintln(os.Stderr, "usage:")
		flag.PrintDefaults()
	}
	flag.Parse()

	db := global.GetDb()
	//db.Open(":memory:") // increase insert performance
	db.Open("scout.sqlite")
	//db.Exec("PRAGMA synchronous = NORMAL;")
	//db.Exec("PRAGMA journal_mode = WAL;")
	defer db.Close()

	order := loadOrder(orderFile)

	arr, err := situation.Parse(&order.Situation)
	if err != nil {
		logger.PrintlnError(err.Error())
		os.Exit(1)
	}

	tasks, err2 := execution.Parse(&order.Execution)
	if err2 != nil {
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
			ssh.Impl.CronSpec = order.Control.Frequency
			ssh.Impl.Wait = &wg
			ssh.Impl.WatchLimit = order.Control.Limit
			targets[i] = ssh
		} else {
			exec := new(situation.TargetExec)
			exec.New(i, arr[i], tasks)
			exec.Impl.Ch = &channels[i]
			exec.Impl.CronSpec = order.Control.Frequency
			exec.Impl.Wait = &wg
			exec.Impl.WatchLimit = order.Control.Limit
			targets[i] = exec
		}
	}

	for i := range targets {
		situation.Target.Watch(targets[i])
	}

	go control.HandleRequests(&order.Control)
	wg.Wait()
	logger.PrintlnInfo("Stopping scout", mission.GetVersion())
}

func loadOrder(fileName *string) Order {
	// defer file close
	file, _ := os.Open(*fileName)
	decoder := json.NewDecoder(file)
	order := Order{}
	err := decoder.Decode(&order)
	if err != nil {
		logger.PrintlnError(err.Error())
	}

	return order
}
