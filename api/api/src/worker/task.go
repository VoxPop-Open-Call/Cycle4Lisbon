package worker

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TaskConfig struct {
	// JobName is the name of the job to run, which should be registered
	// beforehand.
	JobName string

	// Args is the value that will be sent to the job handler.
	Args []byte

	// ScheduledTo is the time at which the task should run.
	// Not setting this value will schedule the task to run as soon as
	// possible.
	ScheduledTo time.Time
}

type Task struct {
	TaskConfig
	ID    uuid.UUID
	Tries uint16
}

func (t *Task) String() string {
	return fmt.Sprintf(
		"%s at %v",
		t.JobName,
		t.ScheduledTo.UTC().Format(time.DateTime),
	)
}
