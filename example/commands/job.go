package commands

import (
	"flag"
	"fmt"

	"git.garena.com/duanzy/motto/motto"
)

type Job struct {
	motto.BaseCommand
	text string
}

func NewJob() *Job {
	return &Job{}
}

func (i *Job) Name() string {
	return "job"
}

func (i *Job) Description() string {
	return "Job enqueues an async job to the queue"
}

func (i *Job) Boot(flagSet *flag.FlagSet) (err error) {
	flagSet.StringVar(&i.text, "text", "Zhiyan", "Part of the test payload")

	return
}

func (i *Job) Run(app motto.Application, args []string) (err error) {

	app.Queue("default").Enqueue(&motto.Job{
		Type:        1,
		Payload:     fmt.Sprintf(`{"name": "%s", "age": 31}`, i.text),
		Attempts:    0,
		LastAttempt: 0,
	})

	return
}
