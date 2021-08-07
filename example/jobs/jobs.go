package jobs

import (
	"encoding/json"
	"fmt"

	"git.garena.com/duanzy/motto/motto"
)

var Jobs = map[int]motto.QueueProcessor{
	1: ProcessTestJob,
}

type Author struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func ProcessTestJob(app motto.Application, logger motto.Logger, job *motto.Job) (err error) {
	author := &Author{}

	json.Unmarshal([]byte(job.Payload), author)

	if job.Attempts < 3 {
		return fmt.Errorf("Attempts (%d) < 3, fail", job.Attempts)
	}

	logger.Data("Author: %+v\n", author)

	return
}
