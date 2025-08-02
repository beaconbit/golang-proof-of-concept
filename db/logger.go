package db

import (
    "log"
)

func logPrefix(prefix string) *log.Logger {
    return log.New(log.Writer(), "db:"+prefix+" ", log.LstdFlags|log.Lshortfile)
}

