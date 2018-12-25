package spec

import "github.com/exmonitor/watcher/interval/status"

type CheckInterface interface {
	RunCheck()
	LogResult(s *status.Status)
	LogRunError(err error, message string)
}
