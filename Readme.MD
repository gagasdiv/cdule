# cdule (pronounce as Schedule)
![dbschema.png](pkg/doc/cdulelogo.png)

Golang based scheduler library with database support.
Users could use any database which is supported by [gorm.io](https://gorm.io/).

### To Download the cdule library
```
go get github.com/deepaksinghvi/cdule
```

### Usage Instruction

In order to schedule jobs with cdule, user needs to
1. Configure persistence
2. Implement cdule.Job Interface &
3. Schedule job with required cron expression.

Job will be persisted in the jobs table.\
Next execution would be persisted in schedules tables.\
Job history would be persisted and maintained in job_histories table.


## Configuration

This fork differs from the original repo in that the responsibility of making the config object is shifted to the caller script; which means cdule's config can be moved/merged/integrated with already existing configuration storage in a project. The classic `.yml` file (or whichever config system you use) can still be used, but again, the responsibility is yours to form the config object.

I forked this package originally because I had all my config in an existing `.env` file.

The following are the config keys:

| Key | Description |
| ----------- | ----------- |
| `Cduletype` | Specify whether it is an In-Memory or Database based configuration. Possible values are `"DATABASE"` and `"MEMORY"`. |
| `Dburl` | The database connection url. For `DATABASE` the supported ones are `postgres` and `mysql`; `MEMORY` will use `sqlite`. |
| `Cduleconsistency` | Reserved for future usage. |
| `Loglevel` | The log level to give `gorm`. |


### Example configuration values:
```go
&CduleConfig{
	Cduletype:        "DATABASE",
	Dburl:            "postgres://user:pass@localhost:5432/cdule?sslmode=disable",
	Cduleconsistency: "AT_MOST_ONCE",
	Loglevel:         logger.Error,
}
```


## Job Interface Implementation

```
var testJobData map[string]string

type TestJob struct {
	Job cdule.Job
}

func (m TestJob) Execute(jobData map[string]string) {
	log.Debug("In TestJob")
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

func (m TestJob) JobName() string {
	return "job.TestJob"
}

func (m TestJob) GetJobData() map[string]string {
	return testJobData
}
```

## Schedule a Job
It is expected that testJob will be Executed five times, once for every minute and program will exit. TestJob jobData map holds the data in the format of map[string]string where gets stored for every execution and gets updated as the next counter value on Execute() method call.

```
cdule := cdule.Cdule{}
cdule.NewCdule()
testJob := TestJob{}
jobData := make(map[string]string)
jobData["one"] = "1"
jobData["two"] = "2"
jobData["three"] = "3"
cdule.NewJob(&testJob, jobData).Build(utils.EveryMinute)

time.Sleep(5 * time.Minute)
cdule.StopWatcher()
```


### Demo Project
This demo describes how cdule library can be used.

https://github.com/deepaksinghvi/cduledemo


### Database Schema
For the sample app, postgresql is used but users can use any db which is supported by gorm.io.
#### DB Tables
* jobs : To store unique jobs.
* job_histories : To store job history with status as result.
* schedules : To store schedule for every next run.
* workers : To store the worker nodes and their health check.


![dbschema.png](pkg/doc/dbschema.png)


### Sample Cron
Users can use the pre-defined crons or use their own which are the [standard cron](https://en.wikipedia.org/wiki/Cron) expression
```
EveryMinute              = "0 * * ? * *"
EveryEvenMinute          = "0 */2 * ? * *"
EveryUnEvenMinute        = "0 1/2 * ? * *"
EveryTwoMinutes          = "0 */2 * ? * *"
EveryHourAtMin153045     = "0 15,30,45 * ? * *"
EveryHour                = "0 0 * ? * *"
EveryEvenHour            = "0 0 0/2 ? * *"
EveryUnEvenHour          = "0 0 1/2 ? * *"
EveryThreeHours          = "0 0 */3 ? * *"
EveryTwelveHours         = "0 0 */12 ? * *"
EveryDayAtMidNight       = "0 0 0 * * ?"
EveryDayAtOneAM          = "0 0 1 * * ?"
EveryDayAtSixAM          = "0 0 6 * * ?"
EverySundayAtNoon        = "0 0 12 ? * "
EveryMondayAtNoon        = "0 0 12 ? *"
EveryWeekDayAtNoon       = "0 0 12 ? * MON-FRI"
EveryWeekEndAtNoon       = "0 0 12 ? * SUN,SAT"
EveryMonthOnFirstAtNoon  = "0 0 12 1 * ?"
EveryMonthOnSecondAtNoon = "0 0 12 2 * ?"
```


### This library is built using

* Cron parser using the library [robfig/cron](https://github.com/robfig/cron)
* ORM usign the library [gorm.io](https://gorm.io/)
