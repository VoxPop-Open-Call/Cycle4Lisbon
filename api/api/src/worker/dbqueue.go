package worker

import (
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"gorm.io/gorm"
)

type DbQueue struct {
	db *gorm.DB
}

// Creates a new DbQueue, which implements a Queue with a database as the
// backend.
func NewDbQueue(db *gorm.DB) *DbQueue {
	return &DbQueue{db}
}

func (q *DbQueue) Enqueue(task *Task) error {
	return q.db.Create(&models.WorkerTask{
		Job:         task.JobName,
		Args:        task.Args,
		ScheduledTo: task.ScheduledTo,
	}).Error
}

func (q *DbQueue) Poll() (*Task, bool, error) {
	task := &models.WorkerTask{}
	result := q.db.Raw(`
		WITH task AS (
			SELECT id FROM worker_tasks
			WHERE
				running = false AND
				scheduled_to <= now()
			ORDER BY scheduled_to ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE worker_tasks SET
			running = true,
			updated_at = now()
		FROM task
		WHERE worker_tasks.id = task.id
		RETURNING *
	`).Scan(&task)

	if result.RowsAffected == 0 {
		return nil, false, nil
	}

	return &Task{
		ID:    task.ID,
		Tries: task.Tries,
		TaskConfig: TaskConfig{
			JobName:     task.Job,
			Args:        task.Args,
			ScheduledTo: task.ScheduledTo,
		},
	}, true, result.Error
}

func (q *DbQueue) Reschedule(task *Task, to time.Time) error {
	return q.db.Exec(`
		WITH task AS (
			SELECT id FROM worker_tasks
			WHERE id = $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE worker_tasks SET
			running = false,
			tries = tries + 1,
			scheduled_to = $2,
			updated_at = now()
		FROM task
		WHERE worker_tasks.id = task.id
	`, task.ID.String(), to).Error
}

func (q *DbQueue) Dequeue(task *Task) error {
	return q.db.Table("worker_tasks").Delete(task).Error
}
