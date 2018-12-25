package spec

import "github.com/exmonitor/watcher/service/status"

type CheckInterface interface {
	RunCheck()
	LogResult(s *status.Status)
	LogRunError(err error, message string)
}
