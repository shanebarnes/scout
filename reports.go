package main

import (
    "strconv"
)

type datapoint struct {
    x, y float64
}

type database struct {
    N uint64
    dp0, dpN datapoint
    max, min, rate, scale float64
    target, task, units string
}

func IsNum(s string) bool {
    _, err := strconv.ParseFloat(s, 64)
    return err == nil
}

func NewDataBase(target, task string, scale float64, units string) database {
    return database{
        N: 0,
        dp0: datapoint{x: 0, y: 0},
        dpN: datapoint{x: 0, y: 0},
        max: 0,
        min: 0,
        rate: 0,
        scale: scale,
        target: target,
        task: task,
        units: units}
}

func NewDataPoint(t uint64, y string) (datapoint, error) {
    var err error = nil
    dp := datapoint{x: 0, y:0}
    dp.x = float64(t)
    dp.y, err = strconv.ParseFloat(y, 64)
    return dp, err
}

func Evaluate(dp *datapoint, db *database) {
    db.N++
    if db.N == 1 {
        db.dp0 = *dp
        db.max = dp.y
        db.min = dp.y
        db.rate = 0
    } else {
        if dp.y > db.max { db.max = dp.y }
        if dp.y < db.min { db.min = dp.y }
        db.rate = CalcRate(db.dpN, *dp, db.scale)
    }
    db.dpN = *dp
}

func CalcRate(dp1, dp2 datapoint, scale float64) float64 {
    var rv float64 = 0.

    if dp2.x == dp1.x {

    } else {
        rv = 1000. * scale * (dp2.y - dp1.y) / (dp2.x - dp1.x)
    }

    return rv
}