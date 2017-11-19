package situation

import (
    "bytes"
    "golang.org/x/crypto/ssh"
    "io/ioutil"
    "log"
    "sync"
    "time"

    "github.com/shanebarnes/goto/logger"
    "github.com/shanebarnes/scout/execution"
)

type TargetSsh struct {
    Impl TargetImpl
    client *ssh.Client
}

func (t *TargetSsh) New(id int, conf TargetEntry, tasks execution.TaskArray) error {
    return NewImpl(&t.Impl, id, conf, tasks)
}

func (t *TargetSsh) Find() error {
    var err error = nil
    var config *ssh.ClientConfig = nil
    defer t.Impl.Wait.Done()

    logger.PrintlnDebug("Attempting to find SSH target " + t.Impl.Conf.Target.Addr)

    if len(t.Impl.Conf.Credentials.Pass) > 0 {
        config = &ssh.ClientConfig {
            User: t.Impl.Conf.Credentials.User,
            Auth: []ssh.AuthMethod {
                ssh.Password(t.Impl.Conf.Credentials.Pass),
            },
            HostKeyCallback: ssh.InsecureIgnoreHostKey(),
// This timeout will block the update thread
            Timeout: 10000 * time.Millisecond,
        }
    } else if len(t.Impl.Conf.Credentials.Cert) > 0 {
        file, _ := ioutil.ReadFile(t.Impl.Conf.Credentials.Cert)
        signer, err := ssh.ParsePrivateKey(file)
        if err != nil {
            log.Fatal(err)
        }
        config = &ssh.ClientConfig {
            User: t.Impl.Conf.Credentials.User,
            Auth: []ssh.AuthMethod {
                ssh.PublicKeys(signer),
            },
            HostKeyCallback: ssh.InsecureIgnoreHostKey(),
// This timeout will block the update thread
            Timeout: 10000 * time.Millisecond,
        }
    }

    t.client, err = ssh.Dial("tcp", t.Impl.Conf.Target.Addr, config)

    if err != nil {
        logger.PrintlnError("Cannot find target " + t.Impl.Conf.Target.Addr + ": " + err.Error())
    }

    return err
}

func (t *TargetSsh) Watch() error {
    var buffer bytes.Buffer
    var session *ssh.Session = nil
    var err error = nil
    defer t.Impl.Wait.Done()

    for {
        start := time.Now()
        if t.client == nil {
            var wg sync.WaitGroup
            wg.Add(1)
            t.Impl.Wait = &wg
            Target.Find(t)
        } else if session, err = t.client.NewSession(); err == nil {
            buffer.Truncate(0)
            session.Stdout = &buffer
            if err = session.Run(t.Impl.cmds); err == nil {

                RecordImpl(&t.Impl, buffer.Bytes(), time.Since(start))
            }

            session.Close()
            session = nil
        } else {
            t.client = nil
        }

        t.Impl.NextWatch = t.Impl.NextWatch.Add(1000 * time.Millisecond)
        for time.Since(t.Impl.NextWatch).Nanoseconds() / int64(time.Millisecond) >= 1000 {
            t.Impl.NextWatch = t.Impl.NextWatch.Add(1000 * time.Millisecond)
        }

        time.Sleep(t.Impl.NextWatch.Sub(time.Now()))
    }

    return err
}

func (t *TargetSsh) Report() (*TargetObs, error) {
    return ReportImpl(&t.Impl)
}

func (t *TargetSsh) GetImpl() *TargetImpl {
    return &t.Impl
}
