package main

import (
    "log"
    "os"
    "time"
    "gorm.io/gorm"
    "graphite/publisher/db"
    "graphite/publisher/scanner"
    "graphite/publisher/supervisor"
    "graphite/publisher/utils"
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
    if true{
        scanner.DoScan()
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

    // testing supervisor implementation
    cfg := supervisor.SupervisorConfig{
        Strategy: supervisor.Backoff,
        Timeout:  25 * time.Second,
        Backoff: []supervisor.BackoffEntry{
            {Delay: 2 * time.Second, MaxTries: 3},
            {Delay: 5 * time.Second, MaxTries: 2},
        },
    }
    runtime := supervisor.NewSupervisor(cfg)

    databaseJobQ := utils.NewJobExecutor(conn, jobCh)
    go databaseJobQ.Run()

    scanner := scanner.NewScanner(jobCh)

    runtime.Supervise(scanner)
    go runtime.Run()
    // testing supervisor implementation

    logger.Println("running ...")
    select {} // block forever
}
