package database

import (
	"fmt"
	"github.com/exmonitor/watcher/check"
	"github.com/exmonitor/watcher/database/dummydb"
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
