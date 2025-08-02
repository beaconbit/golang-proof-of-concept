package scanner

import (
    "log"
)

func logPrefix(prefix string) *log.Logger {
    return log.New(log.Writer(), "scanner:"+prefix+" ", log.LstdFlags|log.Lshortfile)
}
