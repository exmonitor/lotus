package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/exmonitor/exclient"
	"github.com/exmonitor/exlogger"
	"github.com/exmonitor/watcher/service"
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
	MariaUser         string
	MariaPassword     string

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
	rootCmd.PersistentFlags().StringVarP(&flags.ElasticConnection, "db-server", "", "", "Set elastic connection string.")
	rootCmd.PersistentFlags().StringVarP(&flags.MariaConnection, "maria-connection", "", "", "Set maria database connection string.")
	rootCmd.PersistentFlags().StringVarP(&flags.MariaUser, "maria-user", "", "", "Set Maria database user that wil be used for connection.")
	rootCmd.PersistentFlags().StringVarP(&flags.MariaPassword, "maria-password", "", "", "Set Maria database password that will be used for connection.")

	// other
	rootCmd.PersistentFlags().BoolVarP(&flags.Debug, "debug", "v", false, "Enable or disable more verbose log.")
	rootCmd.PersistentFlags().BoolVarP(&flags.TimeProfiling, "time-profiling", "", false, "Enable or disable time profiling.")

	rootCmd.Run = mainExecute

	err := rootCmd.Execute()

	if err != nil {
		panic(err)
	}
}

// main command execute function
func mainExecute(cmd *cobra.Command, args []string) {
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
	// catch Interrupt (Ctrl^C) or SIGTERM and exit
	catchOSSignals(logger)
	// setup db client config
	dbClientConfig := exclient.DBConfig{
		DBDriver:          flags.DBDriver,
		ElasticConnection: flags.ElasticConnection,
		MariaConnection:   flags.MariaConnection,
		MariaUser:         flags.MariaUser,
		MariaPassword:     flags.MariaPassword,
	}
	// init DB client
	dbClient, err := exclient.GetDBClient(dbClientConfig)
	if err != nil {
		logger.LogError(err, "failed to prepare DB Client")
		panic(err)
	}
	// fetch intervals for monitoring
	intervalGroups, err := dbClient.SQL_GetIntervals()
	if err != nil {
		logger.LogError(err, "failed to fetch monitoring intervals")
		panic(err)
	}
	// create thread for each intervalGroup
	service.InitIntervalGroups(intervalGroups, dbClient, logger)

	// sleep little friend
	fmt.Printf(">> Main thread sleeping forever ...\n")
	select {}
}

// catch Interrupt (Ctrl^C) or SIGTERM and exit
func catchOSSignals(l *exlogger.Logger) {
	// catch signals
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		// be sure to close log files
		s := <-c
		if flags.LogToFile {
			l.CloseLogs()
		}
		fmt.Printf("\n>> Caught signal %s, exiting ...\n\n", s.String())
		os.Exit(1)
	}()
}
