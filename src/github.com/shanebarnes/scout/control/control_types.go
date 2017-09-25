package control

import (
    "os"
)

type Report struct {
    Op string `json:"op"`
}

type Control struct {
    Root      string            `json:"root"`
    Frequency string            `json:"frequency"`
    Duration  string            `json:"duration"`
    Reports   map[string]Report `json:"reports"`
}

func Parse(ctl *Control) error {
    _, err := os.Stat(ctl.Root)

    return err
}
