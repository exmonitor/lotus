package http

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/giantswarm/project-lotus/lotus/check"
	"github.com/giantswarm/project-lotus/lotus/key"
)

const (
	msgFailedToExecute       = "failed to execute http request: "
	msgFailedBadStatusCode   = "failed - bad http status code"
	msgFailedContentNotFound = "failed - content not found"
	msgFailedCertExpired     = "failed - certificate expiration issue"

	msgInternalFailedToReadResponse = "INTERNAL: failed to read http response"
	msgInternalFailedHttpClient     = "INTERNAL: failed to prepare http request"
)

var defaultAllowedStatusCodes = []int{200, 201, 202, 203, 204, 205}

// config is used for initializing the check
type CheckConfig struct {
	// general options
	Id      int
	Port    int
	Target  string // IP or URL
	Timeout time.Duration

	// protocol specific options
	Method       string
	Query        string
	ExtraHeaders []HTTPHeader
	AuthEnabled  bool
	AuthUsername string
	AuthPassword string

	// content specific options
	ContentCheckEnabled bool
	ContentCheckString  string

	// allowed http responses
	AllowedHttpStatusCodes []int

	// https options
	TlsSkipVerify              bool
	TlsCheckCertificates       bool
	TlsCertExpirationThreshold time.Duration

	//db client
	DBClient check.DBInterface
}

type Check struct {
	// general options
	id        int    // id of  the check saved in db, always same for the specific check
	requestId string // identification of this current request, always unique across all data in eternity
	port      int
	target    string // IP or URL
	timeout   time.Duration

	// protocol specific options
	method       string
	query        string
	extraHeaders []HTTPHeader
	authEnabled  bool
	authUsername string
	authPassword string

	// content specific options
	contentCheckEnabled bool
	contentCheckString  string

	// allowed http responses status code (ie: [200,404])
	allowedHttpStatusCodes []int

	// https options
	tlsSkipVerify              bool
	tlsCheckCertificates       bool
	tlsCertExpirationThreshold time.Duration

	// db client
	dbClient check.DBInterface

	// internals
	check.CheckInterface
}

type HTTPHeader struct {
	Key   string
	Value string
}

func NewHttpCheck(conf CheckConfig) (*Check, error) {
	// init values
	newCheck := &Check{}
	{
		newCheck.id = conf.Id
		newCheck.port = conf.Port
		newCheck.target = conf.Target
		newCheck.timeout = conf.Timeout

		newCheck.method = conf.Method
		newCheck.query = conf.Query
		newCheck.extraHeaders = conf.ExtraHeaders
		newCheck.authEnabled = conf.AuthEnabled
		newCheck.authUsername = conf.AuthUsername
		newCheck.authPassword = conf.AuthPassword

		newCheck.contentCheckEnabled = conf.ContentCheckEnabled
		newCheck.contentCheckString = conf.ContentCheckString

		newCheck.allowedHttpStatusCodes = conf.AllowedHttpStatusCodes

		newCheck.tlsSkipVerify = conf.TlsSkipVerify
		newCheck.tlsCheckCertificates = conf.TlsCheckCertificates
		newCheck.tlsCertExpirationThreshold = conf.TlsCertExpirationThreshold

		newCheck.dbClient = conf.DBClient
	}

	err := newCheck.validateNewCheck()
	if err != nil {
		return nil, err
	}

	return newCheck, nil
}

