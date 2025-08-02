package worker 

import (
    "math/rand"
    "time"
    "fmt"
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

type ExampleWorker struct {
    communication WorkerCommunication
    running bool
}
func NewExampleWorker() *ExampleWorker {
    wc := WorkerCommunication {
        heartbeat:  make(chan struct{}, 1),
        done:  make(chan struct{}, 1),
        notify:  make(chan struct{}, 1),
        err:  make(chan error, 1),
    }
    return &ExampleWorker{
        communication: wc,
        running: false,
    }
}


func (w *ExampleWorker) Start() {
    logger := logPrefix("Start") 
    w.running = true
    go func() {
        // catch panics and write to err channel before exiting
        defer func() {
            if r := recover(); r != nil {
                var err error
                switch v := r.(type) {
                case error:
                    err = v
                default:
                    err = fmt.Errorf("panic: %v", v)
                }

                select {
                case w.communication.err <- err:
                default:
                }
            }
            // if no error was found then write to done channel
            w.communication.done <- ping

        }()

        ticker := time.NewTicker(10 * time.Second)
        counter := 0
        rand.Seed(time.Now().UnixNano()) 

        for w.running{
            // heart beat
            select {
            case w.communication.heartbeat <-ping:
            default:
                // just continue if previous ping hasn't been read
            }
            // work
            select {
            case <-ticker.C:
                n := rand.Intn(10)
                counter = counter + n
                logger.Printf("worker value: %d", counter)
                if counter % 4 == 0 {
                    // put actual message on channel passed to worker
                    // messageChannel <- message

                    // ping the notify channel to inform the listener there is a message to process
                    logger.Println("send notify")
                    select {
                    case w.communication.notify <- ping:
                    default:
                    }
                }
                if counter % 14 == 0 {
                    logger.Println("send panic")
                    panic("something went very wrong")
                }
            default:
            }
        }
    }()
}

func (w *ExampleWorker) Stop() {
    w.running = false
}

func (w *ExampleWorker) Heartbeat() <-chan struct{} {
    w.communication.heartbeat = make(chan struct{}, 1)  // or unbuffered
    return w.communication.heartbeat
}
func (w *ExampleWorker) Done() <-chan struct{} {
        w.communication.done = make(chan struct{}, 1)
    return w.communication.done
}
func (w *ExampleWorker) Notify() <-chan struct{} {
        w.communication.notify = make(chan struct{}, 1)
    return w.communication.notify
}
func (w *ExampleWorker) Err() <-chan error {
        w.communication.err =  make(chan error, 1)
    return w.communication.err
}

