package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	service "github.com/arangodb/testAgent/service"
	arangodb "github.com/arangodb/testAgent/service/cluster/arangodb"
	"github.com/arangodb/testAgent/service/test"
	"github.com/arangodb/testAgent/tests/simple"
	"github.com/juju/errgo"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
)

// Configuration data with defaults:

const (
	projectName = "testAgent"
)

var (
	cmdMain = cobra.Command{
		Use:   projectName,
		Short: "Test long running operations on ArangoDB clusters while introducing chaos",
		Run:   cmdMainRun,
	}
	log      = logging.MustGetLogger(projectName)
	appFlags struct {
		port int
		service.ServiceConfig
		arangodb.ArangodbConfig
		logLevel string
	}
	maskAny = errgo.MaskFunc(errgo.Any)
)

func init() {
	f := cmdMain.Flags()
	f.IntVar(&appFlags.AgencySize, "agency-size", 3, "Number of agents in the cluster")
	f.IntVar(&appFlags.port, "port", 4200, "First port of range of ports used by the testAgent")
	f.StringVar(&appFlags.logLevel, "log-level", "debug", "Minimum log level (debug|info|warning|error)")
	f.StringVar(&appFlags.ArangodbImage, "arangodb-image", getEnvVar("ARANGODB_IMAGE", "arangodb/arangodb-starter"), "name of the Docker image containing arangodb (the cluster starter)")
	f.StringVar(&appFlags.ArangoImage, "arango-image", getEnvVar("ARANGO_IMAGE", ""), "name of the Docker image containing arangod (the database)")
	f.StringVar(&appFlags.DockerEndpoint, "docker-endpoint", "unix:///var/run/docker.sock", "Endpoint used to reach the docker daemon")
	f.StringVar(&appFlags.DockerHostIP, "docker-host-ip", "", "IP of the docker host")
	f.StringVar(&appFlags.ReportDir, "report-dir", getEnvVar("REPORT_DIR", "."), "Directory in which failure reports will be created")
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
	level, err := logging.LogLevel(appFlags.logLevel)
	if err != nil {
		Exitf("Invalid log-level '%s': %#v", appFlags.logLevel, err)
	}
	logging.SetLevel(level, projectName)

	// Get host IP
	if appFlags.ArangodbConfig.DockerHostIP == "" {
		ip, err := findLocalIP()
		if err != nil {
			log.Fatalf("Cannot detect local IP: %v", err)
		}
		log.Infof("Detected local IP %s", ip)
		appFlags.ArangodbConfig.DockerHostIP = ip
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
		simple.NewSimpleTest(log),
	}

	// Create service
	log.Debug("creating service")
	service, err := service.NewService(appFlags.ServiceConfig, service.ServiceDependencies{
		Logger:         log,
		ClusterBuilder: cb,
		Tests:          tests,
	})
	if err != nil {
		log.Fatalf("Failed to create service: %#v", err)
	}

	// Run the service
	if err := service.Run(stopChan); err != nil {
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
