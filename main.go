package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	service "github.com/arangodb-helper/testagent/service"
	arangodb "github.com/arangodb-helper/testagent/service/cluster/arangodb"
	"github.com/arangodb-helper/testagent/service/test"
	"github.com/arangodb-helper/testagent/tests/simple"
	logging "github.com/op/go-logging"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Configuration data with defaults:

const (
	projectName             = "testAgent"
	defaultOperationTimeout = time.Second * 90 // Should be 15s
	defaultRetryTimeout     = time.Minute * 4  // Should be 1m
)

var (
	projectVersion = "dev"
	projectBuild   = "dev"
	cmdMain        = cobra.Command{
		Use:   projectName,
		Short: "Test long running operations on ArangoDB clusters while introducing chaos",
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
	defaultDockerEndpoints := []string{"unix:///var/run/docker.sock"}
	f.IntVar(&appFlags.AgencySize, "agency-size", 3, "Number of agents in the cluster")
	f.IntVar(&appFlags.port, "port", 4200, "First port of range of ports used by the testAgent")
	f.StringVar(&appFlags.logLevel, "log-level", "debug", "Minimum log level (debug|info|warning|error)")
	f.StringVar(&appFlags.ArangodbImage, "arangodb-image", getEnvVar("ARANGODB_IMAGE", "arangodb/arangodb-starter"), "name of the Docker image containing arangodb (the cluster starter)")
	f.StringVar(&appFlags.ArangoImage, "arango-image", getEnvVar("ARANGO_IMAGE", ""), "name of the Docker image containing arangod (the database)")
	f.StringVar(&appFlags.NetworkBlockerImage, "network-blocker-image", getEnvVar("NETWORK_BLOCKER_IMAGE", ""), "name of the Docker image containing network-blocker")
	f.StringSliceVar(&appFlags.DockerEndpoints, "docker-endpoint", defaultDockerEndpoints, "Endpoints used to reach the docker daemons")
	f.StringVar(&appFlags.DockerHostIP, "docker-host-ip", "", "IP of the docker host")
	f.BoolVar(&appFlags.DockerNetHost, "docker-net-host", false, "If set, run all containers with `--net=host`")
	f.BoolVar(&appFlags.ForceOneShard, "force-one-shard", false, "If set, force one shard arangodb cluster")
	f.BoolVar(&appFlags.ReplicationVersion2, "replication-version-2", false, "If set, use replication version 2")
	f.StringVar(&appFlags.DockerInterface, "docker-interface", "docker0", "Network interface used to connect docker containers to")
	f.StringVar(&appFlags.ReportDir, "report-dir", getEnvVar("REPORT_DIR", "."), "Directory in which failure reports will be created")
	f.BoolVar(&appFlags.Privileged, "privileged", false, "If set, run all containers with `--privileged`")
	f.IntVar(&appFlags.ChaosConfig.MaxMachines, "max-machines", 10, "Upper limit to the number of machines in a cluster")
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

	logging.SetFormatter(logging.MustStringFormatter(`%{time:15:04:05.000} %{shortfunc} %{message}`))
	log.Infof("Starting %s version %s, build %s", projectName, projectVersion, projectBuild)

	level, err := logging.LogLevel(appFlags.logLevel)
	if err != nil {
		Exitf("Invalid log-level '%s': %#v", appFlags.logLevel, err)
	}
	logging.SetLevel(level, projectName)
	appFlags.ArangodbConfig.Verbose = appFlags.logLevel == "debug"

	// Get host IP
	if appFlags.ArangodbConfig.DockerHostIP == "" {
		if !appFlags.DockerNetHost && os.Getenv("RUNNING_IN_DOCKER") != "" {
			log.Fatal("When running in docker you must specify a --docker-host-ip")
		}
		ip, err := findLocalIP()
		if err != nil {
			log.Fatalf("Cannot detect local IP: %v", err)
		}
		log.Infof("Detected local IP %s", ip)
		appFlags.ArangodbConfig.DockerHostIP = ip
	}

	if appFlags.DockerNetHost {
		// Network chaos is not supported with host networking
		appFlags.ChaosConfig.DisableNetworkChaos = true
	}

	// Setup ports
	appFlags.ServerPort = appFlags.port
	appFlags.ArangodbConfig.MasterPort = appFlags.port + 1

	// Interrupt signal:
	sigChannel := make(chan os.Signal)
	stopChan := make(chan struct{}, 10)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
	go handleSignal(sigChannel, stopChan)

	// Create cluster builder
	log.Debug("creating arangodb cluster builder")
	cb, err := arangodb.NewArangodbClusterBuilder(log, appFlags.ArangodbConfig)
	if err != nil {
		log.Fatalf("Failed to create cluster builder: %#v", err)
	}

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
		ClusterBuilder: cb,
		Tests:          tests,
	})
	if err != nil {
		log.Fatalf("Failed to create service: %#v", err)
	}

	// Run the service
	if err := service.Run(stopChan, true); err != nil {
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

func findLocalIP() (string, error) {
	ifas, err := net.InterfaceAddrs()
	if err != nil {
		return "", maskAny(err)
	}
	for _, ia := range ifas {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := ia.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", maskAny(fmt.Errorf("No suitable address found"))
}
