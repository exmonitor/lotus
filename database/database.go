package database

import (
	"fmt"
	"github.com/giantswarm/project-lotus/lotus/check"
	"github.com/giantswarm/project-lotus/lotus/database/dummydb"
	"github.com/pkg/errors"
)

func GetDBClient(dbDriver, dbServer, dbUsername, dbPassword string) (check.DBInterface, error) {
	switch dbDriver {
	case dummydb.DBDriverName:
		{
			fmt.Printf(">> Sucesfully connected to %s.\n", dummydb.DBDriverName)
			return dummydb.DummyDB{}, nil
		}
	// TODO add real DB drivers
	default:
		return nil, errors.Wrap(unknownDBDriverError, dbDriver)
	}
}
