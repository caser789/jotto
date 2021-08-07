package jotto

import (
	"encoding/json"
)

type Job struct {
	Type        int
	Payload     string
	Attempts    int64
	LastAttempt int64
}

func (job *Job) String() string {
	return job.Serialize()
}

func (job *Job) Serialize() (str string) {
	bytes, err := json.Marshal(job)

	if err != nil {
		panic(err)
	}

	return string(bytes)
}

func (job *Job) Unserialize(str string) (err error) {
	return json.Unmarshal([]byte(str), job)
}

type QueueDriver interface {
	Push(queue string, job *Job) error
	Pop(queue string) (*Job, error)
}

type QueueProcessor func(Application, Logger, *Job) error
