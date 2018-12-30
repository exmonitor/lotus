package status

import (
	"github.com/exmonitor/exclient/database"
	"github.com/exmonitor/exclient/database/spec/status"
	"github.com/pkg/errors"
	"time"
)

type Config struct {
	Id    int
	ReqId string

	// extra
	FailThreshold int
	// db client
	DBClient database.ClientInterface
}

type Status struct {
	id       int
	reqId    string
	Result   bool
	Duration time.Duration
	Message  string

	// extra
	failThreshold int
	// db client
	dbClient database.ClientInterface
}

func New(conf Config) (*Status, error) {
	if conf.Id == 0 {
		return nil, errors.Wrap(invalidConfigError, "conf.Id must not be zero")
	}
	if conf.DBClient == nil {
		return nil, errors.Wrap(invalidConfigError, "conf.DBClient must not be nil")
	}
	if conf.FailThreshold == 0 {
		return nil, errors.Wrap(invalidConfigError, "conf.FailThreshold must not be zero")
	}
	if conf.ReqId == "" {
		return nil, errors.Wrap(invalidConfigError, "conf.ReqId must not be empty")
	}

	newStatus := &Status{
		id:     conf.Id,
		reqId:  conf.ReqId,
		Result: false,

		failThreshold: conf.FailThreshold,
		dbClient:      conf.DBClient,
	}

	return newStatus, nil
}
// set the result info
func (s *Status) Set(result bool, err error, msg string) {
	s.Result = result
	if msg != "" {
		s.Message += msg
	}
	if err != nil {
		s.Message += " | error:" + err.Error()
	}
}

func (s *Status) SaveToDB() {
	// init  db structure for saving data about status
	serviceStatus := &status.ServiceStatus{
		Id:            s.id,
		FailThreshold: s.failThreshold,
		Result:        s.Result,
		Duration:      s.Duration,
		ReqId:         s.reqId,
		Message:       s.Message,
		// timestamp for the record
		InsertTimestamp: time.Now(),
	}
	// save to db via Elasticsearch client
	s.dbClient.ES_SaveServiceStatus(serviceStatus)
}
