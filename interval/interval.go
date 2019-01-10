package interval

import (
	"fmt"
	"time"

	"github.com/exmonitor/exclient/database"
	"github.com/exmonitor/exclient/database/spec/service"

	"github.com/exmonitor/exlogger"
	"github.com/pkg/errors"
	"github.com/exmonitor/watcher/interval/parse"
)

type IntervalGroupConfig struct {
	IntervalSec        int
	Logger             *exlogger.Logger
	FetchLoopModulator int //  how often we should fetch checks from DB in terms of loops (ie: fetch data every 10 loops)

	// db client interface
	DBClient database.ClientInterface
}

type IntervalGroup struct {
	intervalSec        int
	logger             *exlogger.Logger
	loopCounter        int
	fetchLoopModulator int //  how often we should fetch checks from DB in terms of loops (ie: fetch data every 10 loops)

	// db client interface
	dbClient database.ClientInterface
}

func NewIntervalGroup(conf IntervalGroupConfig) (*IntervalGroup, error) {
	if conf.IntervalSec == 0 {
		return nil, errors.Wrap(invalidConfigError, "conf.intervalSec must not be zero")
	}
	if conf.IntervalSec < 5 {
		return nil, errors.Wrap(invalidConfigError, fmt.Sprintf("conf.intervalSec %ds is too small for effective monitoring, minmum is 5s", conf.IntervalSec))
	}
	if conf.Logger == nil {
		return nil, errors.Wrap(invalidConfigError, "conf.Logger must not be nil")
	}
	if conf.DBClient == nil {
		return nil, errors.Wrap(invalidConfigError, "conf.DBClient must not be nil")
	}
	if conf.FetchLoopModulator == 0 {
		conf.FetchLoopModulator = 1
	}

	newIG := &IntervalGroup{
		intervalSec:        conf.IntervalSec,
		logger:             conf.Logger,
		fetchLoopModulator: conf.FetchLoopModulator,
		dbClient:           conf.DBClient,
	}

	return newIG, nil
}

// wrapper for running in separate thread
func (ig *IntervalGroup) Boot() {
	var services []*service.Service
	var err error

	// run tick goroutine
	tickChan := make(chan bool)
	go intervalTick(ig.intervalSec, tickChan)

	// run infinite loop
	for {
		// wait until we reached another interval tick
		select {
		case <-tickChan:
		}
		// fetch service data
		if ig.loopCounter%ig.fetchLoopModulator == 0 {
			services, err = ig.dbClient.SQL_GetServices(ig.intervalSec)
			if err != nil {
				ig.logger.LogError(err, "failed to fetch services for interval %d", ig.intervalSec)
			}
			ig.logger.Log("fetched %d services from db for interval %d", len(services), ig.intervalSec)

		}

		// parse metada and than run each service in separate goroutine
		for _, s := range services {
			// TODO caching of already loaded services, we can introduce md5 of metadata to check if there is any change
			// TODO parsing each time is quite time expensive

			check, err := parse.ParseCheck(s, ig.dbClient, ig.logger)
			if err != nil {
				ig.logger.LogError(err, "failed to parse service type %s", s.ServiceTypeString())
			} else {
				go check.RunCheck()
			}
		}
		ig.LoopCounterInc()
	}
}

// returns true if its time to run the interval
func intervalTick(intervalSec int, tickChan chan bool) bool {
	for {
		// extract amount of second and minutes from the now time
		_, min, sec := time.Now().Clock()
		// get sum of total secs in hour as intervals can be bigger than 59 sec
		totalSeconds := min*60 + sec

		// check if we hit the interval
		if totalSeconds%intervalSec == 0 {
			// send msg to the channel that we got tick
			tickChan <- true
			time.Sleep(time.Second)
		}
		//  this is rough value, so we are testing 10 times per sec to not have big offset
		time.Sleep(time.Millisecond * 100)
	}
}

func (ig *IntervalGroup) LoopCounterInc() {
	ig.loopCounter += 1
}

const (
	MsgSuccess = "success"
	MsgTimeout = "failed - timeout"
)

// default array of check intervals
// each value represents the interval in seconds
var DefaultCheckIntervals = []int{10, 30, 60, 120, 300, 600}

func InitIntervalGroups(intervalGroups []int, dbClient database.ClientInterface, logger *exlogger.Logger) {
	// TODO
	// for now use predefined intervals
	if intervalGroups == nil {
		intervalGroups = DefaultCheckIntervals
	}
	// iterate over all intervals and create thread for each group
	for _, interval := range intervalGroups {
		igConfig := IntervalGroupConfig{
			IntervalSec: interval,
			Logger:      logger,
			DBClient:    dbClient,
		}

		ig, err := NewIntervalGroup(igConfig)
		if err != nil {
			logger.LogError(err, "failed to initialise IntervalGroup 'every %ds' ", interval)
			continue
		}

		// run each interval group in separate thread
		go ig.Boot()
	}
}
