package dummydb

import (
	"github.com/giantswarm/project-lotus/lotus/check"
)

const DBDriverName  = "dummydb"

type DummyDB struct {
	// implement db interface
	check.DBInterface
}

func (d DummyDB) GetAllChecks(filter *check.CheckFilter) []check.CheckInterface {
	var checks []check.CheckInterface

	if filter.IntervalSec == 10 {
		checks = append(checks, getTestICMPChecks(&d)...)
	}

	if filter.IntervalSec == 30 {
		checks = append(checks, getTestTCPChecks(&d)...)
	}

	if filter.IntervalSec == 60 {
		checks = append(checks, getTestHttpChecks(&d)...)
	}

	return checks
}

func (d DummyDB) SaveCheckStatus(status *check.Status) error {

	return nil
}
