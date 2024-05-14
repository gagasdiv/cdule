package main

import (
	"strconv"
	"time"

	"github.com/gagasdiv/cdule/pkg"
	"github.com/gagasdiv/cdule/pkg/cdule"

	log "github.com/sirupsen/logrus"
)

/*
TODO This TestJob schedule in main program is for the development debugging.
*/

func main() {
	c := cdule.Cdule{}
	c.NewCduleWithWorker("worker1", &pkg.CduleConfig{
		RunImmediately:    true,
		Cduletype:        "DATABASE",
		Dburl:            "postgres://postgres:postgres@localhost:5432/qweqwe?sslmode=disable",
		Cduleconsistency: "AT_MOST_ONCE",
		Loglevel:         0,
		WatchPast:        true,
		// TablePrefix:      "cdule_",
	})

	log.SetLevel(log.DebugLevel)

	myJob := &TestJob{
		// Name: random.RandNumeralOrLetter(16),
		Name: "TestJob",
	}
	jobData := make(map[string]string)
	jobData["one"] = "1"
	jobData["two"] = "2"
	jobData["three"] = "3"
	// runAt, _ := time.Parse("2006-01-02 15:04:05-07:00", "2024-04-29 17:20:00+07:00")
	// _, err := cdule.NewJob(myJob, jobData).BuildToRunAt(runAt)
	// _, err := cdule.NewJob(myJob, jobData).BuildToRunIn(59 * time.Second)
	_, err := cdule.NewJob(myJob, jobData).BuildToRunNow()
	// time.Sleep(58 * time.Second)
	// cdule.CancelJob("TestJob", "")
	if nil != err {
		log.Error(err)
	}
	time.Sleep(3 * time.Minute)
	c.StopWatcher()
}

var testJobData map[string]string

type TestJob struct {
	Job cdule.Job
	Name string
}

func (m *TestJob) Execute(jobData map[string]string) {
	log.Debug("Execute(): In TestJob")
	println("AAAAAAAAA mgwrngp wrnmg wrmgpmwr pgmpwor mgpowrm gpowrmg wor gmwro gmwA AAAAAAAAAAAAAAAA")
	for k, v := range jobData {
		valNum, err := strconv.Atoi(v)
		if nil == err {
			jobData[k] = strconv.Itoa(valNum + 1)
		} else {
			log.Error(err)
		}

	}
	testJobData = jobData
}

func (m *TestJob) JobName() string {
	return m.Name
}

func (m *TestJob) GetJobData() map[string]string {
	return testJobData
}
