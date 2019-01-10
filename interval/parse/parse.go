package parse

import (
	"github.com/exmonitor/exclient/database"
	"github.com/exmonitor/exclient/database/spec/service"
	"github.com/exmonitor/exlogger"

	"github.com/exmonitor/watcher/interval/http"
	"github.com/exmonitor/watcher/interval/icmp"
	"github.com/exmonitor/watcher/interval/tcp"
	"github.com/exmonitor/watcher/key"
	"github.com/exmonitor/watcher/interval/spec"
)

func ParseCheck(s *service.Service, dbClient database.ClientInterface, logger *exlogger.Logger) (spec.CheckInterface, error) {
	var check spec.CheckInterface
	var err error
	switch s.Type {
	case key.ServiceTypeHttp:
		check, err = http.ParseCheck(s, dbClient, logger)
		break
	case key.ServiceTypeTcp:
		check, err = tcp.ParseCheck(s, dbClient, logger)
		break
	case key.ServiceTypeIcmp:
		check, err = icmp.ParseCheck(s, dbClient, logger)
		break
	}

	return check, err
}
