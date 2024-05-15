package cdule

import (
	"encoding/json"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/gagasdiv/cdule/pkg"
	"github.com/gagasdiv/cdule/pkg/model"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

// ScheduleWatcher struct
type ScheduleWatcher struct {
	Closed         chan struct{}
	WG             sync.WaitGroup
	TickDuration   time.Duration
	Ticker         *time.Ticker
	RunImmediately bool
}

var lastScheduleExecutionTime int64
var nextScheduleExecutionTime int64

// Run to run watcher in a continuous loop
func (t *ScheduleWatcher) Run() {
	runJobs := func () {
		now := time.Now()
		lastScheduleExecutionTime = now.Add(-1 * t.TickDuration).UnixNano()
		nextScheduleExecutionTime = now.UnixNano()

		log.Debugf("lastScheduleExecutionTime %d, nextScheduleExecutionTime %d", lastScheduleExecutionTime, nextScheduleExecutionTime)
		runNextScheduleJobs(lastScheduleExecutionTime, nextScheduleExecutionTime)
	}

	if t.RunImmediately {
		runJobs();
	}

	for {
		select {
		case <-t.Closed:
			return
		case <-t.Ticker.C:
			runJobs()
		}
	}
}

// Stop to stop scheduler watcher
func (t *ScheduleWatcher) Stop() {
	close(t.Closed)
	t.WG.Wait()
}

func runNextScheduleJobs(scheduleStart, scheduleEnd int64) {
	schedules, err := model.CduleRepos.CduleRepository.GetScheduleBetween(scheduleStart, scheduleEnd, WorkerID)
	if nil != err {
		log.Error(err)
		return
	}

	runScheduleJobs(schedules)

	log.Debugf("Schedules Completed For StartTime %d To EndTime %d", scheduleStart, scheduleEnd)
}

func runScheduleJobs(schedules []model.Schedule) {
	defer panicRecoveryForSchedule()

	workers, err := model.CduleRepos.CduleRepository.GetAliveWorkers()
	if nil != err {
		log.Error(err)
		return
	}
	for _, schedule := range schedules {
		scheduledJob, err := model.CduleRepos.CduleRepository.GetJob(schedule.JobID)
		if nil != err {
			log.Errorf("Error while running Schedule for %d : %s", schedule.JobID, err.Error())
			continue
		}
		if scheduledJob == nil {
			log.Debugf("Schedule job is nil for worker_id %s, skipping", WorkerID)
			continue
		}
		log.Debug("====START====")
		log.Debugf("Schedule for JobName: %s, Exeuction Time %d at Worker %s", scheduledJob.JobName, schedule.ExecutionID, schedule.WorkerID)
		jobDataStr := schedule.JobData
		var jobDataMap map[string]string
		if pkg.EMPTYSTRING != jobDataStr {
			err = json.Unmarshal([]byte(jobDataStr), &jobDataMap)
			if nil != err {
				log.Error(err)
				continue
			}
		}
		var jobHistory *model.JobHistory
		if err == nil {
			jobHistory, err = model.CduleRepos.CduleRepository.GetJobHistoryForSchedule(schedule.ID)
			j, ok := JobRegistry[scheduledJob.JobName]
			if !ok {
				log.Errorf("Error while running Schedule for %d : unregistered job %s", schedule.JobID, scheduledJob.JobName)
				// Change run status to failed
				if err != nil && jobHistory != nil {
					jobHistory.Status = model.JobStatusFailed
					model.CduleRepos.CduleRepository.UpdateJobHistory(jobHistory)
				}
				continue
			}
			jobInstance := reflect.New(j).Interface()

			if err != nil && err.Error() == "record not found" && jobHistory != nil {
				// if job history was present but not executed
				if jobHistory.Status == model.JobStatusNew {
					jobHistory.Status = model.JobStatusInProgress
					model.CduleRepos.CduleRepository.UpdateJobHistory(jobHistory)
					jobDataMap = executeJob(jobInstance, jobHistory, &jobDataMap)
				}
			} else {
				// if job history is not there for this schedule, so this should be executed.
				jobHistory = &model.JobHistory{
					JobID:       schedule.JobID,
					ScheduleID:  schedule.ID,
					Status:      model.JobStatusNew,
					WorkerID:    schedule.WorkerID,
					RetryCount:  0,
				}
				model.CduleRepos.CduleRepository.CreateJobHistory(jobHistory)
				jobHistory.Status = model.JobStatusInProgress
				model.CduleRepos.CduleRepository.UpdateJobHistory(jobHistory)

				jobDataMap = executeJob(jobInstance, jobHistory, &jobDataMap)
				log.Debugf("Job Execution Completed For JobName: %s JobID: %d on Worker: %s", scheduledJob.JobName, schedule.JobID, schedule.WorkerID)
				log.Debug("====END====\n")
			}

			if scheduledJob.Once || scheduledJob.CronExpression == "" {
				log.Debugf("Job Only Once For JobName: %s JobID: %d on Worker: %s, skipping calculation for next schedule", scheduledJob.JobName, schedule.JobID, schedule.WorkerID)
				continue
			}

			// Calculate the next schedule for the current job
			jobDataBytes, err := json.Marshal(jobDataMap)
			if nil != err {
				log.Errorf("Error %s for JobName %s and Schedule ID %d ", err.Error(), scheduledJob.JobName, schedule.ExecutionID)
			}
			if string(jobDataBytes) != pkg.EMPTYSTRING {
				jobDataStr = string(jobDataBytes)
			}
			SchedulerParser, err := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse(scheduledJob.CronExpression)
			if err != nil {
				log.Error(err.Error())
				return
			}
			nextRunTime := SchedulerParser.Next(time.Now()).UnixNano()

			workerIDForNextRun, _ := findNextAvailableWorker(workers, schedule)
			newSchedule := model.Schedule{
				ExecutionID: nextRunTime,
				WorkerID:    workerIDForNextRun,
				JobID:       schedule.JobID,
				JobData:     jobDataStr,
			}
			model.CduleRepos.CduleRepository.CreateSchedule(&newSchedule)
			log.Debugf("*** Next Job Scheduled Info ***\n JobName: %s,\n Schedule Cron: %s,\n Job Scheduled Time: %d,\n Worker: %s ",
				scheduledJob.JobName, scheduledJob.CronExpression, newSchedule.ExecutionID, newSchedule.WorkerID)
		}
	}
}

// WorkerJobCount struct
type WorkerJobCount struct {
	WorkerID string `json:"worker_id"`
	Count    int64  `json:"count"`
}

func findNextAvailableWorker(workers []model.Worker, schedule model.Schedule) (string, error) {
	workerName := schedule.WorkerID
	var workerJobCountMetrics []WorkerJobCount
	model.DB.Raw("SELECT worker_id, count(1) FROM job_histories WHERE job_id = ? group by worker_id", schedule.JobID).Scan(&workerJobCountMetrics)
	log.Debugf("workerJobCountMetrics %v", workerJobCountMetrics)
	if len(workerJobCountMetrics) <= 0 {
		log.Debugf("workerName %s would be used", workerName)
		return workerName, nil
	}
	for _, worker := range workers {
		appendWorker := true
		for _, v := range workerJobCountMetrics {
			if v.WorkerID == worker.WorkerID {
				appendWorker = false
				break
			}
		}
		if appendWorker {
			newWorkerMetric := WorkerJobCount{
				WorkerID: worker.WorkerID,
				Count:    0,
			}
			workerJobCountMetrics = append(workerJobCountMetrics, newWorkerMetric)
		}
	}
	sort.Slice(workerJobCountMetrics[:], func(i, j int) bool {
		return workerJobCountMetrics[i].Count < workerJobCountMetrics[j].Count
	})
	return workerJobCountMetrics[0].WorkerID, nil
}

/*
For go 1.17 following method can be used.
func executeJob(jobInstance interface{}, jobHistory *model.JobHistory, jobDataMap map[string]string) {
	defer panicRecovery(jobHistory)
	jobInstance.(job.Job).Execute(jobDataMap)
}
*/

/*
cdule library has been built and developed using go 1.18 (go1.18beta2), if you need to use it for 1.17
then build from source by uncommenting the above method and comment the following
*/
func executeJob(jobInstance any, jobHistory *model.JobHistory, jobDataMap *map[string]string) map[string]string {
	defer panicRecovery(jobHistory)
	job := jobInstance.(Job)
	job.Execute(*jobDataMap)
	return job.GetJobData()
}

// If there is any panic from Job Execution, set the JobStatus as FAILED
func panicRecovery(jobHistory *model.JobHistory) {
	// TODO should be handled for any panic and set the status as FAILED for job history with error message
	jobHistory.Status = model.JobStatusCompleted
	if r := recover(); r != nil {
		log.Warning("Recovered in panicRecovery for job execution ", r)
		jobHistory.Status = model.JobStatusFailed
	}
	model.CduleRepos.CduleRepository.UpdateJobHistory(jobHistory)
}

func panicRecoveryForSchedule() {
	if r := recover(); r != nil {
		log.Warning("Recovered in runNextScheduleJobs ", r)
	}
}
