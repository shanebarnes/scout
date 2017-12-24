package control

import (
    "strconv"

    "github.com/shanebarnes/scout/execution"
)

type datapoint struct {
    X float64 `json:"x" sql:"x REAL NOT NULL"`
    Y float64 `json:"y" sql:"y REAL NOT NULL"`
    d string  `json:"d" sql:"d TEXT NOT NULL"`
}

type Database struct {
    N uint64 `json:"N"`
    dp0 datapoint `json:"dp0"`
    DpN datapoint `json:"dpN"`
    Diff float64 `json:"diff"`
    Max float64 `json:"max"`
    Min float64 `json:"min"`
    Rate float64 `json:"rate"`
    Target string `json:"target"`
    Location string `json:"location"`
    Task string `json:"task"`
    Reports []execution.TaskReport `json:"reports"`
}

func IsNum(s string) bool {
    _, err := strconv.ParseFloat(s, 64)
    return err == nil
}

func NewDataBase(target, location, task string, reports []execution.TaskReport/*scale []float64, units []string, widget string*/) Database {
    // @todo Return reference
    return Database{
        N: 0,
        dp0: datapoint{X: 0, Y: 0},
        DpN: datapoint{X: 0, Y: 0},
        Max: 0,
        Min: 0,
        Diff: 0,
        Rate: 0,
        Target: target,
        Location: location,
        Task: task,
        Reports: reports}
}

func NewDataPoint(taskId int, t uint64, d, y string) (datapoint, error) {
    var err error = nil
    dp := datapoint{X: 0, Y:0, d:d}
    dp.X = float64(t)
    dp.Y, err = strconv.ParseFloat(y, 64)
    return dp, err
}

func Evaluate(dp *datapoint, db *Database) {
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
