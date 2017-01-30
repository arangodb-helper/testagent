package service

import (
	chaos "github.com/arangodb/testAgent/service/chaos"
	cluster "github.com/arangodb/testAgent/service/cluster"
	"github.com/arangodb/testAgent/service/reporter"
	"github.com/arangodb/testAgent/service/server"
	"github.com/arangodb/testAgent/service/test"
	logging "github.com/op/go-logging"
	"golang.org/x/sync/errgroup"
)

type ServiceConfig struct {
	AgencySize int
	ServerPort int
	ReportDir  string
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
func (s *Service) Run(stopChan chan struct{}) error {
	// Start our HTTP server
	server.StartHTTPServer(s.Logger, s.ServerPort, s)

	// Create the cluster
	s.Logger.Infof("Creating initial cluster (size %d)", s.AgencySize)
	c, err := s.ClusterBuilder.Create(s.AgencySize)
	if err != nil {
		return maskAny(err)
	}
	s.cluster = c

	// Create & start a chaos monkey
	s.Logger.Info("Creating chaos monkey")
	s.chaosMonkey = chaos.NewChaosMonkey(s.Logger, s.cluster)
	s.chaosMonkey.Start()

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
	s.Logger.Info("Stopping chaos")
	s.chaosMonkey.Stop()
	s.chaosMonkey.WaitUntilInactive()

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

func (s *Service) Cluster() cluster.Cluster {
	return s.cluster
}

func (s *Service) Tests() []test.TestScript {
	return s.ServiceDependencies.Tests
}

func (s *Service) ChaosMonkey() chaos.ChaosMonkey {
	return s.chaosMonkey
}
