package icmp

import (
	"log"
	"time"

	"github.com/pkg/errors"
	"github.com/sparrc/go-ping"

	"github.com/giantswarm/project-lotus/lotus/check"
	"github.com/giantswarm/project-lotus/lotus/key"
)

const (
	pingCount  = 1
	MsgSuccess = "success"
	MsgTimeout = "failed - timeout"

	msgInternalFailedToInitialisePing = "failed to initialise pinger"
)

type CheckConfig struct {
	Id      int
	ReqId   string
	Target  string
	Timeout time.Duration

	//db client
	DBClient check.DBInterface
}

type Check struct {
	id        int
	requestId string
	target    string
	timeout   time.Duration

	// db client
	dbClient check.DBInterface

	//internals
	check.CheckInterface
}

func NewCheck(conf CheckConfig) (*Check, error) {
	newCheck := &Check{}
	{
		newCheck.id = conf.Id
		newCheck.timeout = conf.Timeout
		newCheck.target = conf.Target

		newCheck.dbClient = conf.DBClient
	}

	err := newCheck.validateNewCheck()
	if err != nil {
		return nil, err
	}
	return newCheck, nil
}

func (c *Check) validateNewCheck() error {
	if c.id == 0 {
		return errors.Wrap(invalidConfigError, "check.Id must not be zero")
	}
	if c.target == "" {
		return errors.Wrap(invalidConfigError, "check.Target must not be empty")
	}
	if c.timeout == 0 {
		return errors.Wrap(invalidConfigError, "check.Timeout must not be zero")
	}
	if c.dbClient == nil {
		return errors.Wrap(invalidConfigError, "check.DbClient must not be nil")
	}

	return nil
}

// wrapper function used to run in separate thread (goroutine)
func (c *Check) RunCheck() {
	// generate unique request ID
	c.requestId = key.GenerateReqId(c.id)
	// run monitoring check
	status := c.doCheck()
	c.LogResult(status)

	// save result to database
	status.SaveToDB()
}

func (c *Check) doCheck() *check.Status {
	status := check.NewStatus(c.dbClient)
	tStart := time.Now()

	pinger, err := ping.NewPinger(c.target)
	{
		if err != nil {
			c.LogRunError(err, msgInternalFailedToInitialisePing)
			status.Set(false, err, msgInternalFailedToInitialisePing, "")
			return status
		}

		pinger.Count = pingCount
		pinger.Timeout = c.timeout
		pinger.SetPrivileged(true)
		pinger.OnRecv = func(pkt *ping.Packet) {
			// we got positive response
			status.Set(true, nil, MsgSuccess, "")
		}
	}

	pinger.Run()
	status.Duration = time.Since(tStart)
	if status.Duration >= c.timeout {
		status.Set(false, nil, MsgTimeout, "")
	}

	return status
}

func (c *Check) LogResult(s *check.Status) {
	logMessage := s.Message
	if s.ExtraInfo != "" {
		logMessage += ", ExtraInfo: " + s.ExtraInfo
	}
	if s.Error != nil {
		logMessage += ", Error: " + s.Error.Error()
	}

	log.Printf("INFO|check-ICMP|id %d|reqID %s|target %s|latency %sms|result '%t'|msg: %s", c.id, c.requestId, c.target, key.MsFromDuration(s.Duration), s.Result, logMessage)
}

func (c *Check) LogRunError(err error, message string) {
	log.Printf("ERROR|CHECK|id %d|reqID %s|type ICMP|target %s|failed :%s, reason: %s", c.id, c.requestId, c.target, err, message)
}
