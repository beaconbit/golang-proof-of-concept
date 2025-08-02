package scraper

import (
    "log"
)

func logPrefix(prefix string) *log.Logger {
    return log.New(log.Writer(), "scraper:"+prefix+" ", log.LstdFlags|log.Lshortfile)
}

