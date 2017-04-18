package main

import (
    "bytes"
    "golang.org/x/crypto/ssh"
    "io/ioutil"
    "log"
    "sync"
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

    if len(t.impl.conf.Credentials.Pass) > 0 {
        config = &ssh.ClientConfig {
            User: t.impl.conf.Credentials.User,
            Auth: []ssh.AuthMethod {
                ssh.Password(t.impl.conf.Credentials.Pass),
            },
// This timeout will block the update thread
            Timeout: 10000 * time.Millisecond,
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
// This timeout will block the update thread
            Timeout: 10000 * time.Millisecond,
        }
    }

    t.client, err = ssh.Dial("tcp", t.impl.conf.Target.Addr + ":22", config)

    return err
}

func (t *TargetSsh) Watch() error {
    var buffer bytes.Buffer
    var session *ssh.Session = nil
    var err error = nil
    defer t.impl.wait.Done()

    for {
        start := time.Now()
        if t.client == nil {
            var wg sync.WaitGroup
            wg.Add(1)
            t.impl.wait = &wg
            target.Find(t)
        } else if session, err = t.client.NewSession(); err == nil {
            buffer.Truncate(0)
            session.Stdout = &buffer
            if err = session.Run(t.impl.cmds); err == nil {

                RecordImpl(&t.impl, buffer.Bytes(), time.Since(start))
            }

            session.Close()
            session = nil
        } else {
            t.client = nil
        }

        t.impl.nextWatch = t.impl.nextWatch.Add(1000 * time.Millisecond)
        for time.Since(t.impl.nextWatch).Nanoseconds() / int64(time.Millisecond) >= 1000 {
            t.impl.nextWatch = t.impl.nextWatch.Add(1000 * time.Millisecond)
        }

        time.Sleep(t.impl.nextWatch.Sub(time.Now()))
    }

    return err
}

func (t *TargetSsh) Report() (*database, error) {
    return ReportImpl(&t.impl)
}

func (t *TargetSsh) GetImpl() *TargetImpl {
    return &t.impl
}
