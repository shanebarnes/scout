package main

import (
    "github.com/shanebarnes/scout/execution"
    "github.com/shanebarnes/scout/situation"
)

type Report1 struct {
    Op string `json:"op"`
}
type ReportMap map[string]Report1

type Protocol struct {
   Protocol []string `json:"protocol"`
}

type Control1 struct {
    Frequency string `json:"frequency"`
    Duration string `json:"duration"`
    Reports ReportMap `json:"reports"`
}

type Order struct {
    Mission string `json:"mission"`
    Situation situation.Situation `json:"situation"`
    Execution execution.Execution1 `json:"execution"`
    Sustainment Protocol `json:"sustainment"`
    Control Control1 `json:"control"`
}

type Report struct {
    Timestamp uint64
    Value     uint64
    Rate      uint64
}
