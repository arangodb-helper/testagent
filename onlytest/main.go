package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	service "github.com/arangodb-helper/testagent/service"
	arangodb "github.com/arangodb-helper/testagent/service/cluster/arangodb"
	"github.com/arangodb-helper/testagent/service/cluster"
	"github.com/arangodb-helper/testagent/service/test"
	"github.com/arangodb-helper/testagent/tests/simple"
	logging "github.com/op/go-logging"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Configuration data with defaults:

const (
	projectName             = "testAgent"
	defaultOperationTimeout = time.Second * 60 // Should be 15s
	defaultRetryTimeout     = time.Minute * 4  // Should be 1m
)

var (
	projectVersion = "dev"
	projectBuild   = "dev"
	cmdMain        = cobra.Command{
		Use:   projectName,
		Short: "Test long running operations on existing ArangoDB cluster",
		Run:   cmdMainRun,
	}
	log      = logging.MustGetLogger(projectName)
	appFlags struct {
		port int
		service.ServiceConfig
		arangodb.ArangodbConfig
		simple.SimpleConfig
		logLevel string
	}
	maskAny = errors.WithStack
)

func init() {
	f := cmdMain.Flags()
	f.IntVar(&appFlags.AgencySize, "agency-size", 3, "Number of agents in the cluster")
	f.IntVar(&appFlags.port, "port", 4200, "First port of range of ports used by the testAgent")
	f.StringVar(&appFlags.logLevel, "log-level", "debug", "Minimum log level (debug|info|warning|error)")
	f.StringVar(&appFlags.ReportDir, "report-dir", getEnvVar("REPORT_DIR", "."), "Directory in which failure reports will be created")
	f.IntVar(&appFlags.SimpleConfig.MaxDocuments, "simple-max-documents", 20000, "Upper limit to the number of documents created in simple test")
	f.IntVar(&appFlags.SimpleConfig.MaxCollections, "simple-max-collections", 10, "Upper limit to the number of collections created in simple test")
	f.DurationVar(&appFlags.SimpleConfig.OperationTimeout, "simple-operation-timeout", defaultOperationTimeout, "Timeout per database operation")
	f.DurationVar(&appFlags.SimpleConfig.RetryTimeout, "simple-retry-timeout", defaultRetryTimeout, "How long are tests retried before giving up")
}

// handleSignal listens for termination signals and stops this process onup termination.
func handleSignal(sigChannel chan os.Signal, stopChan chan struct{}) {
	signalCount := 0
	for s := range sigChannel {
		signalCount++
		fmt.Println("Received signal:", s)
		if signalCount > 1 {
			os.Exit(1)
		}
		stopChan <- struct{}{}
	}
}

func main() {
	cmdMain.Execute()
}

func cmdMainRun(cmd *cobra.Command, args []string) {
	log.Infof("Starting %s version %s, build %s", projectName, projectVersion, projectBuild)

	level, err := logging.LogLevel(appFlags.logLevel)
	if err != nil {
		Exitf("Invalid log-level '%s': %#v", appFlags.logLevel, err)
	}
	logging.SetLevel(level, projectName)
	appFlags.ArangodbConfig.Verbose = appFlags.logLevel == "debug"

	// Interrupt signal:
	sigChannel := make(chan os.Signal)
	stopChan := make(chan struct{}, 10)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
	go handleSignal(sigChannel, stopChan)

	// Create tests
	tests := []test.TestScript{
		simple.NewSimpleTest(log, appFlags.ReportDir, appFlags.SimpleConfig),
	}

	// Create service
	log.Debug("creating service")
	appFlags.ServiceConfig.ProjectVersion = projectVersion
	appFlags.ServiceConfig.ProjectBuild = projectBuild
	service, err := service.NewService(appFlags.ServiceConfig, service.ServiceDependencies{
		Logger:         log,
		ClusterBuilder: cluster.NewFakeCluster(3, 3, 3),
		Tests:          tests,
	})
	if err != nil {
		log.Fatalf("Failed to create service: %#v", err)
	}

	// Run the service
	if err := service.Run(stopChan, false); err != nil {
		log.Fatalf("Run failed: %#v", err)
	}
	log.Info("Test completed")
}

// getEnvVar returns the value of the environment variable with given key of the given default
// value of no such variable exist or is empty.
func getEnvVar(key, defaultValue string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return defaultValue
}

func Exitf(format string, args ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format = format + "\n"
	}
	fmt.Printf(format, args...)
	os.Exit(1)
}

