package situation

import (
    "errors"
)

type Credentials struct {
    User string `json:"user"`
    Pass string `json:"pass"`
    Cert string `json:"cert"`
}
type CredentialsMap map[string]Credentials

type TargetGroup struct {
    Name string `json:"name"`
    Addr []string `json:"addr"`
    Cred string `json:"cred"`
    Prot string `json:"prot"`
    Sys []string `json:"sys"`
}

type TargetDef struct {
    Name string `json:"name"`
    Addr string `json:"addr"`
    Cred string `json:"cred"`
    Prot string `json:"prot"`
    Sys []string `json:"sys"`
}
type TargetMap map[string]TargetGroup

type TargetEntry struct {
    Target TargetDef
    Credentials Credentials
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

    for _, id := range situ.Targets {
        var exists bool
        var group TargetGroup

        if group, exists = definitions[id]; exists {
            var cred Credentials
            var entry TargetEntry

            if cred, exists = credentials[group.Cred]; exists {
                entry.Credentials = cred
            } else {
                err = errors.New("Target '" + id + "' credentials '" + group.Cred + "' not found")
                break
            }

            // todo: check for duplicate addreses?
            for _, addr := range group.Addr {
                entry.Target.Name = group.Name
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

    return ret, err
}
