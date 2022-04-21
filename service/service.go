package service

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	chaos "github.com/arangodb-helper/testagent/service/chaos"
	cluster "github.com/arangodb-helper/testagent/service/cluster"
	"github.com/arangodb-helper/testagent/service/reporter"
	"github.com/arangodb-helper/testagent/service/server"
	"github.com/arangodb-helper/testagent/service/test"
	"github.com/arangodb-helper/testagent/tests/util"
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
	ChaosConfig    chaos.ChaosMonkeyConfig
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

	type versionJSON struct {
		Server  string `json:"server"`
		License string `json:"license"`
		Version string `json:"version"`
	}

	enterpriseLicense := os.Getenv("ARANGO_ENTERPRISE_LICENSE")
	if enterpriseLicense != "" {
		client := util.NewArangoClient(s.Logger, c)
		hdr := make(map[string]string)
		hdr["accept"] = "application/json"
		var versionObj versionJSON
		res, e := client.Get("/_api/version", nil, hdr, &versionObj, []int{200}, []int{}, 60, 1)

		if e != nil || res == nil {
			s.Logger.Errorf("Error getting version")
			return maskAny(e[0])
		}

		// Get normalized version string with format "major.minor"
		var normalizedVersion string
		c := strings.Count(versionObj.Version, ".")
		if c != 2 {
			normalizedVersion = versionObj.Version
		} else {
			i := strings.LastIndex(versionObj.Version, ".")
			normalizedVersion = versionObj.Version[:i]
		}

		// Transform version to float for more convenient comparison
		versionFloat, err := strconv.ParseFloat(normalizedVersion, 32)
		if err != nil {
			s.Logger.Errorf("Version number is incorrect: %s", normalizedVersion)
			return maskAny(e[0])
		}

		// We support license feature since 3.9
		if versionFloat >= 3.9 {

			type CursorResponse struct {
				HasMore bool          `json:"hasMore,omitempty"`
				ID      string        `json:"id,omitempty"`
				Result  []interface{} `json:"result,omitempty"`
			}

			// Try to set license
			var cursorResp CursorResponse
			resp, err := client.Post("/_admin/license", nil, nil, enterpriseLicense, "", &cursorResp, []int{201},
				[]int{400, 404}, 60, 1)

			if err != nil || resp == nil {
				s.Logger.Errorf("Error during license update")
				return maskAny(e[0])
			}

			if resp[0].StatusCode != 201 {
				return maskAny(fmt.Errorf(
					"Recieved %d code during license update", resp[0].StatusCode))

			}

		}

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
