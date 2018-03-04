package control

import (
	"os"
)

type Ops struct {
	Op string `json:"op"`
}

type Control struct {
	Root      string         `json:"root"`
	Duration  string         `json:"duration"`
	Frequency string         `json:"frequency"`
	Limit     int64          `json:"limit"`
	Reports   map[string]Ops `json:"reports"`
}

func Parse(ctl *Control) error {
	_, err := os.Stat(ctl.Root)

	return err
}
