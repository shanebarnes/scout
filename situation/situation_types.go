package situation

import (
    "errors"

    "github.com/shanebarnes/goto/logger"
    "github.com/shanebarnes/scout/global"
)

type Credentials struct {
    Id   int    `json:"id"   sql:"id INTEGER NOT NULL PRIMARY KEY"`
    Name string `json:"name" sql:"name TEXT NOT NULL"`
    User string `json:"user" sql:"username TEXT NOT NULL"`
    Pass string `json:"pass" sql:"password TEXT NOT NULL"`
    Cert string `json:"cert" sql:"key TEXT NOT NULL"`
}
type CredentialsMap map[string]Credentials

type TargetGroup struct {
    Id       int    `json:"id"      sql:"id INTEGER NOT NULL PRIMARY KEY"`
    Name     string `json:"name"    sql:"name TEXT NOT NULL"`
    Addr   []string `json:"addr"    sql:"-"`
    CredId   int    `json:"cred_id" sql:"credential_id INTEGER NOT NULL"`
    Cred     string `json:"cred"    sql:"-"`
    Prot     string `json:"prot"    sql:"protocol TEXT NOT NULL"`
    Sys    []string `json:"sys"     sql:"system TEXT NOT NULL"`
}

type TargetDef struct {
    Id        int    `json:"id"       sql:"id INTEGER NOT NULL PRIMARY KEY"`
    Name      string `json:"name"     sql:"-"`
    GroupId   int    `json:"group_id" sql:"group_id INTEGER NOT NULL"`
    Addr      string `json:"addr"     sql:"address TEXT NOT NULL"`
    Cred      string `json:"cred"     sql:"-"`
    Prot      string `json:"prot"     sql:"-"`
    Sys     []string `json:"sys"      sql:"-"`
}
type TargetMap map[string]TargetGroup

type TargetEntry struct {
    Target TargetDef        `json:"target"      sql:"CREATE TABLE IF NOT EXISTS target_definitions"`
    Credentials Credentials `json:"credentials" sql:"CREATE TABLE IF NOT EXISTS target_credentials"`
}
type TargetArr []TargetEntry

type Situation struct {
    Targets []string `json:"targets"`
    Definitions TargetMap `json:"definitions"`
    Credentials CredentialsMap `json:"credentials"`
}

func Parse(situ *Situation) ([]TargetEntry, error) {
    size := 0
    ret := make([]TargetEntry, size)
    var err error = nil
    definitions := situ.Definitions
    credentials := situ.Credentials

    db := global.GetDb()

    i := 0
    for k, v := range credentials {
        v.Id = i
        v.Name = k
        credentials[k] = v
        i = i + 1
        db.CreateTable(&v)  // @todo only call once
        db.InsertInto(&v)
    }

    for i, id := range situ.Targets {
        var exists bool
        var group TargetGroup

        logger.PrintlnDebug("Parsing target group #", i)

        if group, exists = definitions[id]; exists {
            var cred Credentials
            var entry TargetEntry

            if cred, exists = credentials[group.Cred]; exists {
                entry.Credentials = cred
            } else {
                err = errors.New("Target '" + id + "' credentials '" + group.Cred + "' not found")
                break
            }

            group.Id = i
            group.CredId = cred.Id
            db.CreateTable(&group)  // @todo only call once
            db.InsertInto(&group)

            // todo: check for duplicate addreses? use unique attribute in sql?
            for _, addr := range group.Addr {
                entry.Target.Id = size
                entry.Target.Name = group.Name
                entry.Target.GroupId = group.Id
                entry.Target.Addr = addr
                entry.Target.Cred = group.Cred
                entry.Target.Prot = group.Prot
                entry.Target.Sys = group.Sys
                ret = append(ret, entry)
                size = size + 1
            }
        } else {
            err = errors.New("Target '" + id + "' is not found in definitions")
            break
        }
    }

    if size == 0 && err == nil {
        err = errors.New("No targets found")
    }

    if size > 0 {
        db.CreateTable(&ret[0].Target)
        for _, def := range ret {
            db.InsertInto(&def.Target)
        }
    }

    return ret, err
}
