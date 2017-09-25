package main

import (
    "github.com/shanebarnes/scout/control"
    "github.com/shanebarnes/scout/execution"
    "github.com/shanebarnes/scout/situation"
)

type Protocol struct {
   Protocol []string `json:"protocol"`
}

type Order struct {
    Mission string                `json:"mission"`
    Situation situation.Situation `json:"situation"`
    Execution execution.Execution `json:"execution"`
    Sustainment Protocol          `json:"sustainment"`
    Control control.Control       `json:"control"`
}

type Report struct {
    Timestamp uint64
    Value     uint64
    Rate      uint64
}
