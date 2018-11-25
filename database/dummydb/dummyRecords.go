package dummydb

import (
	"fmt"
	"github.com/giantswarm/project-lotus/lotus/check"
	"github.com/giantswarm/project-lotus/lotus/check/http"
	"github.com/giantswarm/project-lotus/lotus/check/icmp"
	"github.com/giantswarm/project-lotus/lotus/check/tcp"
	"time"
)

func getTestHttpChecks(dbClient check.DBInterface) (array []check.CheckInterface) {
	httpConfig1 := http.CheckConfig{
		Id:      1,
		Port:    80,
		Target:  "http://seznam.cz",
		Timeout: time.Second * 30,

		ExtraHeaders:           []http.HTTPHeader{},
		Query:                  "",
		Method:                 "GET",
		AllowedHttpStatusCodes: []int{200, 201, 202, 203, 204, 205, 206, 207, 208, 209},

		AuthEnabled:          false,
		ContentCheckEnabled:  false,
		TlsCheckCertificates: false,

		DBClient: dbClient,
	}
	http1, err := http.NewHttpCheck(httpConfig1)
	if err != nil {
		fmt.Printf("DUMMYDB: failed to init httpcheck id:%d, error: %s  \n", httpConfig1.Id, err)
	}

	httpConfig2 := http.CheckConfig{
		Id:      2,
		Port:    443,
		Target:  "https://google.cz",
		Timeout: time.Second * 30,

		ExtraHeaders:           []http.HTTPHeader{},
		Query:                  "",
		Method:                 "GET",
		AllowedHttpStatusCodes: []int{200, 201, 202, 203, 204, 205, 206, 207, 208, 209},

		AuthEnabled:                false,
		ContentCheckEnabled:        false,
		TlsCheckCertificates:       true,
		TlsCertExpirationThreshold: time.Hour * 24 * 30,

		DBClient: dbClient,
	}
	http2, err := http.NewHttpCheck(httpConfig2)
	if err != nil {
		fmt.Printf("DUMMYDB: failed to init httpcheck id:%d, error: %s  \n", httpConfig2.Id, err)
	}

	httpConfig3 := http.CheckConfig{
		Id:      3,
		Port:    443,
		Target:  "https://master.cz",
		Timeout: time.Second * 30,

		ExtraHeaders:           []http.HTTPHeader{},
		Query:                  "",
		Method:                 "HEAD",
		AllowedHttpStatusCodes: []int{200, 201, 202, 203, 204, 205, 206, 207, 208, 209},

		AuthEnabled:          false,
		ContentCheckEnabled:  false,
		TlsCheckCertificates: false,

		DBClient: dbClient,
	}
	http3, err := http.NewHttpCheck(httpConfig3)
	if err != nil {
		fmt.Printf("DUMMYDB: failed to init httpcheck id:%d, error: %s  \n", httpConfig3.Id, err)
	}

	return append(array, http1, http2, http3)
}

func getTestICMPChecks(dbClient check.DBInterface) (array []check.CheckInterface) {
	icmpConfig1 := icmp.CheckConfig{
		Id:      4,
		Target:  "8.8.8.8",
		Timeout: time.Second * 3,

		DBClient: dbClient,
	}
	icmp1, err := icmp.NewCheck(icmpConfig1)
	if err != nil {
		fmt.Printf("DUMMYDB: failed to init icmpCheck id:%d, error: %s  \n", icmpConfig1.Id, err)
	}

	icmpConfig2 := icmp.CheckConfig{
		Id:      5,
		Target:  "1.1.1.1",
		Timeout: time.Second * 2,

		DBClient: dbClient,
	}
	icmp2, err := icmp.NewCheck(icmpConfig2)
	if err != nil {
		fmt.Printf("DUMMYDB: failed to init icmpcheck id:%d, error: %s  \n", icmpConfig2.Id, err)
	}

	icmpConfig3 := icmp.CheckConfig{
		Id:      6,
		Target:  "8.8.8.8",
		Timeout: time.Second * 1,

		DBClient: dbClient,
	}
	icmp3, err := icmp.NewCheck(icmpConfig3)
	if err != nil {
		fmt.Printf("DUMMYDB: failed to init icmpheck id:%d, error: %s  \n", icmpConfig3.Id, err)
	}

	return append(array, icmp1, icmp2, icmp3)

}

func getTestTCPChecks(dbClient check.DBInterface) (array []check.CheckInterface) {
	tcpConfig1 := tcp.CheckConfig{
		Id:      7,
		Target:  "seznam.cz",
		Port:    80,
		Timeout: time.Second * 2,

		DBClient: dbClient,
	}
	tcp1, err := tcp.NewCheck(tcpConfig1)
	if err != nil {
		fmt.Printf("DUMMYDB: failed to init tcpCheck id:%d, error: %s  \n", tcpConfig1.Id, err)
	}

	tcpConfig2 := tcp.CheckConfig{
		Id:      8,
		Target:  "master.cz",
		Port:    80,
		Timeout: time.Second * 2,

		DBClient: dbClient,
	}
	tcp2, err := tcp.NewCheck(tcpConfig2)
	if err != nil {
		fmt.Printf("DUMMYDB: failed to init tcpCheck id:%d, error: %s  \n", tcpConfig2.Id, err)
	}

	tcpConfig3 := tcp.CheckConfig{
		Id:      9,
		Target:  "google.com",
		Port:    80,
		Timeout: time.Second * 2,

		DBClient: dbClient,
	}
	tcp3, err := tcp.NewCheck(tcpConfig3)
	if err != nil {
		fmt.Printf("DUMMYDB: failed to init tcpCheck id:%d, error: %s  \n", tcpConfig3.Id, err)
	}

	return append(array, tcp1, tcp2, tcp3)
}
