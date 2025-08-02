package publisher

import (
    "log"
)

func logPrefix(prefix string) *log.Logger {
    return log.New(log.Writer(), "publisher:"+prefix+" ", log.LstdFlags|log.Lshortfile)
}

