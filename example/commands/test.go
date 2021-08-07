package commands

import (
	"flag"
	"fmt"
	"time"

	"git.garena.com/duanzy/motto/motto"
)

type Test struct {
	motto.BaseCommand
	text string
}

func NewTest() *Test {
	return &Test{}
}

func (i *Test) Name() string {
	return "test"
}

func (i *Test) Description() string {
	return "Test is a playground"
}

func (i *Test) Boot(flagSet *flag.FlagSet) (err error) {
	flagSet.StringVar(&i.text, "text", "Zhiyan", "Part of the test payload")

	return
}

func (i *Test) Run(app motto.Application, args []string) (err error) {
	Q := app.Queue("default:main")

	Q.Driver().Truncate("main")

	for j := 0; j < 11; j++ {
		job := &motto.Job{
			Type:        2,
			Payload:     fmt.Sprintf(`{"name": "%s", "age": 31}`, i.text),
			Attempts:    0,
			LastAttempt: 0,
		}
		Q.Enqueue(job)
	}

	return

	job := &motto.Job{
		Type:        1,
		Payload:     fmt.Sprintf(`{"name": "%s", "age": 31}`, i.text),
		Attempts:    0,
		LastAttempt: 0,
	}

	err = Q.Schedule(job, time.Now().Add(time.Second*10))

	fmt.Println("schedule", err)

	return

	err = Q.Enqueue(job)

	fmt.Println("enqueue", err)

	job, err = Q.Dequeue()

	fmt.Println("dequeue", err, job)

	err = Q.Requeue(job)

	fmt.Println("requeue", err)

	job, err = Q.Dequeue()

	fmt.Println("dequeue", err, job)

	err = Q.Defer(job, time.Duration(1)*time.Second*10)

	fmt.Println("defer", err)

	return
}
