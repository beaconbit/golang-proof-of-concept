package main

import (
    "log"
    "os"
    "gorm.io/gorm"
    "graphite/publisher/db"
    "graphite/publisher/scanner"
    "graphite/publisher/utils"
    "graphite/publisher/cookiefinder"
    //"github.com/nats-io/nats.go"
)

// Connect to JetStream
//nc, _ := nats.Connect("nats://nats:4222")
//js, _ := nc.JetStream()

// Example message
//js.Publish("events.device", []byte(`{"mac":"01:23","status":"ok"}`))

func logPrefix(prefix string) *log.Logger {
    return log.New(log.Writer(), "main:"+prefix+" ", log.LstdFlags|log.Lshortfile)
	
}

func main() {
    if false{
        r := make(chan db.Address, 100)
        scanner.DoScan("eno1", r)
    } else {
        log.SetOutput(os.Stdout)
        log.SetFlags(log.LstdFlags | log.Lshortfile) // Timestamp + file:line
        logger := logPrefix("main")

        logger.Println("starting app")


        logger.Println("Running migrations...")
        db.RunMigrations()

        logger.Println("Connecting to database...")
        conn := db.NewSession()

        // creating all interprocess channels
        // think about wrapping this in communication interface structs 
        // think about testing harnesses that could stub these
        jobCh := make(chan utils.Job[*gorm.DB], 10)

        run(conn, jobCh)
    }
}
func run(conn *gorm.DB, jobCh chan utils.Job[*gorm.DB]) {
    logger := logPrefix("run")
    logger.Println("beginning run loop")

    databaseJobQ := utils.NewJobExecutor(conn, jobCh)
    go databaseJobQ.Run()

    scanner := scanner.NewScanner(20, jobCh)
    go scanner.Run()

    cookiefinder := cookiefinder.NewCookieFinder(10, jobCh)
    go cookiefinder.Run()

    logger.Println("running ...")
    select {} // block forever
}
