package check

import (
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
)

type IntervalGroupConfig struct {
	IntervalSec        int
	FetchLoopModulator int //  how often we should fetch checks from DB in terms of loops (ie: fetch data every 10 loops)

	// db client interface
	DBClient DBInterface
}

type IntervalGroup struct {
	intervalSec        int
	fetchLoopModulator int //  how often we should fetch checks from DB in terms of loops (ie: fetch data every 10 loops)
	loopCounter        int

	// db client interface
	dBClient DBInterface
}

func NewIntervalGroup(conf IntervalGroupConfig) (*IntervalGroup, error) {
	newIG := &IntervalGroup{}
	{
		newIG.intervalSec = conf.IntervalSec
		newIG.fetchLoopModulator = conf.FetchLoopModulator
		newIG.dBClient = conf.DBClient
	}

	err := newIG.validateNewIG()
	if err != nil {
		return nil, err
	}

	return newIG, nil
}

func (ig *IntervalGroup) validateNewIG() error {
	if ig.intervalSec == 0 {
		return errors.Wrap(invalidConfigError, "IntervalGroup.intervalSec must not be zero")
	}
	if ig.intervalSec < 5 {
		return errors.Wrap(invalidConfigError, fmt.Sprintf("IntervalGroup.intervalSec %ds is too small for effective monitoring, minmum is 5s", ig.intervalSec))

	}
	if ig.fetchLoopModulator == 0 {
		ig.fetchLoopModulator = 1
	}
	return nil
}

// wrapper for running in separate thread
func (ig *IntervalGroup) Run() {
	if ig == nil {
		fmt.Printf(" base ig in intervalGroup is null wtf? \n")

	}

	ig.runLoop()
}

func (ig *IntervalGroup) runLoop() {
	var checks []CheckInterface
	filter := &CheckFilter{
		IntervalSec: ig.intervalSec,
	}
	// run tick goroutine
	tickChan := make(chan bool)
	go intervalTick(ig.intervalSec, tickChan)

	// run infinite loop
	for {
		// wait until we reached another interval tick
		select {
		case <-tickChan:
			ig.Log("received tick")
		}
		// fetch check data
		if ig.loopCounter%ig.fetchLoopModulator == 0 {
			checks = ig.dBClient.GetAllChecks(filter)
			ig.Log(fmt.Sprintf("fetched %d checks from db", len(checks)))
		}

		// run each check in separate goroutine
		for _, check := range checks {
			go check.RunCheck()
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

func (ig *IntervalGroup) Log(msg string) {
	log.Printf("INFO|INTERVAL|%ds|loop %d|%s", ig.intervalSec, ig.loopCounter, msg)
}
func (ig *IntervalGroup) LogError(msg string) {
	log.Printf("ERROR|interval 'every %ds', loop '%d'|%s", ig.intervalSec, ig.loopCounter, msg)
}

func (ig *IntervalGroup) LoopCounterInc() {
	ig.loopCounter += 1
}
