package check

import (
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

	DBClient DBInterface
}

func NewStatus(dbClient DBInterface) *Status {
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
	s.DBClient.SaveCheckStatus(s)
}
