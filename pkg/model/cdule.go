package model

import (
	"time"

	"gorm.io/gorm"
)

// JobStatus for job status
type JobStatus string

const (
	// JobStatusNew status NEW
	JobStatusNew JobStatus = "NEW"
	// JobStatusInProgress status IN_PROGRESS
	JobStatusInProgress JobStatus = "IN_PROGRESS"
	// JobStatusCompleted status COMPLETED
	JobStatusCompleted JobStatus = "COMPLETED"
	// JobStatusFailed status FAILED
	JobStatusFailed JobStatus = "FAILED"
)

// Model common model
type Model struct {
	ID        int64 `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at; index"`
}

// Job struct
type Job struct {
	Model
	JobName        string `gorm:"index" json:"job_name"`
	SubName        string `json:"sub_name"`
	CronExpression string `json:"cron"`
	Expired        bool   `json:"expired"`
	Once           bool   `json:"once"`
	JobData        string `json:"job_data"`
}

// Schedule used by Execution Routine to execute a scheduled job in the evert one minute duration
type Schedule struct {
	Model
	ExecutionID int64 `json:"execution_id"`
	JobID       int64 `json:"job_id"`
	Job         Job   `gorm:"foreignKey:job_id;references:id;constraint:OnDelete:CASCADE"`
	WorkerID    string         `json:"worker_id"`
	JobData     string         `json:"job_data"`
}

// JobHistory struct
type JobHistory struct {
	Model
	JobID       int64          `json:"job_id"`
	Job         Job       `gorm:"foreignKey:job_id;references:id;constraint:OnDelete:CASCADE"`
	ScheduleID  int64          `json:"schedule_id"`
	Schedule    Schedule  `gorm:"foreignKey:schedule_id;references:id;constraint:OnDelete:CASCADE"`
	Status      JobStatus      `json:"status"`
	WorkerID    string         `json:"worker_id"`
	RetryCount  int            `json:"retry_count"`
}

// Worker Node health check via the heartbeat
type Worker struct {
	WorkerID  string `gorm:"primaryKey" json:"worker_id"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// WorkerJobCount struct
type WorkerJobCount struct {
	WorkerID string `json:"worker_id"`
	Count    int64  `json:"count"`
}