// validate check configuration
func (c *Check) validateNewCheck() error {
	if c.id == 0 {
		return errors.Wrap(invalidConfigError, "check.Id must not be zero")
	}
	if c.port == 0 {
		return errors.Wrap(invalidConfigError, "check.Port must not be zero")
	}
	if c.target == "" {
		return errors.Wrap(invalidConfigError, "check.Target must not be empty")
	}
	if c.timeout == 0 {
		return errors.Wrap(invalidConfigError, "check.Timeout must not be zero")
	}
	if c.method == "" {
		return errors.Wrap(invalidConfigError, "check.Method must not be empty")
	}
	if c.method != http.MethodGet && c.method != http.MethodHead && c.method != http.MethodPost {
		return errors.Wrap(invalidConfigError, "http method "+c.method+" is not supported")
	}
	if c.authEnabled && c.authUsername == "" {
		return errors.Wrapf(invalidConfigError, "check.Username must not be empty, when BasicAuth is enabled")
	}
	if len(c.allowedHttpStatusCodes) == 0 {
		c.allowedHttpStatusCodes = defaultAllowedStatusCodes
	}
	if c.tlsCheckCertificates && c.tlsCertExpirationThreshold == 0 {
		return errors.Wrapf(invalidConfigError, "check.tlsCertExpirationThreshold must not be zero, when tlsCheckCertificates is enabled")
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

// run monitoring check with all options
func (c *Check) doCheck() *check.Status {
	status := check.NewStatus(c.dbClient)
	tStart := time.Now()

	// set tls config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: c.tlsSkipVerify,
	}
	// set http transport configuration
	transportConf := &http.Transport{
		ResponseHeaderTimeout: c.timeout,
		IdleConnTimeout:       c.timeout,
		TLSClientConfig:       tlsConfig,
	}
	// initialize http client
	client := http.Client{
		Transport:     transportConf,
		CheckRedirect: c.redirectPolicyFunc,
	}
	// prepare http request
	req, err := http.NewRequest(c.method, c.target+"/"+c.query, nil)
	if err != nil {
		c.LogRunError(err, msgInternalFailedHttpClient)
		status.Set(false, err, msgInternalFailedHttpClient, "")
		return status
	}
	// set basic auth if its enabled
	if c.authEnabled {
		req.SetBasicAuth(c.authUsername, c.authPassword)
	}
	// add all extra http headers
	c.addExtraHeaders(req)

	// execute http request
	resp, err := client.Do(req)
	if err != nil {
		status.Set(false, err, msgFailedToExecute, "")
		return status
	} else {
		httpCodeOK := false
		// check if http response code is allowed
		for _, allowedStatusCode := range c.allowedHttpStatusCodes {
			if resp.StatusCode == allowedStatusCode {
				httpCodeOK = true
				break
			}
		}
		if !httpCodeOK {
			msg := fmt.Sprintf("HTTP status code: %d is in within allowed codes %s", resp.StatusCode, c.allowedHttpStatusCodes)
			status.Set(false, nil, msgFailedBadStatusCode, msg)
			return status
		}

		// check for content
		if c.contentCheckEnabled && httpCodeOK {
			// read http response body
			respData, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				c.LogRunError(err, msgInternalFailedToReadResponse)
				status.Set(false, err, msgInternalFailedToReadResponse, "")
				return status
			}
			// check if response contains requested content in http body
			if !strings.Contains(string(respData), c.contentCheckString) {
				msg := "the page successfully retrieved, but the required content was not found"
				status.Set(false, nil, msgFailedContentNotFound, msg)
				return status
			}
		}
	}
	// check certificates
	if c.tlsCheckCertificates {
		certsOK, message := c.checkTLS(resp.TLS)
		if !certsOK {
			status.Set(false, nil, msgFailedCertExpired, message)
			return status
		}
	}

	status.Duration = time.Since(tStart)
	status.Set(true, nil, check.MsgSuccess, "")

	return status
}

// redirect policy, in case the target URL is not real page but is redirecting to somewhere else
// we need to re-add all the http headers
func (c *Check) redirectPolicyFunc(req *http.Request, via []*http.Request) error {
	// set basic auth if its enabled
	if c.authEnabled {
		req.SetBasicAuth(c.authUsername, c.authPassword)
	}
	// add all extra http headers
	c.addExtraHeaders(req)

	return nil
}

// add extra http headers to the request
func (c *Check) addExtraHeaders(req *http.Request) {
	// add all extra http headers
	for i := 0; i < len(c.extraHeaders); i++ {
		req.Header.Add(c.extraHeaders[i].Key, c.extraHeaders[i].Value)
	}
}

// check TTL of tls certs
func (c *Check) checkTLS(conn *tls.ConnectionState) (bool, string) {
	certsOK := true
	message := ""
	// check certs
	for _, cert := range conn.PeerCertificates {
		// check if now() + tlsExpirationThreshold > CertExpirationDate
		if time.Now().Add(c.tlsCertExpirationThreshold).After(cert.NotAfter) {
			certsOK = false
			message += fmt.Sprintf("certificate %s will expire in less than %.0f hours", cert.DNSNames, c.tlsCertExpirationThreshold.Hours())
		}
	}

	return certsOK, message
}

func (c *Check) LogResult(s *check.Status) {
	logMessage := s.Message
	if s.ExtraInfo != "" {
		logMessage += ", ExtraInfo: " + s.ExtraInfo
	}
	if s.Error != nil {
		logMessage += ", Error: " + s.Error.Error()
	}

	log.Printf("INFO|check-HTTP|id %d|reqID %s|target %s|port %d|latency %sms|result '%t'|msg: %s", c.id, c.requestId, c.target, c.port, key.MsFromDuration(s.Duration), s.Result, logMessage)
}

func (c *Check) LogRunError(err error, message string) {
	log.Printf("ERROR| running check id:%d reqID:%s type:http/https target:%s failed with error:%s, reason: %s", c.id, c.requestId, c.target, err, message)
}
