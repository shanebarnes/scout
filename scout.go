package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "golang.org/x/crypto/ssh"
    "io/ioutil"
    "log"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"

    gc "github.com/rthornton128/goncurses"
)

type Subject struct {
    Name string `json:"name"`
    Addr string `json:"addr"`
    User string `json:"user"`
    Pass string `json:"pass"`
    Cert string `json:"cert"`
}

type Conf struct {
    Report string `json:"report"`
    Label string `json:"label"`
    Subjects []Subject `json:"subjects"`
}

type Report struct {
    Timestamp uint64
    Value     uint64
    Rate      uint64
}

var stdscr *gc.Window = nil

func main() {
    var err1 error
    stdscr, err1 = gc.Init()
    if err1 != nil {
        log.Fatal(err1)
    }
    defer gc.End()

    initializeWindow()
    conf := readConfig()

    clients := make([]*ssh.Client, len(conf.Subjects))

    for key, val := range conf.Subjects {
        clients[key] = connectSubject(&val)
    }

    var wg sync.WaitGroup
    channels := make([]chan uint64, len(clients))
    reports := make([]Report, len(clients))

    for i := range channels {
        channels[i] = make(chan uint64, 2)
        reports[i] = Report{Timestamp:0, Value:0, Rate:0}
    }

    for {
        wg.Add(len(channels))
        for key, val := range clients {
            go observeSubject(val, conf.Report, &wg, &channels[key])
        }
        wg.Wait()

        for i := range channels {
            timestamp := <-channels[i]
            value := <-channels[i]
            evaluateSubject(timestamp, value, &reports[i])
        }

        reportSubjects(conf, reports)

        time.Sleep(time.Millisecond * 500)
    }
}

func initializeWindow() {
    gc.Echo(false)
    gc.CBreak(true)
    gc.Cursor(0)
    gc.StartColor()

    gc.InitPair(gc.C_WHITE, gc.C_WHITE, gc.C_BLACK)
    gc.InitPair(gc.C_YELLOW, gc.C_YELLOW, gc.C_BLACK)
    gc.InitPair(gc.C_RED, gc.C_RED, gc.C_BLACK)
    gc.InitPair(gc.C_CYAN, gc.C_CYAN, gc.C_BLACK)
}

func readConfig() Conf {
    file, _ := os.Open("config.json")
    decoder := json.NewDecoder(file)
    config := Conf{}
    err := decoder.Decode(&config)
    if err != nil {
        fmt.Println("error: ", err)
    }
    return config
}

func connectSubject(subject *Subject) *ssh.Client {
    var client *ssh.Client = nil
    var config *ssh.ClientConfig = nil

    stdscr.MovePrintf(0, 0, "Connecting to %s...\n", subject.Addr)
    stdscr.Refresh()

    if len(subject.Pass) > 0 {
        config = &ssh.ClientConfig {
            User: subject.User,
            Auth: []ssh.AuthMethod {
                ssh.Password(subject.Pass),
            },
            Timeout: 2000 * time.Millisecond,
        }
    } else if len(subject.Cert) > 0 {
        file, _ := ioutil.ReadFile(subject.Cert)
        signer, err3 := ssh.ParsePrivateKey(file)
        if err3 != nil {
            log.Fatal(err3)
        }
        config = &ssh.ClientConfig {
            User: subject.User,
            Auth: []ssh.AuthMethod {
                ssh.PublicKeys(signer),
            },
            Timeout: 2000 * time.Millisecond,
        }
    }

    client, err := ssh.Dial("tcp", subject.Addr + ":22", config)
    if err != nil {
        log.Fatal("Failed to dial: ", err)
    }

    return client
}

func observeSubject(client *ssh.Client, cmd string, wg *sync.WaitGroup, data *chan uint64) {
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

    select {
        case *data <- uint64(time.Now().UnixNano()) / uint64(time.Millisecond):
        default:
    }

    if num, err := strconv.ParseUint(strings.Trim(b.String(), " \n"), 10, 64); err == nil {
        select {
            case *data <- num:
            default:
        }
    }
}

func evaluateSubject(timestamp uint64, value uint64, report *Report) {
    report.Rate = (value - report.Value) * 8 / (timestamp - report.Timestamp) / 1000
    report.Timestamp = timestamp
    report.Value = value
}

func reportSubjects(conf Conf, reports []Report) {
    var maxTicks int = 20
    var yScale uint64 = 50
    var total uint64 = 0
    stdscr.Move(0,0)

    for i := range reports {
        x := 0
        stdscr.MovePrintf(i, x, "%-15s", conf.Subjects[i].Addr)

        x += 15
        stdscr.MovePrintln(i, x, "[")

        x += 1
        ticks := reports[i].Rate / yScale

        for j := 0; j < maxTicks; j++ {
            if (j * 100 / maxTicks >= 66) {
                stdscr.ColorOn(gc.C_RED)
            } else if (j * 100 / maxTicks >= 33) {
                stdscr.ColorOn(gc.C_YELLOW)
            } else {
                stdscr.ColorOn(gc.C_CYAN)
            }

            if int(ticks) > j {
                stdscr.MovePrintln(i, x, "|")
            } else {
                stdscr.MovePrintln(i, x, " ")
            }

            x++
        }

        stdscr.ColorOff(gc.C_RED)
        stdscr.ColorOff(gc.C_YELLOW)
        stdscr.ColorOff(gc.C_CYAN)
        stdscr.MovePrintf(i, x, "%6d Mbps] %s\n", reports[i].Rate, conf.Report)

        total += reports[i].Rate
    }

    stdscr.MovePrintf(len(reports), 0,  "%-15s", "Total")
    stdscr.MovePrintf(len(reports), 36, "%6d Mbps\n", total)
    stdscr.Refresh()
}
