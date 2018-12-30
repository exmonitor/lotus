package icmp

import (
	"time"

	"github.com/sparrc/go-ping"

	"fmt"
	"github.com/exmonitor/exclient/database"
	"github.com/exmonitor/exlogger"
	"github.com/exmonitor/watcher/interval/spec"
	"github.com/exmonitor/watcher/interval/status"
	"github.com/exmonitor/watcher/key"
	"github.com/pkg/errors"
)

const (
	pingCount  = 1
	MsgSuccess = "success"
	MsgTimeout = "failed - timeout"

	msgInternalFailedToInitialisePing = "failed to initialise pinger"
)

type CheckConfig struct {
	Id            int
	FailThreshold int
	Target        string
	Timeout       time.Duration

	//db client
	DBClient database.ClientInterface
	Logger   *exlogger.Logger
}

type Check struct {
	id            int
	failThreshold int
	requestId     string
	target        string
	timeout       time.Duration

	// db client
	dbClient database.ClientInterface
	// logger
	log *exlogger.Logger

	//internals
	spec.CheckInterface
}

func NewCheck(conf CheckConfig) (*Check, error) {
	if conf.Id == 0 {
		return nil, errors.Wrap(invalidConfigError, "check.id must not be zero")
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
		id:            conf.Id,
		failThreshold: conf.FailThreshold,
		timeout:       conf.Timeout,
		target:        conf.Target,

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
	statusConfig := status.Config{
		Id:            c.id,
		ReqId:         c.requestId,
		FailThreshold: c.failThreshold,
		DBClient:      c.dbClient,
	}
	s, err := status.New(statusConfig)
	if err != nil {
		c.LogRunError(err, fmt.Sprintf("failed to init new status for ICMP service ID %d", c.id))
	}
	tStart := time.Now()

	pinger, err := ping.NewPinger(c.target)
	{
		if err != nil {
			c.LogRunError(err, msgInternalFailedToInitialisePing)
			s.Set(false, err, msgInternalFailedToInitialisePing)
			return s
		}

		pinger.Count = pingCount
		pinger.Timeout = c.timeout
		pinger.SetPrivileged(true)
		pinger.OnRecv = func(pkt *ping.Packet) {
			// we got positive response
			s.Set(true, nil, MsgSuccess)
		}
	}

	pinger.Run()
	s.Duration = time.Since(tStart)
	if s.Duration >= c.timeout {
		s.Set(false, nil, MsgTimeout)
	}

	return s
}

func (c *Check) LogResult(s *status.Status) {
	c.log.Log("check-ICMP|id %d|reqID %s|target %s|latency %sms|result '%t'|msg: %s", c.id, c.requestId, c.target, key.MsFromDuration(s.Duration), s.Result, s.Message)
}

func (c *Check) LogRunError(err error, message string) {
	c.log.LogError(err, "CHECK|id %d|reqID %s|type ICMP|target %s| reason: %s", c.id, c.requestId, c.target, message)
}
