package status

import (
	"github.com/exmonitor/exclient/database"
	"github.com/exmonitor/exclient/database/spec/status"
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

func NewStatus(dbClient database.ClientInterface) *Status {
	return &Status{
		Result:   false,
		Error:    nil,
		DBClient: dbClient,
	}
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
