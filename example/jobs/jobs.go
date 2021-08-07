package jobs

import (
	"encoding/json"
	"errors"
	"time"

	"git.garena.com/duanzy/motto/motto"
)

var Jobs = map[int]motto.QueueProcessor{
	1: ProcessTestJob,
	2: LongRunningJob,
}

type Author struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func ProcessTestJob(Q *motto.Queue, job *motto.Job, app motto.Application, logger motto.Logger) (err error) {
	author := &Author{}

	json.Unmarshal([]byte(job.Payload), author)

	if job.Attempts < 3 {
		logger.Error("Job attempts (%d) < 3, fail", job.Attempts)
		return errors.New("fail")
	}

	logger.Data("Author: %+v\n", author)

	return
}

func LongRunningJob(Q *motto.Queue, job *motto.Job, app motto.Application, logger motto.Logger) (err error) {
	logger.Trace("Sleeping")
	time.Sleep(time.Second * 3)
	logger.Trace("Waking up")

	return
}
