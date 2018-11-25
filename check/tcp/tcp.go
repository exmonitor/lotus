package tcp

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/pkg/errors"

	"github.com/giantswarm/project-lotus/lotus/check"
	"github.com/giantswarm/project-lotus/lotus/key"
)

type CheckConfig struct {
	Id      int
	ReqId   string
	Target  string
	Port    int
	Timeout time.Duration

	//db client
	DBClient check.DBInterface
}

type Check struct {
	id        int
	requestId string
	target    string
	port      int
	timeout   time.Duration

	// db client
	dbClient check.DBInterface

	// internals
	check.CheckInterface
}

func NewCheck(conf CheckConfig) (*Check, error) {
	newCheck := &Check{}
	{
		newCheck.id = conf.Id
		newCheck.timeout = conf.Timeout
		newCheck.port = conf.Port
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
	if c.port == 0 {
		return errors.Wrap(invalidConfigError, "check.Port must not be zero")
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

	conn, err := net.DialTimeout("tcp", tcpTargetAddress(c.target, c.port), c.timeout)
	if err != nil {
		status.Set(false, err, "failed to open tcp connection", "")
		return status
	} else {
		defer conn.Close()
		//if _, err := fmt.Fprintf(conn, testMsg); err != nil {
		//	t.Fatal(err)
		//}
		status.Set(true, nil, check.MsgSuccess, "")
	}

	status.Duration = time.Since(tStart)
	return status
}

func tcpTargetAddress(target string, port int) string {
	return fmt.Sprintf("%s:%d", target, port)
}

func (c *Check) LogResult(s *check.Status) {
	logMessage := s.Message
	if s.ExtraInfo != "" {
		logMessage += ", ExtraInfo: " + s.ExtraInfo
	}
	if s.Error != nil {
		logMessage += ", Error: " + s.Error.Error()
	}
	log.Printf("INFO|check-TCP|id %d|reqID %s|target %s|port %d|latency %sms|result '%t'|msg: %s", c.id, c.requestId, c.target, c.port, key.MsFromDuration(s.Duration), s.Result, logMessage)
}

func (c *Check) LogRunError(err error, message string) {
	log.Printf("ERROR| running check id:%d reqID:%s type:tcp target:%s%d failed with error:%s, reason: %s", c.id, c.requestId, c.target, c.port, err, message)
}
