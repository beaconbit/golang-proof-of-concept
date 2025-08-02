package supervisor

import (
    "log"
)

func logPrefix(prefix string) *log.Logger {
    return log.New(log.Writer(), "supervisor:"+prefix+" ", log.LstdFlags|log.Lshortfile)
}

