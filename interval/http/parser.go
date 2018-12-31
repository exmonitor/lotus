package http

import (
	"encoding/json"
	"fmt"
	"github.com/exmonitor/exclient/database"
	"github.com/exmonitor/exclient/database/spec/service"
	"github.com/exmonitor/exlogger"
	"github.com/pkg/errors"
	"time"
)

/*
Example metadata:
{
	"id": 1,
	"port": 443,
	"target": "test.domain.cz",
	"timeout": 5,
    "proto": "http",
	"method": "GET",
	"query": "?var1=value1&var2=value2",
	"postData": [
		{
			"name": "var1",
			"value": "value1"
		}
	],
	"extraHeaders": [
		{
			"name": "MyHeader",
			"value": "My Value"
		}
	],
	"authEnabled": true,
	"authUsername": "admin",
	"authPassword": "adminPass",
	"contentCheckEnabled": true,
	"contentCheckString": "my_string",
	"allowedHttpStatusCodes": [
		200,
		201,
		403,
		404
	],
	"tlsSkipVerify": false,
	"tlsCheckCertificates": true,
	"tlsCertExpirationThreshold": 10,
}
*/

type RawCheck struct {
	Id                         int            `json:"id"`
	Port                       int            `json:"port"`
	Target                     string         `json:"target"`
	Timeout                    int            `json:"timeout"`
	Proto                      string         `json:"proto"`
	Method                     string         `json:"method"`
	Query                      string         `json:"query"`
	PostData                   []HTTPKeyValue `json:"postData"`
	ExtraHeaders               []HTTPKeyValue `json:"extraHeaders"`
	AuthEnabled                bool           `json:"authEnabled"`
	AuthUsername               string         `json:"authUsername"`
	AuthPassword               string         `json:"authPassword"`
	ContentCheckEnabled        bool           `json:"contentCheckEnabled"`
	ContentCheckString         string         `json:"contentCheckString"`
	AllowedHttpStatusCodes     []int          `json:"allowedHttpStatusCodes"`
	TlsSkipVerify              bool           `json:"tlsSkipVerify"`
	TlsCheckCertificates       bool           `json:"tlsCheckCertificates"`
	TlsCertExpirationThreshold int            `json:"tlsCertExpirationThreshold"`
}

func ParseCheck(service *service.Service, dbClient database.ClientInterface, logger *exlogger.Logger) (*Check, error) {
	var rawCheck RawCheck
	err := json.Unmarshal([]byte(service.Metadata), &rawCheck)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to parse ICMP json metadata for check id %d", service.ID))
	} else {
		logger.LogDebug("Successfully parsed HTTP json metadata for check id %d", service.ID)
	}

	checkConfig := CheckConfig{
		Id:                         service.ID,
		FailThreshold:              service.FailThreshold,
		Interval:                   service.Interval,
		Port:                       rawCheck.Port,
		Target:                     rawCheck.Target,
		Timeout:                    time.Second * time.Duration(rawCheck.Timeout),
		Proto:                      rawCheck.Proto,
		Method:                     rawCheck.Method,
		Query:                      rawCheck.Query,
		PostData:                   rawCheck.PostData,
		ExtraHeaders:               rawCheck.ExtraHeaders,
		AuthEnabled:                rawCheck.AuthEnabled,
		AuthUsername:               rawCheck.AuthUsername,
		AuthPassword:               rawCheck.AuthPassword,
		ContentCheckEnabled:        rawCheck.ContentCheckEnabled,
		ContentCheckString:         rawCheck.ContentCheckString,
		AllowedHttpStatusCodes:     rawCheck.AllowedHttpStatusCodes,
		TlsSkipVerify:              rawCheck.TlsSkipVerify,
		TlsCheckCertificates:       rawCheck.TlsCheckCertificates,
		TlsCertExpirationThreshold: time.Hour * 24 * time.Duration(rawCheck.TlsCertExpirationThreshold), // convert to days

		Logger:   logger,
		DBClient: dbClient,
	}

	return New(checkConfig)
}
