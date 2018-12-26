package multi

import (
	"database/sql"
	"fmt"

	"github.com/exmonitor/exlogger"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"

	"github.com/exmonitor/exclient/database"
	"github.com/exmonitor/chronos"
)

const (
	sqlDriver = "mysql"
)

func DBDriverName() string {
	return "multi"
}

// config for multi db client with mariaDB and elastic search
type Config struct {
	// elastic search
	ElasticConnection string
	// maria db
	MariaConnection   string
	MariaUser         string
	MariaPassword     string
	MariaDatabaseName string

	Logger        *exlogger.Logger
	TimeProfiling bool
}

type Client struct {
	sqlClient *sql.DB

	logger        *exlogger.Logger
	timeProfiling bool
	// implement client db interface
	database.ClientInterface
}

func New(conf Config) (*Client, error) {
	if conf.Logger == nil {
		return nil, errors.Wrapf(invalidConfigError, "conf.Logger must not be nil")
	}

	t1 := chronos.New()
	// create sql connection string
	sqlConnectionString := mysqlConnectionString(conf.MariaConnection, conf.MariaUser, conf.MariaPassword, conf.MariaDatabaseName)
	// init sql connection
	db, err := sql.Open(sqlDriver, sqlConnectionString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create sql connection")
	}
	err = db.Ping()
	if err != nil {
		return nil, errors.Wrap(err, "failed to ping sql connection")
	}
	t1.Finish()
	conf.Logger.Log("successfully connected to sql db %s", conf.MariaConnection)
	if conf.TimeProfiling {
		conf.Logger.LogDebug("TIME_PROFILING: created sql connection in %sms",t1.StringMilisec())
	}
	// elastic search connection
	// TODO
	newClient := &Client{
		sqlClient: db,

		logger:        conf.Logger,
		timeProfiling: conf.TimeProfiling,
	}
	return newClient, nil
}

func mysqlConnectionString(mariaConnection string, mariaUser string, mariaPassword string, mariaDatabaseName string) string {
	return fmt.Sprintf("%s:%s@%s/%s", mariaUser, mariaPassword, mariaConnection, mariaDatabaseName)
}

// close db connections
func (c *Client) Close() {
	c.sqlClient.Close()
	c.logger.Log("successfully closed sql connection")
}