package supervisor

import (
    "sync"
    "time"
    "graphite/publisher/worker"
)

type RestartStrategy string

const (
    Backoff RestartStrategy = "backoff"
)

type BackoffEntry struct {
    Delay     time.Duration
    MaxTries  int
}

type SupervisorConfig struct {
    Strategy RestartStrategy
    Backoff  []BackoffEntry
    Timeout  time.Duration // heartbeat timeout
}

type managedWorker struct {
    id       int
    instance worker.Worker
    hb          <-chan struct{}
    done        <-chan struct{}
    notify      <-chan struct{}
    err         <-chan error
    lastHB      time.Time
    lastRestart time.Time
    restarts    int
}

type Supervisor struct {
    config     SupervisorConfig
    mu         sync.Mutex
    workers    map[int]*managedWorker
    nextID     int
    restartQ   chan restartMessage
}

type restartMessage struct {
    worker      worker.Worker
    id          int
    restarts    int
    lastRestart time.Time
}

func NewSupervisor(config SupervisorConfig) *Supervisor {
    return &Supervisor{
        config:   config,
        workers:  make(map[int]*managedWorker),
        restartQ: make(chan restartMessage, 100),
    }
}

func (s *Supervisor) Supervise(w worker.Worker) {

    logger := logPrefix("Supervise")

    s.mu.Lock()
    // get next id
    id := s.nextID
    s.nextID++
    s.mu.Unlock()

    logger.Println("queued worker for start: ", id, " nextID: ", s.nextID)

    msg := restartMessage {
        worker:      w,
        id:          id,
        restarts:    0,
        lastRestart: time.Now(),
    }

    s.restartQ <- msg
}

func (s *Supervisor) Run() {
    logger := logPrefix("Run")
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            logger.Println("tick")
            s.checkWorkers()
            s.handleRestarts()
        }
    }
}

func (s *Supervisor) checkWorkers() {
    logger := logPrefix("checkWorkers")
    logger.Println("checking workers")
    s.mu.Lock()
    defer s.mu.Unlock()

    for id, w := range s.workers {
        logger.Printf("Worker ID: %d, Pointer: %p\n", id, w)
    }

    now := time.Now()
    for id, mw := range s.workers {
        select {
        case <-mw.done:
            logger.Printf("[Supervisor] Worker %d done", id)
            delete(s.workers, id)
        case <-mw.notify:
            logger.Printf("[Supervisor] Worker %d sent notify", id)
        case err := <-mw.err:
            logger.Printf("[Supervisor] Worker %d error: %v", id, err)
            msg := restartMessage {
                worker:      mw.instance,
                id:          mw.id,
                restarts:    mw.restarts,
                lastRestart: mw.lastRestart,
            }
            s.restartQ <- msg
        case <-mw.hb:
            logger.Printf("heartbeat detected")
            mw.lastHB = now
        default:
            if now.Sub(mw.lastHB) > s.config.Timeout {
		logger.Println(
		    "[Supervisor] Worker ", 
		    id, 
		    " timed out. Last heartbeat ", 
		    mw.lastHB,
		    " time since last heartbeat ",
		    now.Sub(mw.lastHB),
		    " config timeout ",
		    s.config.Timeout,
		)
                mw.instance.Stop()
                msg := restartMessage {
                    worker:      mw.instance,
                    id:          mw.id,
                    restarts:    mw.restarts,
                    lastRestart: mw.lastRestart,
                }
                s.restartQ <- msg
            }
        }
    }
}

func (s *Supervisor) handleRestarts() {
    logger := logPrefix("handleRestarts")
    for {
        select {
        case msg := <-s.restartQ:
            logger.Printf("calling restart worker sequence. worker %d. restarts so far %d", msg.id, msg.restarts)
            delay := s.getBackoffDelay(msg.lastRestart, msg.restarts)
            if delay < 0 {
                go s.restartWorker(msg.worker, msg.id, msg.restarts)
            } else {
		logger.Printf("putting it back on the queue motherfucker")
                // put it back on the queue
                s.restartQ <- msg
            }
        default:
            return
        }
    }
}


func (s *Supervisor) restartWorker(w worker.Worker, id int, restarts int) {

    logger := logPrefix("restartWorker")

    // retry logic

    logger.Println("supervisor added and will start worker: ", id)

    // Create and supervise a fresh worker instance
    s.mu.Lock()
    mw := &managedWorker{
        id:          id,
        instance:    w,
        hb:          w.Heartbeat(),
        done:        w.Done(),
        notify:      w.Notify(),
        err:         w.Err(),
        lastHB:      time.Now(),
        lastRestart: time.Now(),
        restarts:    restarts + 1,
    }

    s.workers[id] = mw
    s.mu.Unlock()

    w.Start()
}

func (s *Supervisor) getBackoffDelay(lastRestart time.Time, attempt int) time.Duration {
    logger := logPrefix("getBackoffDelay")
    // TODO add logic to reset if last restart was ages ago
    total := 0
    for _, entry := range s.config.Backoff {
        if attempt < total+entry.MaxTries {
            logger.Println("returning delay of ", entry.Delay)
            return -1 // TODO placeholder
            return entry.Delay
        }
        total += entry.MaxTries
    }
    logger.Println("returning delay of ", -1)
    return -1 // exhausted retries
}

