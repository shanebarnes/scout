package control

import (
    "database/sql"
    "os"
//    "strconv"

    _ "github.com/mattn/go-sqlite3"
    "github.com/shanebarnes/goto/logger"
    "github.com/shanebarnes/scout/situation"
)

type DbImpl struct {
    db *sql.DB
}

type Db interface {
    Open(dataSourceName string) error
    Close() error
    InitTables() error
    InsertReports(reports []Database) error
    InsertTargets(targets []situation.Target) error
}

func (d *DbImpl) Open(dataSourceName string) error {
    var err error = nil

    os.Remove(dataSourceName)

    if d.db, err = sql.Open("sqlite3", dataSourceName); err == nil {
    } else {
        logger.PrintlnError(err.Error())
    }

    return err
}

func (d *DbImpl) Close() error {
    return d.db.Close()
}

func (d *DbImpl) InsertReports(reports []Database) error {
    tx, err := d.db.Begin()

    if err != nil {
        //log.Fatal(err)
    }
    stmt, err := tx.Prepare("INSERT INTO reports (epoch_timestamp, target_id, task_id, x_val, y_val) VALUES (?, ?, ?, ?, ?)")
    if err != nil {
//        log.Fatal(err)
    }
    defer stmt.Close()

    for i := range reports {

//logger.PrintlnError("here we are" + strconv.FormatFloat(reports[i].DpN.Y, 'E', -1, 64))

        if _, err = stmt.Exec(0.0, 0, i, reports[i].DpN.X, reports[i].DpN.Y); err != nil {
            break
        }
    }

    if err == nil {
        tx.Commit()
    }

    return nil
}

func (d *DbImpl) InsertTargets(targets []situation.Target) error {
    tx, err := d.db.Begin()

    if err != nil {
        //log.Fatal(err)
    }
    stmt, err := tx.Prepare("INSERT INTO targets (id, name) VALUES (?, ?)")
    if err != nil {
//        log.Fatal(err)
    }
    defer stmt.Close()

    for i := range targets {
        impl := situation.Target.GetImpl(targets[i])
        if _, err = stmt.Exec(i, impl.Conf.Target.Name); err != nil {
            break
        }
    }

    if err == nil {
        tx.Commit()
    }

    return err
}

func (d *DbImpl) InitTables() error {
    var err error = nil

    sqlCmd := `
        CREATE TABLE IF NOT EXISTS targets (id INTEGER NOT NULL PRIMARY KEY, name TEXT NOT NULL);
        CREATE TABLE IF NOT EXISTS reports (epoch_timestamp REAL NOT NULL, target_id INTEGER NOT NULL, task_id INTEGER NOT NULL, x_val REAL NOT NULL, y_val REAL NOT NULL);
    `

    if _, err = d.db.Exec(sqlCmd); err != nil {
        logger.PrintlnError(err.Error() + ":" + sqlCmd)
    }

    return err
}
