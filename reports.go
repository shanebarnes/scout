package main

import (
    "strconv"
)

type datapoint struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
    d string `json:"d"`
}

type database struct {
    N uint64 `json:"N"`
    dp0 datapoint `json:"dp0"`
    DpN datapoint `json:"dpN"`
    Diff float64 `json:"diff"`
    Max float64 `json:"max"`
    Min float64 `json:"min"`
    Rate float64 `json:"rate"`
    Scale []float64 `json:"scale"`
    Target string `json:"target"`
    Task string `json:"task"`
    Units []string `json:"units"`
}

func IsNum(s string) bool {
    _, err := strconv.ParseFloat(s, 64)
    return err == nil
}

func NewDataBase(target, task string, scale []float64, units []string) database {
    return database{
        N: 0,
        dp0: datapoint{X: 0, Y: 0},
        DpN: datapoint{X: 0, Y: 0},
        Max: 0,
        Min: 0,
        Diff: 0,
        Rate: 0,
        Scale: scale,
        Target: target,
        Task: task,
        Units: units}
}

func NewDataPoint(t uint64, d, y string) (datapoint, error) {
    var err error = nil
    dp := datapoint{X: 0, Y:0, d:d}
    dp.X = float64(t)
    dp.Y, err = strconv.ParseFloat(y, 64)
    return dp, err
}

func Evaluate(dp *datapoint, db *database) {
    db.N++
    if db.N == 1 {
        db.dp0 = *dp
        db.Diff = dp.Y
        db.Max = dp.Y
        db.Min = dp.Y
        db.Rate = 0
    } else {
        db.Diff = dp.Y - db.DpN.Y
        if dp.Y > db.Max { db.Max = dp.Y }
        if dp.Y < db.Min { db.Min = dp.Y }
        db.Rate = CalcRate(db.DpN, *dp)
    }
    db.DpN = *dp
}

func CalcRate(dp1, dp2 datapoint) float64 {
    var rv float64 = 0.

    if dp2.X == dp1.X {

    } else {
        rv = 1000. * (dp2.Y - dp1.Y) / (dp2.X - dp1.X)
    }

    return rv
}
