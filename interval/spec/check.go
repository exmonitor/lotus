package spec

import (
	"github.com/exmonitor/watcher/interval/status"
)

type CheckInterface interface {
	RunCheck()
	GetStringPort() string
	LogResult(s *status.Status)
	LogRunError(err error, message string)
}
