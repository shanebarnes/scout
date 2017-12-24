package global

import (
    "database/sql"
    "os"
    "reflect"
    "strconv"
    "strings"
    "sync"

    "github.com/jmoiron/sqlx"
    _ "github.com/mattn/go-sqlite3"
    "github.com/shanebarnes/goto/logger"
)

const dbTag = "sql"
const ignoreField = "-"

type DbImpl struct {
    db *sqlx.DB
}

var dbInstance *DbImpl = nil
var dbOnce      sync.Once

type Db interface {
    Open(dataSourceName string) error
    Close() error
    CreateTable(f interface{}, table string) error
    Exec(query string) error
    GetFields(f interface{}) string
    Insert(query string) error
    InsertInto(f interface{}, table string) error
}

func GetDb() *DbImpl {
    dbOnce.Do(func() {
        dbInstance = new(DbImpl)
    })
    return dbInstance
}

func (d *DbImpl) Open(dataSourceName string) error {
    var err error = nil

    os.Remove(dataSourceName)

    if d.db, err = sqlx.Open("sqlite3", dataSourceName); err == nil {
        logger.PrintlnDebug("Opened database:", dataSourceName)
    } else {
        logger.PrintlnError(err.Error())
    }

    return err
}

func (d *DbImpl) Close() error {
    err := d.db.Close()
    logger.PrintlnDebug("Closed database")
    return err
}

func (d *DbImpl) CreateTable(f interface{}) error {
    var err error = nil
    table := reflect.TypeOf(f).Elem().Name()
    v := reflect.ValueOf(f).Elem()

    if v.NumField() > 0 {
        query := "CREATE TABLE IF NOT EXISTS " + table + " ("
        for i := 0; i < v.NumField(); i++ {
            f := v.Field(i)

            if f.CanInterface() {
                t := v.Type().Field(i)
                tag := t.Tag.Get(dbTag)

                if tag != ignoreField {
                    if i > 0 {
                        query = query + ", "
                    }

                    query = query + tag
                }
            }
        }

        query = query + ");"
        logger.PrintlnDebug("Create query:", query)
        err = d.Exec(query)
    }

    return err
}

func (d *DbImpl) Exec(query string) error {
    var err error = nil

    if _, err = d.db.Exec(query); err == nil {
        logger.PrintlnDebug("Executed query:", query)
    } else {
        logger.PrintlnError("Failed to execute query '" + query + "':", err.Error())
    }

    return err
}

func (d *DbImpl) GetFields(f interface{}) string {
    fields := ""

    v := reflect.ValueOf(f).Elem()

    for i := 0; i < v.NumField(); i++ {
        field := v.Field(i)

        if field.CanInterface() {
            t := v.Type().Field(i)
            tag := t.Tag.Get(dbTag)

            if tag != ignoreField {
                if i > 0 {
                    fields = fields + ", "
                }

                if n := strings.Index(tag, " "); n > 0 {
                    fields = fields + tag[:n]
                } else {
                    fields = fields + tag
                }
            }
        }
    }

    return fields
}

func (d *DbImpl) Insert(query []string) error {
    var err error = nil
    var tx *sql.Tx = nil

    if tx, err = d.db.Begin(); err == nil {
        for _, q := range query {
            var stmt *sql.Stmt = nil
            stmt, err = tx.Prepare(q)
            if err != nil {
                break
            }
            defer stmt.Close()

            _, err = stmt.Exec()

            if err == nil {
                tx.Commit()
            } else {
                tx.Rollback()
            }
        }
    }

    if err != nil {
        logger.PrintlnError("Transaction failed to start:", err)
    }

    return err
}

func (d *DbImpl) InsertInto(f interface{}) error {
    var err error = nil

    table := reflect.TypeOf(f).Elem().Name()
    v := reflect.ValueOf(f).Elem()

    if v.NumField() > 0 {
        insert := "INSERT INTO " + table + " ("
        values := "VALUES ("

        for i := 0; i < v.NumField(); i++ {
            field := v.Field(i)
            t := v.Type().Field(i)

            if field.CanInterface() {
                tag := t.Tag.Get(dbTag)

                if tag != ignoreField {
                    if i > 0 {
                        insert = insert + ", "
                        values = values + ", "
                    }

                    if n := strings.Index(tag, " "); n > 0 {
                        insert = insert + tag[:n]
                    } else {
                        insert = insert + tag
                    }

                    val := reflect.ValueOf(field.Interface())

                    switch val.Kind() {
                    case reflect.Float32, reflect.Float64:
                        values = values + strconv.FormatFloat(val.Float(), 'f', 6, 64)
                    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
                        values = values + strconv.FormatInt(val.Int(), 10)
                    case reflect.String:
                        values = values + "\"" + val.String() + "\""
                    default:
                        values = values + "\"" + val.String() + "\""
                    }
                }
            }
        }

        query := insert + ") " + values + ");"
        logger.PrintlnDebug("Insert query:", query)
        err = d.Insert([]string{query})
    }

    return err
}

func (d *DbImpl) Select(query string, result interface{}) error {
    var err error = nil

    if t := reflect.TypeOf(result); t.Kind() == reflect.Ptr {
        // @todo Eliminate need to parse package path
        //typeStr := t.String()
        //n := strings.LastIndexByte(typeStr, '.')

        //if n != -1 && n < len(typeStr) {
            //table := typeStr[n+1:len(typeStr)]
            //query := "SELECT * FROM " + table + ";"
            logger.PrintlnDebug("Insert query:", query)

            var rows *sql.Rows = nil
            if rows, err = d.db.Query(query); err == nil {
                if err = sqlx.StructScan(rows, result); err != nil {
                    logger.PrintlnError("Scan failed:", err)
                }
                rows.Close()
            } else {
                logger.PrintlnError("Query failed:", err)
            }
        //}
    }

    return err
}
