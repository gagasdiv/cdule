package cdule

import (
	"time"

	"github.com/gagasdiv/cdule/pkg/model"

	log "github.com/sirupsen/logrus"
)

// ScheduleWatcher struct
type PastScheduleWatcher struct {
	ScheduleWatcher
}

// Run to run watcher in a continuous loop
func (t *PastScheduleWatcher) Run() {
	runJobs := func () {
		// Adjust with schedule watcher so that there's no collision/duplication/race condition
		now := time.Now().Add(-1 * t.TickDuration)
		runPassedScheduleJobs(now.UnixNano())
	}

	if t.RunImmediately {
		runJobs()
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

func runPassedScheduleJobs(beforeTime int64) {
	schedules, err := model.CduleRepos.CduleRepository.GetPassedSchedule(beforeTime, WorkerID, true)
	if nil != err {
		log.Error(err)
		return
	}

	// Filter to only take passed single-run schedules
	filtered := make([]model.Schedule, 0)
	for _, s := range schedules {
		if s.Job.Once {
			filtered = append(filtered, s)
		}
	}

	if len(filtered) == 0 {
		return
	}

	runScheduleJobs(filtered)

	log.Debugf("Passed Schedules Completed Before %d", beforeTime)
}
