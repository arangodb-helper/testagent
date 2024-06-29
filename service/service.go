package service

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"

	chaos "github.com/arangodb-helper/testagent/service/chaos"
	cluster "github.com/arangodb-helper/testagent/service/cluster"
	"github.com/arangodb-helper/testagent/service/reporter"
	"github.com/arangodb-helper/testagent/service/server"
	"github.com/arangodb-helper/testagent/service/test"
	logging "github.com/op/go-logging"
	"golang.org/x/sync/errgroup"
)

type ServiceConfig struct {
	ProjectVersion string
	ProjectBuild   string
	AgencySize     int
	ForceOneShard  bool
	ServerPort     int
	ReportDir      string
	MetricsDir     string
	CollectMetrics bool
	ChaosConfig    chaos.ChaosMonkeyConfig
	EnableTests    []string
}

type ServiceDependencies struct {
	Logger         *logging.Logger
	ClusterBuilder cluster.ClusterBuilder
	Tests          []test.TestScript
}

type Service struct {
	ServiceConfig
	ServiceDependencies

	cluster     cluster.Cluster
	chaosMonkey chaos.ChaosMonkey
	reporter    reporter.Reporter
	startedAt   time.Time
}

// NewService instantiates a new Service from the given config
func NewService(config ServiceConfig, deps ServiceDependencies) (*Service, error) {
	if config.ReportDir == "" {
		config.ReportDir = "."
	}
	s := &Service{
		ServiceConfig:       config,
		ServiceDependencies: deps,
	}
	s.reporter = reporter.NewReporter(config.ReportDir, deps.Logger, s)
	return s, nil
}

// Run performs the tests
func (s *Service) Run(stopChan chan struct{}, withChaos bool) error {
	s.startedAt = time.Now()
	// Start our HTTP server
	server.StartHTTPServer(s.Logger, s.ServerPort, s.ReportDir, s)

	// Create the cluster
	s.Logger.Infof("Creating initial cluster (size %d)", s.AgencySize)
	c, err := s.ClusterBuilder.Create(s.AgencySize, s.ForceOneShard)
	if err != nil {
		return maskAny(err)
	}
	s.cluster = c

	// Wait for cluster to become ready
	s.Logger.Info("Waiting for cluster ready")
	if err := c.WaitUntilReady(); err != nil {
		return maskAny(err)
	}

	// Start metrics collection
	if s.CollectMetrics {
		if err := c.StartMetricsCollection(); err != nil {
			return maskAny(err)
		}
	}

	s.Logger.Debug("Try to set enterprise license")
	enterpriseLicense := os.Getenv("ARANGO_ENTERPRISE_LICENSE")
	if enterpriseLicense != "" {
		m, _ := c.Machines()
		host := "http://" + m[0].DBServerURL().Host // Get address of DBServer

		// Perform request to get Arango version
		response, err := http.Get(host + "/_api/version")

		if err != nil {
			s.Logger.Error("ERROR making request for getting version")
			return maskAny(err)
		}

		// Struct for getting response
		type versionJSON struct {
			Server  string `json:"server"`
			License string `json:"license"`
			Version string `json:"version"`
		}

		// Decode response to JSON
		var versionObj versionJSON
		if err := json.NewDecoder(response.Body).Decode(&versionObj); err != nil {
			s.Logger.Error("ERROR parsing response body for version")
			return maskAny(err)
		}

		currVersion := semver.New(versionObj.Version) // Get current version of ArangoDB
		supportedVersion := semver.New("3.8.99")      // Version number to compare with

		// We support license feature since 3.9
		if supportedVersion.LessThan(*currVersion) {

			// Try to update license
			client := &http.Client{}
			req, _ := http.NewRequest("PUT", host+"/_admin/license", strings.NewReader(enterpriseLicense))
			req.SetBasicAuth("root", "")
			req.ContentLength = int64(len(enterpriseLicense))

			// Perform request to update license
			resp, err := client.Do(req)
			if err != nil {
				s.Logger.Error("ERROR making request for updating license")
				return maskAny(err)
			}

			// If not success
			if resp.StatusCode != 201 {
				// Struct for getting error response
				type errResponseJSON struct {
					Code         int    `json:"code"`
					Error        bool   `json:"error"`
					ErrorMessage string `json:"errorMessage"`
					ErrorNum     int    `json:"errorNum"`
				}

				// Decode response to JSON
				var response errResponseJSON
				if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
					s.Logger.Error("ERROR parsing response body for license")
					return maskAny(err)
				}

				s.Logger.Error("ERROR during license update: %s", response.ErrorMessage)
				return maskAny(err)

			} else {
				s.Logger.Debug("License successfully updated")
			}

		} else {
			s.Logger.Debug("License feature is supproted since 3.9.0")
		}

	} else {
		s.Logger.Debug("Enterprise license is not specified")
	}

	// Create & start a chaos monkey
	if withChaos {
		s.Logger.Info("Creating chaos monkey")
		s.chaosMonkey = chaos.NewChaosMonkey(s.Logger, s.cluster, s.ChaosConfig)
		s.chaosMonkey.Start()
	}

	// Run tests
	for _, t := range s.Tests() {
		s.Logger.Infof("Starting test %s", t.Name())
		if err := t.Start(s.cluster, s.reporter); err != nil {
			return maskAny(err)
		}
	}

	// Wait until stop
	s.Logger.Info("All tests started, waiting until termination")
	<-stopChan

	// Stop introducting chaos
	if withChaos {
		s.Logger.Info("Stopping chaos")
		s.chaosMonkey.Stop()
		s.chaosMonkey.WaitUntilInactive()
	}

	// Stop all tests
	s.Logger.Info("Stopping test scripts")
	g := errgroup.Group{}
	for _, t := range s.Tests() {
		t := t // t is used in nested func
		g.Go(func() error {
			if err := t.Stop(); err != nil {
				s.Logger.Errorf("Failed to stop test %s: %#v", t.Name(), err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		s.Logger.Errorf("Failed to stop tests: %#v", err)
	}

	// Destroy cluster
	s.Logger.Info("Destroying cluster")
	if err := s.cluster.Destroy(); err != nil {
		return maskAny(err)
	}

	return nil
}

func (s *Service) StartedAt() time.Time {
	return s.startedAt
}

func (s *Service) ProjectVersion() string {
	return s.ServiceConfig.ProjectVersion
}

func (s *Service) ProjectBuild() string {
	return s.ServiceConfig.ProjectBuild
}

func (s *Service) Cluster() cluster.Cluster {
	return s.cluster
}

func (s *Service) Tests() []test.TestScript {
	return s.ServiceDependencies.Tests
}

func (s *Service) ChaosMonkey() chaos.ChaosMonkey {
	return s.chaosMonkey
}

func (s *Service) Reports() []reporter.FailureReport {
	return s.reporter.Reports()
}
