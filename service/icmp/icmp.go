package icmp

import (
	"time"

	"github.com/sparrc/go-ping"

	"github.com/exmonitor/exclient/database"
	"github.com/exmonitor/exlogger"
	"github.com/exmonitor/watcher/key"
	"github.com/exmonitor/watcher/service/spec"
	"github.com/exmonitor/watcher/service/status"
	"github.com/pkg/errors"
)

const (
	pingCount  = 1
	MsgSuccess = "success"
	MsgTimeout = "failed - timeout"

	msgInternalFailedToInitialisePing = "failed to initialise pinger"
)

type CheckConfig struct {
	Id      int
	Target  string
	Timeout time.Duration

	//db client
	DBClient database.ClientInterface
	Logger   *exlogger.Logger
}

type Check struct {
	id        int
	requestId string
	target    string
	timeout   time.Duration

	// db client
	dbClient database.ClientInterface
	// logger
	log *exlogger.Logger

	//internals
	spec.CheckInterface
}

func NewCheck(conf CheckConfig) (*Check, error) {
	if conf.Id == 0 {
		return nil, errors.Wrap(invalidConfigError, "check.Id must not be zero")
	}
	if conf.Target == "" {
		return nil, errors.Wrap(invalidConfigError, "check.Target must not be empty")
	}
	if conf.Timeout == 0 {
		return nil, errors.Wrap(invalidConfigError, "check.Timeout must not be zero")
	}
	if conf.DBClient == nil {
		return nil, errors.Wrap(invalidConfigError, "check.DbClient must not be nil")
	}

	newCheck := &Check{
		id:      conf.Id,
		timeout: conf.Timeout,
		target:  conf.Target,

		dbClient: conf.DBClient,
		log:      conf.Logger,
	}

	return newCheck, nil
}

// wrapper function used to run in separate thread (goroutine)
func (c *Check) RunCheck() {
	// generate unique request ID
	c.requestId = key.GenerateReqId(c.id)
	// run monitoring check
	s := c.doCheck()
	c.LogResult(s)

	// save result to database
	s.SaveToDB()
}

func (c *Check) doCheck() *status.Status {
	s := status.NewStatus(c.dbClient)
	tStart := time.Now()

	pinger, err := ping.NewPinger(c.target)
	{
		if err != nil {
			c.LogRunError(err, msgInternalFailedToInitialisePing)
			s.Set(false, err, msgInternalFailedToInitialisePing, "")
			return s
		}

		pinger.Count = pingCount
		pinger.Timeout = c.timeout
		pinger.SetPrivileged(true)
		pinger.OnRecv = func(pkt *ping.Packet) {
			// we got positive response
			s.Set(true, nil, MsgSuccess, "")
		}
	}

	pinger.Run()
	s.Duration = time.Since(tStart)
	if s.Duration >= c.timeout {
		s.Set(false, nil, MsgTimeout, "")
	}

	return s
}

func (c *Check) LogResult(s *status.Status) {
	logMessage := s.Message
	if s.ExtraInfo != "" {
		logMessage += ", ExtraInfo: " + s.ExtraInfo
	}
	if s.Error != nil {
		logMessage += ", Error: " + s.Error.Error()
	}

	c.log.Log("check-ICMP|id %d|reqID %s|target %s|latency %sms|result '%t'|msg: %s", c.id, c.requestId, c.target, key.MsFromDuration(s.Duration), s.Result, logMessage)
}

func (c *Check) LogRunError(err error, message string) {
	c.log.LogError(err, "CHECK|id %d|reqID %s|type ICMP|target %s| reason: %s", c.id, c.requestId, c.target, message)
}
