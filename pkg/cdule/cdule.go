package cdule

import (
	"os"
	"time"

	"github.com/gagasdiv/cdule/pkg"
	"github.com/gagasdiv/cdule/pkg/model"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// WorkerID string
var WorkerID string

// Cdule holds watcher objects
type Cdule struct {
	*WorkerWatcher
	*ScheduleWatcher
	PastScheduleWatcher *PastScheduleWatcher
}

func init() {
	WorkerID = getWorkerID()
}

// NewCduleWithWorker to create new scheduler with worker
func (cdule *Cdule) NewCduleWithWorker(workerName string, config ...*pkg.CduleConfig) {
	WorkerID = workerName
	cdule.NewCdule(config...)
}

// NewCdule to create new scheduler with default worker name as hostname
func (cdule *Cdule) NewCdule(config ...*pkg.CduleConfig) {
	cfg := pkg.ResolveConfig(config...)

	model.ConnectDataBase(cfg)
	worker, err := model.CduleRepos.CduleRepository.GetWorker(WorkerID)
	if nil != err {
		log.Errorf("Error getting worker %s ", err.Error())
		return
	}
	if nil != worker {
		worker.UpdatedAt = time.Now()
		model.CduleRepos.CduleRepository.UpdateWorker(worker)
	} else {
		// First time cdule started on a worker node
		worker := model.Worker{
			WorkerID:  WorkerID,
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
			DeletedAt: gorm.DeletedAt{},
		}
		model.CduleRepos.CduleRepository.CreateWorker(&worker)
	}

	cdule.createWatcherAndWaitForSignal(cfg)
}

func (cdule *Cdule) createWatcherAndWaitForSignal(config *pkg.CduleConfig) {
	/*
		schedule watcher stop logic to abort program with signal like ctrl + c
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt)*/

	workerWatcher := createWorkerWatcher()
	schedulerWatcher := createSchedulerWatcher(config)
	cdule.WorkerWatcher = workerWatcher
	cdule.ScheduleWatcher = schedulerWatcher

	var pastScheduleWatcher *PastScheduleWatcher
	if config.WatchPast {
		pastScheduleWatcher = createPastSchedulerWatcher(config)
		cdule.PastScheduleWatcher = pastScheduleWatcher
	}
	/*select {
	case sig := <-c:
		fmt.Printf("Received %s signal. Aborting...\n", sig)
		workerWatcher.Stop()
		schedulerWatcher.Stop()
		if pastScheduleWatcher != nil {
			pastScheduleWatcher.Stop()
		}
	}*/
}

// StopWatcher to stop watchers
func (cdule Cdule) StopWatcher() {
	cdule.WorkerWatcher.Stop()
	cdule.ScheduleWatcher.Stop()

	if cdule.PastScheduleWatcher != nil {
		cdule.PastScheduleWatcher.Stop()
	}
}
func createWorkerWatcher() *WorkerWatcher {
	workerWatcher := &WorkerWatcher{
		Closed: make(chan struct{}),
		Ticker: time.NewTicker(time.Second * 30), // used for worker health check update in db.
	}

	workerWatcher.WG.Add(1)
	go func() {
		defer workerWatcher.WG.Done()
		workerWatcher.Run()
	}()
	return workerWatcher
}

func createSchedulerWatcher(config *pkg.CduleConfig) *ScheduleWatcher {
	tick, err := time.ParseDuration(config.TickDuration)
	if err != nil {
		panic(err)
	}
	scheduleWatcher := &ScheduleWatcher{
		Closed: make(chan struct{}),
		TickDuration: tick,
		Ticker: time.NewTicker(tick),
		RunImmediately: config.RunImmediately,
	}

	scheduleWatcher.WG.Add(1)
	go func() {
		defer scheduleWatcher.WG.Done()
		scheduleWatcher.Run()
	}()
	return scheduleWatcher
}

func createPastSchedulerWatcher(config *pkg.CduleConfig) *PastScheduleWatcher {
	tick, err := time.ParseDuration(config.TickDuration)
	if err != nil {
		panic(err)
	}
	pastScheduleWatcher := &PastScheduleWatcher{
		ScheduleWatcher: ScheduleWatcher{
			TickDuration: tick,
			Ticker: time.NewTicker(tick),
			RunImmediately: config.RunImmediately,
		},
	}

	pastScheduleWatcher.WG.Add(1)
	go func() {
		defer pastScheduleWatcher.WG.Done()
		pastScheduleWatcher.Run()
	}()
	return pastScheduleWatcher
}

func getWorkerID() string {
	hostname, err := os.Hostname()
	if err != nil {
		os.Exit(1)
	}
	return hostname
}
