package service

import (
	chaos "github.com/arangodb/testAgent/service/chaos"
	cluster "github.com/arangodb/testAgent/service/cluster"
	logging "github.com/op/go-logging"
)

type ServiceConfig struct {
	AgencySize int
}

type ServiceDependencies struct {
	Logger         *logging.Logger
	ClusterBuilder cluster.ClusterBuilder
}

type Service struct {
	ServiceConfig
	ServiceDependencies

	cluster     cluster.Cluster
	chaosMonkey chaos.ChaosMonkey
}

// NewService instantiates a new Service from the given config
func NewService(config ServiceConfig, deps ServiceDependencies) (*Service, error) {
	return &Service{
		ServiceConfig:       config,
		ServiceDependencies: deps,
	}, nil
}

// Run performs the tests
func (s *Service) Run(stopChan chan struct{}) error {
	// Create the cluster
	s.Logger.Infof("Creating initial cluster (size %d)", s.AgencySize)
	c, err := s.ClusterBuilder.Create(s.AgencySize)
	if err != nil {
		return maskAny(err)
	}
	s.cluster = c

	// Create & start a chaos monkey
	s.Logger.Info("Creating chaos monkey")
	s.chaosMonkey = chaos.NewChaosMonkey(s.cluster)
	s.chaosMonkey.Start()

	// Run tests

	// Destroy cluster
	s.Logger.Info("Destroying chaos monkey")
	if err := s.cluster.Destroy(); err != nil {
		return maskAny(err)
	}

	return nil
}
