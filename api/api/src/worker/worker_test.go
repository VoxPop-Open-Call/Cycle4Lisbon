package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/config"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type WorkerTestSuite struct {
	suite.Suite
	queue Queue
	db    *gorm.DB
}

func (s *WorkerTestSuite) SetupSuite() {
	s.db.Exec("DELETE FROM worker_tasks")
}

func (s *WorkerTestSuite) TearDownTest() {
	s.db.Exec("DELETE FROM worker_tasks")
}

// Can't schedule a non-registered task.
func (s *WorkerTestSuite) TestScheduleNonRegistered() {
	worker := New(s.queue)
	worker.LogLevel = Silent

	err := worker.Schedule(&TaskConfig{JobName: "no-task"})
	s.Error(err)
	s.Equal("the job 'no-task' isn't registered", err.Error())
}

// The worker calls the function that pre-calculates the delay function variables.
func (s *WorkerTestSuite) TestJobInit() {
	worker := New(s.queue)
	job := &Job{
		Name:    "constant-delay",
		Retries: 1,
		Delay:   time.Minute,
		Handler: func(_ context.Context, _ []byte) error {
			return nil
		},
	}
	err := worker.Register(job)
	s.NoError(err)

	s.Equal(59.0, job.a)
	s.Equal(1.0, job.b)
	s.Equal(time.Minute, job.nextDelay(0))
}

// The job handler is called when a task is scheduled.
func (s *WorkerTestSuite) TestHandlerCalled() {
	worker := New(s.queue)
	worker.LogLevel = Silent

	ch := make(chan int, 1)
	job := &Job{
		Name: "test",
		Handler: func(_ context.Context, _ []byte) error {
			ch <- 1
			return nil
		},
	}
	err := worker.Register(job)
	s.NoError(err)
	err = worker.Schedule(&TaskConfig{
		JobName:     "test",
		ScheduledTo: time.Now(),
	})
	s.NoError(err)

	worker.update()

	select {
	case <-ch:
		// the handler was called, good
	case <-time.After(time.Second):
		s.Fail("timeout: the 'test' handler was not called")
	}
}

// No more than MaxRoutineNum routines run simultaneously.
func (s *WorkerTestSuite) TestMaxRoutineNum() {
	MaxRoutineNum = 2
	worker := New(s.queue)
	worker.LogLevel = Silent
	var m sync.Mutex
	m.Lock()
	chLocked := make(chan int, 1)
	job := &Job{
		Name: "locked",
		Handler: func(_ context.Context, _ []byte) error {
			chLocked <- 1
			m.Lock()
			m.Unlock()
			return nil
		},
	}
	err := worker.Register(job)
	s.NoError(err)

	chDelayed := make(chan int, 1)
	job2 := &Job{
		Name: "delayed",
		Handler: func(_ context.Context, _ []byte) error {
			chDelayed <- 1
			return nil
		},
	}
	err = worker.Register(job2)

	worker.Schedule(&TaskConfig{
		JobName:     "locked",
		ScheduledTo: time.Now().Add(-2 * time.Minute),
		Args:        []byte(random.String(30)),
	})
	worker.Schedule(&TaskConfig{
		JobName:     "locked",
		ScheduledTo: time.Now().Add(-time.Minute),
		Args:        []byte(random.String(30)),
	})
	worker.Schedule(&TaskConfig{
		JobName:     "delayed",
		ScheduledTo: time.Now(),
	})
	go worker.update() // the first locked should be called
	// wait for the handler to be called, to prevent the next call to update
	// from trying to dequeue the same task, and thus fail
	<-chLocked
	go worker.update() // the second locked should be called
	<-chLocked
	go worker.update() // the delayed job should not start

	select {
	case <-chDelayed:
		s.FailNow("the 'delayed' handler was called before it should")
	case <-time.After(400 * time.Millisecond):
		// the handler was NOT called, nice
	}

	// 'delayed' should run after the lock is released
	m.Unlock()

	select {
	case <-chDelayed:
		// the handler was called, sweet
	case <-time.After(time.Second):
		s.FailNow("timeout: the 'delayed' handler was not called")
	}
}

