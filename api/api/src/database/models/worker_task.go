package models

import "time"

// WorkerTask is an indication to perform a job at a certain time and with a set
// of arguments. The unique constraint on Job and Args ensures no duplication.
// See the `worker` package.
type WorkerTask struct {
	BaseModel
	ScheduledTo time.Time `gorm:"not null;default:now();index:,sort:asc"`
	Job         string    `gorm:"uniqueIndex:idx_job_args;type:varchar(20);not null"`
	Args        []byte    `gorm:"uniqueIndex:idx_job_args;not null;default:'\\000'::bytea"`
	Tries       uint16    `gorm:"not null;default:0"`
	Running     bool      `gorm:"not null;default:false"`
}
