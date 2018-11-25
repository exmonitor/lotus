package check

import "log"

const (
	MsgSuccess = "success"
	MsgTimeout = "failed - timeout"
)

// default array of check intervals
// each value represents the interval in seconds
var DefaultCheckIntervals = []int{10, 30, 60, 120, 300, 600}

type CheckInterface interface {
	RunCheck()
	LogResult(s *Status)
	LogRunError(err error, message string)
}

// interface that each db driver needs to implement so we can easily use them
type DBInterface interface {
	GetAllChecks(filter *CheckFilter) []CheckInterface
	SaveCheckStatus(status *Status) error
}

// used for querying database for all checks that match the filter
type CheckFilter struct {
	IntervalSec int
}

func InitIntervalGroups(intervalGroups []int, dbClient DBInterface) {
	// TODO
	// for now use predefined intervals
	if intervalGroups == nil {
		intervalGroups = DefaultCheckIntervals
	}
	// iterate over all intervals and create thread for each group
	for _, interval := range intervalGroups {
		igConfig := IntervalGroupConfig{
			IntervalSec: interval,
			DBClient:    dbClient,
		}

		ig, err := NewIntervalGroup(igConfig)
		if err != nil {
			log.Printf("ERROR| failed to intialise IntervalGroup 'every %ds' error: %s", interval, err)
			continue
		}

		// run each interval group in separate thread
		go ig.Run()
	}
}
