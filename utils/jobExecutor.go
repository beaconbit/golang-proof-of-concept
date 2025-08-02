package utils

type Job[T any] func(resource T)

// JobExecutor is a generic executor that operates on jobs using a shared resource of type T
type JobExecutor[T any] struct {
    resource T
    jobCh    <-chan Job[T]
}

// NewJobExecutor creates a new JobExecutor for a given resource and job channel
func NewJobExecutor[T any](resource T, jobCh <-chan Job[T]) *JobExecutor[T] {
    exec := &JobExecutor[T]{
        resource: resource,
        jobCh:    jobCh,
    }
    return exec
}

// Run starts processing jobs using the shared resource
func (e *JobExecutor[T]) Run() {
    for job := range e.jobCh {
        job(e.resource)
    }
}
