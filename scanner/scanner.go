package scanner 

import (
    "time"
    "gorm.io/gorm"
    "graphite/publisher/db"
    "graphite/publisher/utils"
)

type Scanner struct {
    running 	      bool
    jobCh             chan utils.Job[*gorm.DB]
    scanTicker        *time.Ticker
    updateQ           chan []db.Address
}

func NewScanner(period time.Duration, jobCh chan utils.Job[*gorm.DB]) *Scanner {
    return &Scanner{
        running: false,
	jobCh: jobCh,
	scanTicker: time.NewTicker(period * time.Second),
	updateQ: make(chan []db.Address, 10),
    }
}


func (s *Scanner) Run() {
    logger := logPrefix("Start") 
    logger.Println("Scanner started")
    s.running = true
    for s.running {
	select {
	case <-s.scanTicker.C:
	    go s.scanNetwork()
	case scanResult := <-s.updateQ:
//	    logger.Println("read item off s.updateQ and call s.handleUpdate(scanResult)")
	    go s.handleUpdate(scanResult)
        }
    }
}

func (s *Scanner) Stop() {
    s.running = false
}

func (s *Scanner) scanNetwork() {
    logger := logPrefix("scanNetwork")
    logger.Println("scanning network")

    var all []db.Address

    responseCh := make(chan db.Address, 100)
    go DoScan("eno1", responseCh)

    // batch process into a single message []db.Address
    for d := range responseCh {
//	logger.Println(d)
        all = append(all, db.Address{Mac: d.Mac, IP: d.IP})
    }
 //   logger.Println("closing response channel")
  //  logger.Println("placing all active addresses on s.updateQ")
    s.updateQ <- all
}

func (s *Scanner) handleUpdate(scanResult []db.Address) {
    logger := logPrefix("handleUpdate")
   // logger.Println("processing scan result")
    // bulk update 
    job, resultCh, errCh := db.BulkUpsertAddresses(scanResult)
    s.jobCh <- job
   // logger.Println("sent job to bulk upsert queue")
    select {
    case _ = <-errCh:
        // logger.Println("error: ", err)
    case _ = <-resultCh:
        logger.Println("bulk upsert succeeded")
    }
}
