package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	service "github.com/arangodb-helper/testagent/service"
	arangodb "github.com/arangodb-helper/testagent/service/cluster/arangodb"
	"github.com/arangodb-helper/testagent/service/test"
	complex "github.com/arangodb-helper/testagent/tests/complex"
	"github.com/arangodb-helper/testagent/tests/simple"
	logging "github.com/op/go-logging"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Configuration data with defaults:

const (
	projectName             = "testAgent"
	defaultOperationTimeout = time.Minute * 6
	defaultRetryTimeout     = time.Minute * 8
	defaultStepTimeout      = time.Second * 15
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
		complex.ComplextTestConfig
		complex.DocColConfig
		complex.GraphTestConf
		logLevel string
	}
	maskAny = errors.WithStack
)

func init() {
	f := cmdMain.Flags()
	defaultDockerEndpoints := []string{"unix:///var/run/docker.sock"}
	// Full test list:
	// defaultTestList := []string{"simple", "DocColTest", "OneShardTest", "CommunityGraphTest", "SmartGraphTest", "EnterpriseGraphTest"}
	// Use only "simple" test by default for backwards compatibility
	defaultTestList := []string{"simple"}
	f.IntVar(&appFlags.AgencySize, "agency-size", 3, "Number of agents in the cluster")
	f.IntVar(&appFlags.port, "port", 4200, "First port of range of ports used by the testAgent")
	f.StringVar(&appFlags.logLevel, "log-level", "debug", "Minimum log level (debug|info|warning|error)")
	f.IntVar(&appFlags.ServiceConfig.ChaosConfig.ChaosLevel, "chaos-level", 4, "Chaos level. Default: 4.")
	f.StringVar(&appFlags.ArangodbImage, "arangodb-image", getEnvVar("ARANGODB_IMAGE", "arangodb/arangodb-starter"), "name of the Docker image containing arangodb (the cluster starter)")
	f.StringVar(&appFlags.ArangoImage, "arango-image", getEnvVar("ARANGO_IMAGE", ""), "name of the Docker image containing arangod (the database)")
	f.StringVar(&appFlags.NetworkBlockerImage, "network-blocker-image", getEnvVar("NETWORK_BLOCKER_IMAGE", ""), "name of the Docker image containing network-blocker")
	f.StringSliceVar(&appFlags.DockerEndpoints, "docker-endpoint", defaultDockerEndpoints, "Endpoints used to reach the docker daemons")
	f.StringVar(&appFlags.DockerHostIP, "docker-host-ip", "", "IP of the docker host")
	f.BoolVar(&appFlags.DockerNetHost, "docker-net-host", false, "If set, run all containers with `--net=host`")
	f.BoolVar(&appFlags.ForceOneShard, "force-one-shard", false, "If set, force one shard arangodb cluster")
	f.BoolVar(&appFlags.ReplicationVersion2, "replication-version-2", false, "If set, use replication version 2")
	f.BoolVar(&appFlags.FailedWriteConcern403, "return-403-on-failed-write-concern", false, "If set, option `--cluster.failed-write-concern-status-code` will not be set for DB servers, bringing it to the default value of 403. Otherwise this parameter will be set to 503.")
	f.StringVar(&appFlags.DockerInterface, "docker-interface", "docker0", "Network interface used to connect docker containers to")
	f.StringVar(&appFlags.ReportDir, "report-dir", getEnvVar("REPORT_DIR", "."), "Directory in which failure reports will be created")
	f.BoolVar(&appFlags.CollectMetrics, "collect-metrics", false, "If set, metrics will be collected and saved into files.")
	f.StringVar(&appFlags.MetricsDir, "metrics-dir", getEnvVar("METRICS_DIR", "."), "Directory in which metrics will be stored")
	f.BoolVar(&appFlags.Privileged, "privileged", false, "If set, run all containers with `--privileged`")
	f.IntVar(&appFlags.ChaosConfig.MaxMachines, "max-machines", 10, "Upper limit to the number of machines in a cluster")
	f.StringSliceVar(&appFlags.EnableTests, "enable-test", defaultTestList, "Enable particular test. Default: run all tests. Available tests: simple, DocColTest, OneShardTest, CommunityGraphTest, SmartGraphTest, EnterpriseGraphTest")
	f.IntVar(&appFlags.SimpleConfig.MaxDocuments, "simple-max-documents", 20000, "Upper limit to the number of documents created in simple test")
	f.IntVar(&appFlags.SimpleConfig.MaxCollections, "simple-max-collections", 10, "Upper limit to the number of collections created in simple test")
	f.DurationVar(&appFlags.SimpleConfig.OperationTimeout, "simple-operation-timeout", defaultOperationTimeout, "Timeout per database operation")
	f.DurationVar(&appFlags.SimpleConfig.RetryTimeout, "simple-retry-timeout", defaultRetryTimeout, "How long are tests retried before giving up")
	f.Int64Var(&appFlags.GraphTestConf.MaxVertices, "graph-max-vertices", 50000, "Upper limit to the number of vertices (graph tests)")
	f.IntVar(&appFlags.GraphTestConf.VertexSize, "graph-vertex-size", 512, "The size of the payload field in bytes in all vertices (graph tests)")
	f.IntVar(&appFlags.GraphTestConf.EdgeSize, "graph-edge-size", 256, "The size of the payload field in bytes in all vertices (graph tests)")
	f.IntVar(&appFlags.GraphTestConf.TraversalOperationsPerCycle, "graph-traversal-ops", 100, "How many traversal operations to perform in one test cycle (graph tests)")
	f.IntVar(&appFlags.GraphTestConf.BatchSize, "graph-batch-size", 500, "Batch size for creating documents (graph tests)")
	f.IntVar(&appFlags.DocColConfig.MaxDocuments, "doc-max-documents", 50000, "Upper limit to the number of documents created in document collection tests")
	f.IntVar(&appFlags.DocColConfig.BatchSize, "doc-batch-size", 500, "Batch size for creating documents in bulk mode in document collection tests")
	f.IntVar(&appFlags.DocColConfig.DocumentSize, "doc-document-size", 10240, "The size of the payload field in bytes in regular documents in document collection tests")
	f.IntVar(&appFlags.DocColConfig.MaxUpdates, "doc-max-updates", 3, "Number of update operations to be performed on each document, before dropping collection, in document collection tests.")
	f.IntVar(&appFlags.ComplextTestConfig.NumberOfShards, "complex-shards", 10, "Number of shards (\"complex\" test suite)")
	f.IntVar(&appFlags.ComplextTestConfig.ReplicationFactor, "complex-replicationFactor", 2, "Replication factor (\"complex\" test suite)")
	f.DurationVar(&appFlags.ComplextTestConfig.OperationTimeout, "complex-operation-timeout", defaultOperationTimeout, "Timeout per database operation (\"complex\" test suite)")
	f.DurationVar(&appFlags.ComplextTestConfig.RetryTimeout, "complex-retry-timeout", defaultRetryTimeout, "How long are tests retried before giving up (\"complex\" test suite)")
	f.DurationVar(&appFlags.ComplextTestConfig.StepTimeout, "complex-step-timeout", defaultStepTimeout, "Pause between test actions (\"complex\" test suite)")
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

	if appFlags.CollectMetrics {
		if _, err := os.Stat(appFlags.MetricsDir); os.IsNotExist(err) {
			log.Info("Metrics directory does not exist. Creating: %s", appFlags.MetricsDir)
			if err := os.Mkdir(appFlags.MetricsDir, 0755); err != nil {
				Exitf("Can't create metrics directory: %#v", err)
			}
		}
	}

	cb, err := arangodb.NewArangodbClusterBuilder(log, appFlags.MetricsDir, appFlags.CollectMetrics, appFlags.ArangodbConfig)
	if err != nil {
		log.Fatalf("Failed to create cluster builder: %#v", err)
	}

	// Create tests
	tests := []test.TestScript{}
	if slices.Contains(appFlags.EnableTests, "simple") {
		tests = append(tests, simple.NewSimpleTest(log, appFlags.ReportDir, appFlags.SimpleConfig))
	}
	if slices.Contains(appFlags.EnableTests, "DocColTest") {
		tests = append(tests, complex.NewRegularDocColTest(log, appFlags.ReportDir, appFlags.ComplextTestConfig, appFlags.DocColConfig))
	}
	if slices.Contains(appFlags.EnableTests, "OneShardTest") {
		tests = append(tests, complex.NewOneShardTest(log, appFlags.ReportDir, appFlags.ComplextTestConfig, appFlags.DocColConfig))
	}
	if slices.Contains(appFlags.EnableTests, "CommunityGraphTest") {
		tests = append(tests, complex.NewComGraphTest(log, appFlags.ReportDir, appFlags.ComplextTestConfig, appFlags.GraphTestConf))
	}
	if slices.Contains(appFlags.EnableTests, "SmartGraphTest") {
		tests = append(tests, complex.NewSmartGraphTest(log, appFlags.ReportDir, appFlags.ComplextTestConfig, appFlags.GraphTestConf))
	}
	if slices.Contains(appFlags.EnableTests, "EnterpriseGraphTest") {
		tests = append(tests, complex.NewEnterpriseGraphTest(log, appFlags.ReportDir, appFlags.ComplextTestConfig, appFlags.GraphTestConf))
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
