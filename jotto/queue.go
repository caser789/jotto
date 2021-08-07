package jotto

type Job interface {
	Type() string
	Payload() string
	Attempts() int
}

type QueueDriver interface {
	Push(Job) error
	Pop() (Job, error)
}