// The worker panics if it waits too long for a routine to become available.
func (s *WorkerTestSuite) TestWorkerTimeout() {
	MaxRoutineNum = 0
	worker := New(s.queue)
	worker.LogLevel = Silent
	worker.PollingSleepTime = time.Millisecond
	worker.GlobalTimeout = 10 * time.Millisecond

	job := &Job{
		Name: "never-runs",
		Handler: func(_ context.Context, _ []byte) error {
			return nil
		},
	}
	err := worker.Register(job)
	s.NoError(err)

	worker.Schedule(&TaskConfig{
		JobName:     "never-runs",
		ScheduledTo: time.Now(),
	})

	s.Panics(worker.update)
}

// When stoping the worker, it waits for runing tasks and times-out if it takes
// too long.
func (s *WorkerTestSuite) TestWaitGroupTimeout() {
	MaxRoutineNum = 1
	worker := New(s.queue)
	worker.LogLevel = Silent
	worker.PollingSleepTime = 10 * time.Millisecond

	ch := make(chan int)
	job := &Job{
		Name: "takes-forever",
		Handler: func(_ context.Context, _ []byte) error {
			ch <- 1
			time.Sleep(time.Minute)
			return nil
		},
	}
	err := worker.Register(job)
	s.NoError(err)

	go worker.Run()

	worker.Schedule(&TaskConfig{
		JobName:     "takes-forever",
		ScheduledTo: time.Now(),
	})
	<-ch

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err = worker.Stop(ctx)
	s.Error(err)
	s.Equal("timeout waiting for running tasks, stopping anyway", err.Error())
}

// The task is deleted from the queue after it runs successfully.
func (s *WorkerTestSuite) TestTaskDeleted() {
	worker := New(s.queue)
	worker.LogLevel = Silent
	job := &Job{
		Name: "don't-do-much",
		Handler: func(_ context.Context, _ []byte) error {
			return nil
		},
	}
	err := worker.Register(job)
	s.NoError(err)

	worker.Schedule(&TaskConfig{
		JobName:     job.Name,
		ScheduledTo: time.Now(),
	})

	task, ok, err := worker.queue.Poll()
	s.NotEmpty(task)
	s.True(ok)
	s.NoError(err)

	worker.handle(task)

	var dbTask models.WorkerTask
	result := s.db.First(&dbTask, "id = ?", task.ID.String())
	s.Error(result.Error)
	s.Equal(gorm.ErrRecordNotFound, result.Error)
	s.Empty(dbTask)
}

// A task is rescheduled if it fails.
func (s *WorkerTestSuite) TestTaskReschedule() {
	worker := New(s.queue)
	worker.LogLevel = Silent
	job := &Job{
		Name:    "failure",
		Retries: 5,
		Delay:   30 * time.Second,
		Handler: func(_ context.Context, _ []byte) error {
			return errors.New("AH!")
		},
	}
	err := worker.Register(job)
	s.NoError(err)

	worker.Schedule(&TaskConfig{
		JobName:     job.Name,
		ScheduledTo: time.Now(),
	})

	task, ok, err := worker.queue.Poll()
	s.NotEmpty(task)
	s.True(ok)
	s.NoError(err)

	worker.handle(task)

	var dbTask models.WorkerTask
	result := s.db.First(&dbTask, "id = ?", task.ID.String())
	s.NoError(result.Error)
	s.NotEmpty(dbTask)

	s.Equal(uint16(1), dbTask.Tries)
	s.False(dbTask.Running)
	s.WithinDuration(time.Now().Add(30*time.Second), dbTask.ScheduledTo, time.Second)
}

