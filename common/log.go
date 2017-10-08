package common

import (
    "log"
    "os"
    "sync"
)

var instance *log.Logger = nil
var once sync.Once

func getLogInstance() *log.Logger {
    once.Do(func() {
        flags := log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile
        file, err := os.OpenFile("scout.log", os.O_APPEND | os.O_CREATE | os.O_RDWR, 0644)

        if err == nil {
            instance = log.New(file, "", flags)
        } else {
            instance = log.New(os.Stderr, "", flags)
        }
        //defer file.Close()
    })
    return instance
}

func LogDebug(msg string) {
    getLogInstance().Println("[DEBUG] " + msg)
}

func LogError(msg string) {
    getLogInstance().Println("[ERROR] " + msg)
}

func LogInfo(msg string) {
    getLogInstance().Println("[INFO] " + msg)
}
