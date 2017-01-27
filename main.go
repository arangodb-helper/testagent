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
		service.ServiceConfig
		arangodb.ArangodbConfig
		logLevel string
	}
	maskAny = errgo.MaskFunc(errgo.Any)
)

func init() {
	f := cmdMain.Flags()
	f.IntVar(&appFlags.AgencySize, "agencySize", 3, "Number of agents in the cluster")
	f.IntVar(&appFlags.MasterPort, "masterPort", 4000, "Port to listen on for other arangodb's to join")
	f.StringVar(&appFlags.logLevel, "log-level", "debug", "Minimum log level (debug|info|warning|error)")
	f.StringVar(&appFlags.ArangodbImage, "arangodb-image", getEnvVar("ARANGODB_IMAGE", "arangodb/arangodb-starter"), "name of the Docker image containing arangodb (the cluster starter)")
	f.StringVar(&appFlags.ArangoImage, "arango-image", "", "name of the Docker image containing arangod (the database)")
	f.StringVar(&appFlags.DockerEndpoint, "dockerEndpoint", "unix:///var/run/docker.sock", "Endpoint used to reach the docker daemon")
	f.StringVar(&appFlags.DockerHostIP, "docker-host-ip", "", "IP of the docker host")
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

	// Create service
	log.Debug("creating service")
	service, err := service.NewService(appFlags.ServiceConfig, service.ServiceDependencies{
		Logger:         log,
		ClusterBuilder: cb,
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
