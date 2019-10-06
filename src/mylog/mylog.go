package mylog

import (
    "time"
    "os"
    "log"
)

const LOG_FILE_NAME = "machine.i.log"

func Log_writeln(line string) {
    f, err := os.OpenFile(LOG_FILE_NAME, os.O_APPEND | os.O_WRONLY, 0600)
    if err != nil {
        log.Fatal(err)
    }

    defer f.Close()

    if _, err := f.WriteString(time.Now().String() + ": " + line + "\n"); err != nil {
        log.Fatal(err)
    }
}

func Log_init() {
    _, err := os.Create(LOG_FILE_NAME)
    if err != nil {
        log.Fatal(err)
    }
}
