package worker

import (
    "log"
)

func logPrefix(prefix string) *log.Logger {
    return log.New(log.Writer(), "worker:"+prefix+" ", log.LstdFlags|log.Lshortfile)
}

