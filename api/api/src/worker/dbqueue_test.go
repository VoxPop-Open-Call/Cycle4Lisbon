package worker

import (
	"encoding/json"
	"regexp"
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

type DbQueueTestSuite struct {
	suite.Suite
	db    *gorm.DB
	queue Queue
}

type JobArgs struct {
	Arg0 string
	Arg1 int
	Arg2 bool
}

func (s *DbQueueTestSuite) SetupSuite() {
	s.db.Exec("DELETE FROM worker_tasks")
}

func (s *DbQueueTestSuite) TearDownTest() {
	s.db.Exec("DELETE FROM worker_tasks")
}

func (s *DbQueueTestSuite) TestEnqueue() {
	args, err := json.Marshal(&JobArgs{
		Arg0: "abc",
		Arg1: 16,
		Arg2: true,
	})

	// ----------------------- //
	// Enqueue task to run now //
	// ----------------------- //
	jobName := "job-name"
	err = s.queue.Enqueue(&Task{
		TaskConfig: TaskConfig{
			JobName: jobName,
			Args:    args,
		},
	})
	s.NoError(err)

	var task models.WorkerTask
	result := s.db.First(&task, "job = ?", jobName)
	s.NoError(result.Error)

	s.NotEmpty(task)
	s.NotEmpty(task.ID)
	s.NotEmpty(task.CreatedAt)
	s.NotEmpty(task.UpdatedAt)
	s.NotEmpty(task.Args)
	s.WithinDuration(time.Now(), task.ScheduledTo, time.Second)
	s.Equal(uint16(0), task.Tries)
	s.False(task.Running)

	s.Equal(args, task.Args)

	// ---------------------------------- //
	// Enqueue task with a scheduled time //
	// ---------------------------------- //
	jobName = "job-name-scheduled"
	scheduledTo := time.Now().Add(time.Hour)
	err = s.queue.Enqueue(&Task{
		TaskConfig: TaskConfig{
			JobName:     jobName,
			Args:        args,
			ScheduledTo: scheduledTo,
		},
	})
	s.NoError(err)

	var scheduledTask models.WorkerTask
	result = s.db.First(&scheduledTask, "job = ?", jobName)
	s.NoError(result.Error)

	s.NotEmpty(scheduledTask)
	s.NotEmpty(scheduledTask.ID)
	s.NotEmpty(scheduledTask.CreatedAt)
	s.NotEmpty(scheduledTask.UpdatedAt)
	s.NotEmpty(scheduledTask.Args)
	s.Equal(uint16(0), scheduledTask.Tries)
	s.False(scheduledTask.Running)

	s.WithinDuration(scheduledTo, scheduledTask.ScheduledTo, time.Second)

	s.Equal(args, task.Args)

	// ------------------------------------------------------ //
	// Cannot enqueue the same task (same job and args) twice //
	// ------------------------------------------------------ //
	err = s.queue.Enqueue(&Task{
		TaskConfig: TaskConfig{
			JobName:     jobName,
			Args:        args,
			ScheduledTo: time.Now().Add(10 * time.Minute),
		},
	})
	s.Error(err)
	s.Regexp(
		regexp.MustCompile("duplicate key value violates unique constraint"),
		err.Error(),
	)
}

func scheduleRandomTask(q Queue, t time.Time) (jobName string, err error) {
	args, err := json.Marshal(&JobArgs{
		Arg0: random.String(10),
		Arg1: random.Int(0, 100),
		Arg2: false,
	})
	if err != nil {
		return "", err
	}

	jobName = random.String(20)
	err = q.Enqueue(&Task{
		TaskConfig: TaskConfig{
			JobName:     jobName,
			Args:        args,
			ScheduledTo: t,
		},
	})

	return
}

func (s *DbQueueTestSuite) TestPollUpdatesRunningFlag() {
	jobName, err := scheduleRandomTask(s.queue, time.Now())
	s.NotNil(jobName)
	s.NoError(err)

	task, ok, err := s.queue.Poll()
	s.NotEmpty(task)
	s.True(ok)
	s.NoError(err)
	s.Equal(jobName, task.JobName)

	var dbTask models.WorkerTask
	result := s.db.First(&dbTask, "job = ?", jobName)
	s.NoError(result.Error)
	s.NotEmpty(dbTask)
	s.True(dbTask.Running)
}

func (s *DbQueueTestSuite) TestPollOrder() {
	for i := 0; i < 5; i++ {
		r := time.Duration(random.Int(1, 30))

		jobName, err := scheduleRandomTask(
			s.queue,
			time.Now().Add(-r*time.Minute),
		)
		s.NotEmpty(jobName)
		s.NoError(err)
	}

	var tasks []*models.WorkerTask
	result := s.db.Order("scheduled_to asc").Find(&tasks)
	s.NoError(result.Error)
	s.NotEmpty(tasks)
	s.Equal(5, len(tasks))

	lastTime := time.Now().Add(-time.Hour)
	for _, task := range tasks {
		qTask, ok, err := s.queue.Poll()
		s.True(ok)
		s.NotEmpty(qTask)
		s.NoError(err)

		s.LessOrEqual(lastTime.UnixMicro(), task.ScheduledTo.UnixMicro())
		lastTime = task.ScheduledTo
		s.Equal(task.ID, qTask.ID)
		s.Equal(task.Job, string(qTask.JobName))
		s.Equal(task.ScheduledTo, qTask.ScheduledTo)
	}

	// No more tasks in the queue
	task, ok, err := s.queue.Poll()
	s.Empty(task)
	s.False(ok)
	s.NoError(err)
}

func (s *DbQueueTestSuite) TestPoll() {
	var tasks []*models.WorkerTask
	result := s.db.Order("scheduled_to asc").Find(&tasks)
	s.NoError(result.Error)
	s.Empty(tasks)
	s.Equal(0, len(tasks))

	// ----------------------------------- //
	// Returns false if there are no tasks //
	// ----------------------------------- //
	task, ok, err := s.queue.Poll()
	s.Nil(task)
	s.False(ok)
	s.NoError(err)

	// ----------------------------------------------- //
	// Returns true after enqueueing an immediate task //
	// ----------------------------------------------- //
	job, err := scheduleRandomTask(s.queue, time.Now())
	s.NotEmpty(job)
	s.NoError(err)
	task, ok, err = s.queue.Poll()
	s.NotEmpty(task)
	s.True(ok)
	s.NoError(err)

	task, ok, err = s.queue.Poll()
	s.Nil(task)
	s.False(ok)
	s.NoError(err)

	// ---------------------------------------------------- //
	// Returns false after scheduling a task for the future //
	// ---------------------------------------------------- //
	job, err = scheduleRandomTask(s.queue, time.Now().Add(time.Hour))
	s.NotEmpty(job)
	s.NoError(err)
	task, ok, err = s.queue.Poll()
	s.Nil(task)
	s.False(ok)
	s.NoError(err)

	// ----------------------------------------------------- //
	// Returns true even after the scheduled time has passed //
	// ----------------------------------------------------- //
	job, err = scheduleRandomTask(s.queue, time.Now().Add(-2*time.Hour))
	s.NotEmpty(job)
	s.NoError(err)
	task, ok, err = s.queue.Poll()
	s.NotEmpty(task)
	s.True(ok)
	s.NoError(err)

	task, ok, err = s.queue.Poll()
	s.Nil(task)
	s.False(ok)
	s.NoError(err)

	// ---------------------------------------------------- //
	// Returns an immediate task with other tasks scheduled //
	// ---------------------------------------------------- //
	job, err = scheduleRandomTask(s.queue, time.Now().Add(time.Hour))
	s.NotEmpty(job)
	s.NoError(err)

	job, err = scheduleRandomTask(s.queue, time.Now().Add(2*time.Hour))
	s.NotEmpty(job)
	s.NoError(err)

	job, err = scheduleRandomTask(s.queue, time.Now().Add(time.Minute))
	s.NotEmpty(job)
	s.NoError(err)

	job, err = scheduleRandomTask(s.queue, time.Now())
	s.NotEmpty(job)
	s.NoError(err)

	task, ok, err = s.queue.Poll()
	s.NotEmpty(task)
	s.True(ok)
	s.NoError(err)

	s.Equal(job, task.JobName)
}

func (s *DbQueueTestSuite) TestReschedule() {
	job, err := scheduleRandomTask(s.queue, time.Now())
	s.NotEmpty(job)
	s.NoError(err)

	task, ok, err := s.queue.Poll()
	s.NotEmpty(task)
	s.True(ok)
	s.NoError(err)

	err = s.queue.Reschedule(task, time.Now().Add(time.Hour))
	s.NoError(err)

	var dbTask models.WorkerTask
	result := s.db.First(&dbTask, "id = ?", task.ID.String())
	s.NotEmpty(dbTask)
	s.NoError(result.Error)

	s.Equal(dbTask.ID, dbTask.ID)
	s.Equal(dbTask.Job, dbTask.Job)

	s.WithinDuration(time.Now(), task.ScheduledTo, time.Second)
	s.WithinDuration(time.Now().Add(time.Hour), dbTask.ScheduledTo, time.Second)

	s.False(dbTask.Running)
	s.Equal(uint16(1), dbTask.Tries)
	s.WithinDuration(time.Now(), dbTask.UpdatedAt, time.Second)
}

func (s *DbQueueTestSuite) TestDelete() {
	job, err := scheduleRandomTask(s.queue, time.Now())
	s.NotEmpty(job)
	s.NoError(err)

	task, ok, err := s.queue.Poll()
	s.NotEmpty(task)
	s.True(ok)
	s.NoError(err)

	err = s.queue.Dequeue(task)
	s.NoError(err)

	var dbTask models.WorkerTask
	result := s.db.First(&dbTask, "id = ?", task.ID.String())
	s.Empty(dbTask)
	s.Error(result.Error)
	s.Equal(gorm.ErrRecordNotFound, result.Error)
}

func TestPgQueue(t *testing.T) {
	config, err := config.Load("../../.env")
	require.NoError(t, err)

	testDb, err := database.Init(config.DbDsn())
	require.NoError(t, err)

	suite.Run(t, &DbQueueTestSuite{
		db:    testDb,
		queue: NewDbQueue(testDb),
	})
}
