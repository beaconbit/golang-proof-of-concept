package scanner 

import (
    "time"
    "errors"
    "gorm.io/gorm"
    "graphite/publisher/db"
    "graphite/publisher/utils"
)

var ping = struct{}{}

type Worker interface {
    Start()
    Stop()
    Heartbeat() <-chan struct{}
    Done()      <-chan struct{}
    Notify()    <-chan struct{}
    Err()       <-chan error
}

type WorkerCommunication struct {
    heartbeat chan struct{}
    done chan struct{}
    notify chan struct{}
    err chan error
}

type Scanner struct {
    communication     WorkerCommunication
    running 	      bool
    jobCh             chan utils.Job[*gorm.DB]
    scanTicker        *time.Ticker
    heartbeatTicker   *time.Ticker
    updateQ           chan []db.Address
}

func NewScanner(jobCh chan utils.Job[*gorm.DB]) *Scanner {
    wc := WorkerCommunication {
        heartbeat:  make(chan struct{}, 1),
        done:  make(chan struct{}, 1),
        notify:  make(chan struct{}, 1),
        err:  make(chan error, 1),
    }
    return &Scanner{
        communication: wc,
        running: false,
	jobCh: jobCh,
	scanTicker: time.NewTicker(120 * time.Second),
	heartbeatTicker: time.NewTicker(5 * time.Second),
	updateQ: make(chan []db.Address, 10),
    }
}


func (s *Scanner) Start() {
    logger := logPrefix("Start") 
    logger.Println("Scanner started")
    s.running = true
    go func() {
        // catch panics and write to err channel before exiting
        defer func() {
            if r := recover(); r != nil {
                var err error
                switch v := r.(type) {
                case error:
                    err = v
                default:
                    err = errors.New("something went wrong. scanner died")
                }

                select {
                case s.communication.err <- err:
                default:
                }
            }
            // if no error was found then write to done channel
            s.communication.done <- ping
        }()

        for s.running{
            // heart beat
	    logger.Println("heartbeat and work loop")
            select {
            case s.communication.heartbeat <-ping:
		logger.Println("sending heartbeat")
            default:
		logger.Println("skip sending heartbeat, last one not read")
                // just continue if previous ping hasn't been read
            }
            // work
            select {
            case <-s.scanTicker.C:
		// dispatch scan
		go s.scanNetwork()

		// check update queue
            case scanResult := <-s.updateQ:
                go s.handleUpdate(scanResult)
		// pass new updates to network topology writer
	    case <-s.heartbeatTicker.C:
            	s.communication.heartbeat <-ping
            }
        }
    }()
}

func (s *Scanner) scanNetwork() {
    logger := logPrefix("scanNetwork")
    logger.Println("scanNetwork")
    // dummy discovery

    var all []db.Address

    // normalized mac address
    dummyAddress := db.Address{Mac: "74fe486c1fb4", IP: "192.168.1.100"}
    dummyAddress2 := db.Address{Mac: "74fe486c1f20", IP: "192.168.1.195"}
    result := []db.Address{dummyAddress, dummyAddress2}

    // send results
    for _, d := range result {
        all = append(all, db.Address{Mac: d.Mac, IP: d.IP})
    }
    s.updateQ <- all
}

func (s *Scanner) handleUpdate(scanResult []db.Address) {
    logger := logPrefix("handleUpdate")
    logger.Println("handleUpdate")
    // bulk update 
    job, resultCh, errCh := db.BulkUpsertAddresses(scanResult)
    s.jobCh <- job
    select {
    case err := <-errCh:
        logger.Println("error: ", err)
    case _ = <-resultCh:
        logger.Println("bulk upsert succeeded")
    }
}

func (s *Scanner) Stop() {
    s.scanTicker.Stop()
    s.running = false
}

func (s *Scanner) Heartbeat() <-chan struct{} {
    s.communication.heartbeat = make(chan struct{}, 1)  // or unbuffered
    return s.communication.heartbeat
}
func (s *Scanner) Done() <-chan struct{} {
        s.communication.done = make(chan struct{}, 1)
    return s.communication.done
}
func (s *Scanner) Notify() <-chan struct{} {
        s.communication.notify = make(chan struct{}, 1)
    return s.communication.notify
}
func (s *Scanner) Err() <-chan error {
        s.communication.err =  make(chan error, 1)
    return s.communication.err
}

