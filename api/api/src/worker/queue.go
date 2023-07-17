package worker

import "time"

type Queue interface {
	// Enqueue adds a task to the queue.
	Enqueue(task *Task) error

	// Poll returns a task from the queue, if one is scheduled to run at the
	// time of calling.
	Poll() (task *Task, found bool, err error)

	// Reschedules the task to a later time.
	Reschedule(task *Task, to time.Time) error

	// Dequeue deletes the task from the queue.
	Dequeue(task *Task) error
}