// A task is deleted when the retry number is reached.
func (s *WorkerTestSuite) TestMaxRetries() {
	worker := New(s.queue)
	worker.LogLevel = Silent
	job := &Job{
		Name:    "failure-2",
		Retries: 5,
		Delay:   30 * time.Second,
		Handler: func(_ context.Context, _ []byte) error {
			return errors.New("AH!")
		},
	}
	err := worker.Register(job)
	s.NoError(err)

	worker.Schedule(&TaskConfig{
		JobName:     job.Name,
		ScheduledTo: time.Now(),
	})

	result := s.db.Model(&models.WorkerTask{}).
		Where("job = ?", "failure-2").
		Update("tries", 5)
	s.NoError(result.Error)
	s.Equal(int64(1), result.RowsAffected)

	task, ok, err := worker.queue.Poll()
	s.NotEmpty(task)
	s.True(ok)
	s.NoError(err)

	worker.handle(task)

	var dbTask models.WorkerTask
	result = s.db.First(&dbTask, "id = ?", task.ID.String())
	s.Error(result.Error)
	s.Equal(gorm.ErrRecordNotFound, result.Error)
	s.Empty(dbTask)
}

// The worker calls the handler the correct number of times.
func (s *WorkerTestSuite) TestTaskRetries() {
	MaxRoutineNum = 1
	worker := New(s.queue)
	worker.PollingSleepTime = 10 * time.Millisecond
	worker.LogLevel = Silent
	ch := make(chan int, 1)
	job := &Job{
		Name:    "failure-3",
		Retries: 5,
		Handler: func(_ context.Context, _ []byte) error {
			ch <- 1
			return errors.New("AH!")
		},
	}
	err := worker.Register(job)
	s.NoError(err)

	worker.Schedule(&TaskConfig{
		JobName:     job.Name,
		ScheduledTo: time.Now(),
	})

	go worker.Run()
	defer worker.Stop(context.Background())

	// 1 run + 5 retries = 6
	for i := 0; i < 6; i++ {
		select {
		case <-ch:
			// the handler was called, cool
		case <-time.After(500 * time.Millisecond):
			s.FailNow(fmt.Sprintf("the handler was not called at iteration %d", i))
		}
	}

	go worker.update()
	select {
	case <-ch:
		s.FailNow("the handler was called more than 6 times")
	case <-time.After(300 * time.Millisecond):
		// and the handler was never called again
	}
}

// The worker calls OnSuccess and OnFailure accordingly.
func (s *WorkerTestSuite) TestJobCallbacks() {
	worker := New(s.queue)
	worker.LogLevel = Silent

	chS := make(chan int, 1)
	chF := make(chan int, 1)
	successJob := &Job{
		Name: "test-success",
		Handler: func(_ context.Context, _ []byte) error {
			return nil
		},
		OnSuccess: func() {
			chS <- 1
		},
		OnFailure: func() {
			chF <- 1
		},
	}
	err := worker.Register(successJob)
	s.NoError(err)

	err = worker.Schedule(&TaskConfig{
		JobName:     successJob.Name,
		ScheduledTo: time.Now(),
	})
	s.NoError(err)

	worker.update()

	select {
	case <-chS:
		// the success handler was called, good
	case <-time.After(time.Second):
		s.Fail("timeout: the OnSuccess callback was not called")
	}
	select {
	case <-chF:
		s.Fail("the OnFailure callback was called on a successful task")
	case <-time.After(500 * time.Millisecond):
		// the failure handler was NOT called, good
	}

	failureJob := &Job{
		Name: "test-failure",
		Handler: func(_ context.Context, _ []byte) error {
			return errors.New("oh no")
		},
		OnSuccess: func() {
			chS <- 1
		},
		OnFailure: func() {
			chF <- 1
		},
	}
	err = worker.Register(failureJob)
	s.NoError(err)

	err = worker.Schedule(&TaskConfig{
		JobName:     failureJob.Name,
		ScheduledTo: time.Now(),
	})
	s.NoError(err)

	worker.update()

	select {
	case <-chS:
		s.Fail("the OnSuccess callback was called on a failed task")
	case <-time.After(500 * time.Millisecond):
		// the handler was NOT called, good
	}
	select {
	case <-chF:
		// the success handler was called, good
	case <-time.After(time.Second):
		s.Fail("timeout: the OnFailure callback was not called")
	}
}

func TestWorker(t *testing.T) {
	config, err := config.Load("../../.env")
	require.NoError(t, err)

	db, err := database.Init(config.DbDsn())
	require.NoError(t, err)

	suite.Run(t, &WorkerTestSuite{
		queue: NewDbQueue(db),
		db:    db,
	})
}
