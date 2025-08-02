package authflow

import (
    "log"
)

func logPrefix(prefix string) *log.Logger {
    return log.New(log.Writer(), "authflow:"+prefix+" ", log.LstdFlags|log.Lshortfile)
}
