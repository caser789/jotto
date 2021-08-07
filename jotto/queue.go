package jotto

import (
	"encoding/json"
	"fmt"
	"time"
)

// Job represents a async job that can be queued.
type Job struct {
	TraceID     string
	Type        int
	Payload     string
	Attempts    int64
	LastAttempt int64
}

func (job *Job) String() string {
	return job.Serialize()
}

// Serialize serializes a job into string
func (job *Job) Serialize() (str string) {
	bytes, err := json.Marshal(job)

	if err != nil {
		panic(err)
	}

	return string(bytes)
}

// Unserialize decodes a job from a string
func (job *Job) Unserialize(str string) (err error) {
	return json.Unmarshal([]byte(str), job)
}

// QueueStats is a collection of stats of a queue
type QueueStats struct {
	Pending int64 // Number of jobs ready to be consumed
	Working int64 // Number of jobs currently being processed
	Failure int64 // Number of jobs that have failed
	Delayed int64 // Number of jobs that have been delayed
	Backlog int64 // Total number of jobs
	Waiting int64 // Number of jobs currently in the delayed queue and waiting to be placed back to the pending queue
}

func (qs *QueueStats) String() string {
	return fmt.Sprintf("QueueStats (pending=%d, working=%d, failure=%d, delayed=%d, backlog=%d, waiting=%d)",
		qs.Pending, qs.Working, qs.Failure, qs.Delayed, qs.Backlog, qs.Waiting,
	)
}

// QueueDriver defines the interface for a queue driver
type QueueDriver interface {
	// Send a job to queue
	Enqueue(queue string, job *Job) error

	// Retrieve jobs from queue
	Dequeue(queue string) (*Job, error)

	// Increase the attempt count of a job and requeue it
	Requeue(queue string, job *Job) error

	// Mark a job as completed and remove it from the queue
	Complete(queue string, job *Job) error

	// Defer a job to be processed at a later time
	Defer(queue string, job *Job, after time.Duration) error

	// Mark a job as failed and move it to the failed list
	Fail(queue string, job *Job) error

	// Clear the queue (All data will be lost)
	Truncate(queue string) error

	// Get the stats of a quuee
	Stats(queue string) (*QueueStats, error)

	// Move deferred jobs that are ready to the pending queue
	ScheduleDeferred(queue string) (int64, error)
}

type QueueProcessor func(Application, Logger, *Job) error

type Queue interface {
}
