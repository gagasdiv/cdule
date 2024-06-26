package cdule

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/gagasdiv/cdule/pkg"
	"github.com/gagasdiv/cdule/pkg/model"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

// JobRegistry job registry
var JobRegistry = make(map[string]reflect.Type)

// ScheduleParser cron parser
var ScheduleParser cron.Parser

func RegisterType(job Job) {
	t := reflect.TypeOf(job).Elem()
	JobRegistry[job.JobName()] = t
}

// AbstractJob for holding job and jobdata
type AbstractJob struct {
	Job     Job
	JobData map[string]string
	SubName string
}

// NewJob to create new abstract job
func NewJob(job Job, jobData map[string]string, subName ...string) *AbstractJob {
	aj := &AbstractJob{
		Job:     job,
		JobData: jobData,
	}
	if len(subName) > 0 {
		aj.SubName = subName[0]
	}
	return aj
}

// Build to build job and store in the database
func (j *AbstractJob) Build(cronExpression string) (*model.Job, error) {
	jobDataBytes, err := json.Marshal(j.JobData)
	/*if nil != err {
		log.Errorf("Error %s for JobName %s", err.Error(), j.Job.JobName())
		return nil, fmt.Errorf("invalid Job Data %v", j.JobData)
	}*/
	var jobDataStr = ""
	if string(jobDataBytes) != pkg.EMPTYSTRING {
		jobDataStr = string(jobDataBytes)
	}
	newJob := &model.Job{
		JobName:        j.Job.JobName(),
		SubName:        j.SubName,
		CronExpression: cronExpression,
		Expired:        false,
		JobData:        jobDataStr,
		Once:           false,
	}
	SchedulerParser, err := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse(newJob.CronExpression)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	nextRunTime := SchedulerParser.Next(time.Now()).UnixNano()
	firstSchedule := &model.Schedule{
		ExecutionID: nextRunTime,
		WorkerID:    WorkerID,
		JobData:     newJob.JobData,
	}
	job, _, err := j.buildFirstSchedule(newJob, firstSchedule)
	return job, err
}

// BuildToRunAt to build job to run only once and store in the database
func (j *AbstractJob) BuildToRunAt(t time.Time) (*model.Job, error) {
	jobDataBytes, err := json.Marshal(j.JobData)
	/*if nil != err {
		log.Errorf("Error %s for JobName %s", err.Error(), j.Job.JobName())
		return nil, fmt.Errorf("invalid Job Data %v", j.JobData)
	}*/
	var jobDataStr = ""
	if string(jobDataBytes) != pkg.EMPTYSTRING {
		jobDataStr = string(jobDataBytes)
	}
	newJob := &model.Job{
		JobName:        j.Job.JobName(),
		SubName:        j.SubName,
		CronExpression: "",
		Expired:        false,
		JobData:        jobDataStr,
		Once:           true,
	}
	nextRunTime := t.UnixNano()
	firstSchedule := &model.Schedule{
		ExecutionID: nextRunTime,
		WorkerID:    WorkerID,
		JobData:     newJob.JobData,
	}
	job, _, err := j.buildFirstSchedule(newJob, firstSchedule)
	return job, err
}

// BuildToRunIn to build job to run only once and store in the database
func (j *AbstractJob) BuildToRunIn(n time.Duration) (*model.Job, error) {
	return j.BuildToRunAt(time.Now().Add(n))
}

// BuildToRunNow to build job to run immediately only once and store in the database
func (j *AbstractJob) BuildToRunNow() (*model.Job, error) {
	return j.BuildToRunAt(time.Now())
}

// Build to build job and store in the database
func (j *AbstractJob) buildFirstSchedule(job *model.Job, schedule *model.Schedule) (*model.Job, *model.Schedule, error) {
	// register job, this is used later to get the type of a job
	RegisterType(j.Job)

	existingJob, err := model.CduleRepos.CduleRepository.GetRepeatingJobByName(j.Job.JobName())
	if err != nil {
		log.Error(err.Error())
		return nil, nil, err
	}
	if nil != existingJob && !job.Once {
		log.Debugf("Found a non-once Job with the same Name: %s", existingJob.JobName)
		CancelJob(existingJob.JobName, existingJob.SubName)
	}

	log.Debugf("Making new Job with Name: %s", job.JobName)
	job, err = model.CduleRepos.CduleRepository.CreateJob(job)
	if err != nil {
		log.Error(err.Error())
		return nil, nil, err
	}

	// Make first schedule
	schedule.JobID = job.ID
	_, err = model.CduleRepos.CduleRepository.CreateSchedule(schedule)
	if err != nil {
		log.Error(err.Error())
		return job, nil, err
	}
	log.Debugf("*** Job Scheduled Info ***\n JobName: %s,\n Schedule Cron: %s,\n Job Scheduled Time: %d,\n Worker: %s ",
		job.JobName, job.CronExpression, schedule.ExecutionID, schedule.WorkerID)
	return job, schedule, err
}

// CancelJob to delete schedules for a job in the database by jobName and subName
func CancelJob(jobName string, subName string) (error) {
	schedules, err := model.CduleRepos.CduleRepository.DeleteScheduleForJobName(jobName, subName)
	if err == nil {
		log.Debugf("Cancelled schedule(s) based on jobName: %#v and subName: %#v ; %d schedule(s) ", jobName, subName, len(schedules))
	} else {
		log.Warnf("Failed cancelling schedule(s) based on jobName: %#v and subName: %#v ; err: %s ", jobName, subName, err.Error())
	}
	return err
}
