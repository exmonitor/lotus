package status

import (
	"github.com/exmonitor/exclient/database"
	"github.com/exmonitor/exclient/database/spec/status"
	"github.com/pkg/errors"
	"time"
)

type Status struct {
	Id        int
	ReqId     string
	Result    bool
	Error     error
	Duration  time.Duration
	Message   string
	ExtraInfo string

	DBClient database.ClientInterface
}

func NewStatus(dbClient database.ClientInterface) (*Status, error) {
	if dbClient == nil {
		return nil, errors.Wrap(invalidConfigError, "dbClient must not be nil")
	}

	newStatus := &Status{
		Result:   false,
		Error:    nil,
		DBClient: dbClient,
	}

	return newStatus, nil
}

func (s *Status) Set(result bool, err error, msg string, extraMsg string) {
	s.Result = result
	s.Error = err
	if msg != "" {
		s.Message += msg
	}
	if extraMsg != "" {
		s.ExtraInfo += extraMsg
	}
}

func (s *Status) SaveToDB() {
	serviceStatus := &status.ServiceStatus{
		Id:     s.Id,
		Result: s.Result,
		// TODO
	}

	s.DBClient.ES_SaveServiceStatus(serviceStatus)
}