package main

import (
    "bytes"
    "golang.org/x/crypto/ssh"
    "io/ioutil"
    "log"
    "strconv"
    "strings"
    "time"
)

type TargetSsh struct {
    impl TargetImpl
    client *ssh.Client
}

func (t *TargetSsh) New(conf TargetEntry, tasks TaskArr) error {
    return NewImpl(&t.impl, conf, tasks)
}

func (t *TargetSsh) Find() error {
    var err error = nil
    var config *ssh.ClientConfig = nil
    defer t.impl.wait.Done()

    _stdscr.MovePrintf(0, 0, "Finding target %s...\n", t.impl.conf.Target.Addr)
    _stdscr.Refresh()

    if len(t.impl.conf.Credentials.Pass) > 0 {
        config = &ssh.ClientConfig {
            User: t.impl.conf.Credentials.User,
            Auth: []ssh.AuthMethod {
                ssh.Password(t.impl.conf.Credentials.Pass),
            },
            Timeout: 5000 * time.Millisecond,
        }
    } else if len(t.impl.conf.Credentials.Cert) > 0 {
        file, _ := ioutil.ReadFile(t.impl.conf.Credentials.Cert)
        signer, err := ssh.ParsePrivateKey(file)
        if err != nil {
            log.Fatal(err)
        }
        config = &ssh.ClientConfig {
            User: t.impl.conf.Credentials.User,
            Auth: []ssh.AuthMethod {
                ssh.PublicKeys(signer),
            },
            Timeout: 5000 * time.Millisecond,
        }
    }

    t.client, err = ssh.Dial("tcp", t.impl.conf.Target.Addr + ":22", config)
    if err != nil {
        log.Fatal("Failed to dial: ", err)
    }

    return err
}

func (t *TargetSsh) Watch() error {
    var buffer bytes.Buffer
    var session *ssh.Session = nil
    var err error = nil
    defer t.impl.wait.Done()

    for {
        for i := range t.impl.task {
            if session == nil {
                if session, err = t.client.NewSession(); err == nil {
                    buffer.Truncate(0)
                    session.Stdout = &buffer
                }
            } else if err = session.Run(t.impl.task[i].Cmd); err == nil {
                value := strings.Trim(buffer.String(), " \r\n")
                select {
                    case *t.impl.ch <- strconv.Itoa(i):
                    default:
                }
                select {
                    case *t.impl.ch <- value:
                    default:
                }
            } else {
                session.Close()
                session = nil
            }
        }

        time.Sleep(time.Millisecond * 500)
    }

    return err
}

func (t *TargetSsh) Report() (*database, error) {
    return ReportImpl(&t.impl)
}

func (t *TargetSsh) GetImpl() *TargetImpl {
    return &t.impl
}

func (t *TargetSsh) IsLost() bool {
    //t.client.Close()
    return true
}
