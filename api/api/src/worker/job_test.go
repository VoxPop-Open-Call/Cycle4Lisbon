package worker

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Delay and MaxDelay different from zero.
func TestDelayFunc(t *testing.T) {
	job := &Job{
		Retries:  6,
		Delay:    500 * time.Second,
		MaxDelay: 1000 * time.Second,
	}
	job.calculateDelayFunc()

	assert.Equal(t, 499.0, job.a)
	assert.True(t, math.Abs(job.b-3.467109) < 0.000001)

	tests := []time.Duration{
		500,
		502,
		511,
		540,
		643,
		1000,
	}

	for i, test := range tests {
		delay := job.nextDelay(uint16(i))
		assert.Equal(t, test*time.Second, delay)
	}
}

// Zero delay, and MaxDelay different from zero.
func TestZeroDelayFunc(t *testing.T) {
	job := &Job{
		Retries:  6,
		Delay:    0,
		MaxDelay: time.Minute,
	}
	job.calculateDelayFunc()

	assert.Equal(t, -1.0, job.a)
	assert.True(t, math.Abs(job.b-2.27544) < 0.00001)

	tests := []time.Duration{
		0,
		time.Second,
		4 * time.Second,
		10 * time.Second,
		25 * time.Second,
		time.Minute,
	}

	for i, test := range tests {
		delay := job.nextDelay(uint16(i))
		assert.Equal(t, test, delay)
	}
}

// Non-zero Delay, MaxDelay not defined.
func TestConstantDelay(t *testing.T) {
	job := &Job{
		Retries: 5,
		Delay:   time.Second,
	}
	job.calculateDelayFunc()

	for i := 0; i < 5; i++ {
		assert.Equal(t, time.Second, job.nextDelay(uint16(i)))
	}
}

// Zero Delay, zero MaxDelay.
func TestZeroDelay(t *testing.T) {
	job := &Job{
		Retries: 10,
	}
	job.calculateDelayFunc()

	for i := 0; i < 10; i++ {
		assert.Equal(t, time.Duration(0), job.nextDelay(uint16(i)))
	}
}

// 1 retry (division by zero).
func TestOneRetry(t *testing.T) {
	job := &Job{
		Retries: 1,
		Delay:   time.Second,
	}
	job.calculateDelayFunc()

	assert.Equal(t, time.Second, job.nextDelay(0))
}

func TestLongDelay(t *testing.T) {
	job := &Job{
		Retries:  9,
		Delay:    time.Minute,
		MaxDelay: 5 * time.Hour,
	}
	job.calculateDelayFunc()

	tests := []time.Duration{
		60 * time.Second,
		62 * time.Second,
		70 * time.Second,
		98 * time.Second,
		192 * time.Second,  // ~3m
		514 * time.Second,  // ~8m30s
		1609 * time.Second, // ~27m
		5332 * time.Second, // ~1h30m
		5 * time.Hour,
	}

	for i, test := range tests {
		assert.Equal(t, test, job.nextDelay(uint16(i)))
	}
}
