package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/exmonitor/exclient"
	"github.com/exmonitor/exclient/database"
	"github.com/exmonitor/exlogger"
	"github.com/exmonitor/watcher/interval"
	"time"
)

var Flags struct {
	// conf file
	ConfigFile string

	// logs
	LogToFile    bool
	LogFile      string
	LogErrorFile string

	// db
	DBDriver          string
	ElasticConnection string
	MariaConnection   string
	MariaDatabaseName string
	MariaUser         string
	MariaPassword     string
	CacheEnabled      bool
	CacheTTl          string

	// other
	TimeProfiling bool
	Debug         bool
}

var flags = Flags
var rootCmd = &cobra.Command{
	Use:   "watcher",
	Short: "watcher is a backend monitoring service for exmonitor system",
	Long: `Watcher is a backend monitoring service for exmonitor system.
Lotus fetches data from database and then run periodically (depending on IntervalGroups) monitoring checks. 
Result of checks is stored back into database.
Every monitoring check run in separate thread to avoid delays because of IO operations.`,
}

func main() {

	// config
	rootCmd.PersistentFlags().StringVarP(&flags.ConfigFile, "config", "c", "", "Set config file which will be used for fetching configuration.")

	// logs
	rootCmd.PersistentFlags().BoolVarP(&flags.LogToFile, "log-to-file", "", false, "Enable or disable logging to file.")
	rootCmd.PersistentFlags().StringVarP(&flags.LogFile, "log-file", "", "./notification.log", "Set filepath of log output. Used only when log-to-file is set to true.")
	rootCmd.PersistentFlags().StringVarP(&flags.LogErrorFile, "log-error-file", "", "./notification.error.log", "Set filepath of error log output. Used only when log-to-file is set to true.")

	// database
	rootCmd.PersistentFlags().StringVarP(&flags.DBDriver, "db-driver", "", "dummydb", "Set database driver that wil be used for connection")
	rootCmd.PersistentFlags().StringVarP(&flags.ElasticConnection, "elastic-connection", "", "http://127.0.0.1:9200", "Set elastic connection string.")
	rootCmd.PersistentFlags().StringVarP(&flags.MariaConnection, "maria-connection", "", "", "Set maria database connection string.")
	rootCmd.PersistentFlags().StringVarP(&flags.MariaDatabaseName, "maria-database-name", "", "monitoring_prod", "Set maria database name.")
	rootCmd.PersistentFlags().StringVarP(&flags.MariaUser, "maria-user", "", "", "Set Maria database user that wil be used for connection.")
	rootCmd.PersistentFlags().StringVarP(&flags.MariaPassword, "maria-password", "", "", "Set Maria database password that will be used for connection.")
	// cache
	rootCmd.PersistentFlags().BoolVarP(&flags.CacheEnabled, "cache", "", false, "Enable or disable caching of db records")
	rootCmd.PersistentFlags().StringVarP(&flags.CacheTTl, "cache-ttl", "", "5m", "Set cache ttl. Must be in time.Duration format. Value lower than 1m doesnt make sense.")

	// other
	rootCmd.PersistentFlags().BoolVarP(&flags.Debug, "debug", "v", false, "Enable or disable more verbose log.")
	rootCmd.PersistentFlags().BoolVarP(&flags.TimeProfiling, "time-profiling", "", false, "Enable or disable time profiling. Logs are printed via debug log.")

	rootCmd.Run = mainExecute

	err := rootCmd.Execute()

	if err != nil {
		panic(err)
	}
}

func validateFlags() {
	if flags.TimeProfiling && !flags.Debug {
		fmt.Printf("WARNING: time profiling is shown via debug log,  if you dont enabled debug log you wont see time profiling output.\n")
	}
}

// main command execute function
func mainExecute(cmd *cobra.Command, args []string) {
	validateFlags()

	// setup logger
	logConfig := exlogger.Config{
		Debug:        flags.Debug,
		LogToFile:    flags.LogToFile,
		LogFile:      flags.LogFile,
		LogErrorFile: flags.LogErrorFile,
	}

	logger, err := exlogger.New(logConfig)
	if err != nil {
		panic(err)
	}
	defer logger.CloseLogs()
	logger.Log("started logger")

	// parse cache ttl
	cacheTTL, err := time.ParseDuration(flags.CacheTTl)
	if err != nil {
		fmt.Printf("Failed to parse cache TTL. %s is not valid format for time.Duration\n", flags.CacheTTl)
		panic(err)
	}

	// setup db client config
	dbClientConfig := exclient.DBConfig{
		DBDriver:          flags.DBDriver,
		ElasticConnection: flags.ElasticConnection,
		MariaConnection:   flags.MariaConnection,
		MariaDatabaseName: flags.MariaDatabaseName,
		MariaUser:         flags.MariaUser,
		MariaPassword:     flags.MariaPassword,
		CacheEnabled:      flags.CacheEnabled,
		CacheTTL:          cacheTTL,

		Logger:        logger,
		TimeProfiling: flags.TimeProfiling,
	}
	// init DB client
	dbClient, err := exclient.GetDBClient(dbClientConfig)
	if err != nil {
		logger.LogError(err, "failed to prepare DB Client")
		panic(err)
	}
	defer dbClient.Close()
	// catch Interrupt (Ctrl^C) or SIGTERM and exit
	catchOSSignals(logger, dbClient)

	// fetch intervals for monitoring
	intervalGroups, err := dbClient.SQL_GetIntervals()
	if err != nil {
		logger.LogError(err, "failed to fetch monitoring intervals")
		panic(err)
	}
	// create thread for each intervalGroup
	interval.InitIntervalGroups(intervalGroups, dbClient, logger)

	// sleep little friend
	fmt.Printf(">> Main thread sleeping forever ...\n")
	select {}
}

// catch Interrupt (Ctrl^C) or SIGTERM and exit
func catchOSSignals(l *exlogger.Logger, dbClient database.ClientInterface) {
	// catch signals
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-c
		// be sure to close log files
		if flags.LogToFile {
			l.Log(">> Caught signal %s, exiting ...",s.String())
			l.LogError(nil,">> Caught signal %s, exiting ...",s.String())
			l.CloseLogs()
		}
		// close DB Connection
		dbClient.Close()

		fmt.Printf("\n>> Caught signal %s, exiting ...\n\n", s.String())
		os.Exit(0)
	}()
}
