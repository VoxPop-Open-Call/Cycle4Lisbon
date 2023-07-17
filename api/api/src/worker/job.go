package worker

import (
	"context"
	"math"
	"time"
)

// Job defines a handler to do some work, and the retry plan in case it fails.
type Job struct {
	// Name is the unique identifier of the job.
	// Scheduling tasks will reference this name.
	Name string

	// The handler function of the job.
	// Returning an error schedules a retry.
	Handler func(ctx context.Context, args []byte) error

	// OnSuccess is called iff the handler succeeds, after the task is
	// dequeued.
	OnSuccess func()

	// OnFailure is called iff the handler fails all retries, after the task
	// is dequeued.
	OnFailure func()

	// Maximum number of retries a task for this job will have.
	// A value of 0 means the task will only try to run once.
	Retries uint16

	// Delay is the minimum delay between retries.
	Delay time.Duration

	// MaxDelay is the maximum delay between retries, which means the delay
	// between the second-to-last and last try.
	//
	// If MaxDelay is zero, the delay will be constant and equal to `Delay`.
	//
	// The delay between retries is calculated according to:
	// f(0) = delay
	// f(retries-1) = maxDelay
	// f(x) = a + b^x
	// `a` and `b` are pre-calculated.
	MaxDelay time.Duration
	a        float64
	b        float64
}

// Calculates `a` and `b` of the delay function.
func (j *Job) calculateDelayFunc() {
	if j.MaxDelay == 0 {
		j.MaxDelay = j.Delay
	}

	delay := j.Delay / time.Second
	maxDelay := j.MaxDelay / time.Second

	j.a = float64(delay) - 1.0
	base := float64(maxDelay - delay + 1.0)
	exp := 1.0 / (float64(j.Retries) - 1.0)
	j.b = math.Pow(base, exp)
}

// Calculates the delay until the next try (from `tries` to `tries+1`).
func (j *Job) nextDelay(tries uint16) time.Duration {
	result := j.a + math.Pow(j.b, float64(tries))
	return time.Duration(result) * time.Second
}
