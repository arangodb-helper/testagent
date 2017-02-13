package arangodb

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/arangodb/testAgent/pkg/arangostarter"
	"github.com/arangodb/testAgent/pkg/docker"
	"github.com/arangodb/testAgent/service/cluster"
	logging "github.com/op/go-logging"
	"golang.org/x/sync/errgroup"
)

const (
	serverReadyTimeout = time.Minute * 5
)

type ArangodbConfig struct {
	MasterPort          int      // MasterPort for arangodb
	ArangodbImage       string   // Docker image containing arangodb
	ArangoImage         string   // Docker image containing arangod (can be empty)
	NetworkBlockerImage string   // Docker image container network-blocker
	DockerHostIP        string   // IP of docker host
	DockerEndpoints     []string // Endpoint used to reach the docker daemon(s)
	DockerNetHost       bool     // If set, run containers with `--net=host`
	Verbose             bool     // Turn on debug logging
	Privileged          bool     // Start containers with `--privileged`
}

// arangodbClusterBuilder implements a ClusterBuilder using arangodb.
type arangodbClusterBuilder struct {
	log *logging.Logger
	ArangodbConfig
}

type arangodbCluster struct {
	ArangodbConfig

	mutex            sync.Mutex
	log              *logging.Logger
	dockerHosts      []*docker.DockerHost
	id               string
	agencySize       int
	machines         []*arangodb
	lastMachineIndex int32
}

// NewArangodbClusterBuilder creates a new ClusterBuilder using arangodb.
func NewArangodbClusterBuilder(log *logging.Logger, config ArangodbConfig) (cluster.ClusterBuilder, error) {
	if config.MasterPort == 0 {
		return nil, maskAny(fmt.Errorf("MasterPort missing"))
	}
	if config.ArangodbImage == "" {
		return nil, maskAny(fmt.Errorf("ArangodbImage missing"))
	}
	if config.DockerHostIP == "" {
		return nil, maskAny(fmt.Errorf("DockerHostIP missing"))
	}
	if len(config.DockerEndpoints) == 0 {
		return nil, maskAny(fmt.Errorf("DockerEndpoints missing"))
	}
	return &arangodbClusterBuilder{
		log:            log,
		ArangodbConfig: config,
	}, nil
}

// Create creates and starts a new cluster.
// The number of "machines" created equals the given agency size.
// This function returns when the cluster is operational (or an error occurs)
func (cb *arangodbClusterBuilder) Create(agencySize int) (cluster.Cluster, error) {
	// Create docker hosts
	dockerHosts, err := docker.NewDockerHosts(cb.DockerEndpoints, cb.DockerHostIP)
	if err != nil {
		return nil, maskAny(err)
	}

	// Create random ID
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return nil, maskAny(err)
	}
	id := hex.EncodeToString(b)

	// Instantiate
	c := &arangodbCluster{
		log:              cb.log,
		ArangodbConfig:   cb.ArangodbConfig,
		dockerHosts:      dockerHosts,
		agencySize:       agencySize,
		id:               id,
		lastMachineIndex: 0,
	}

	// Start arangodb master
	if _, err := c.add(); err != nil {
		return nil, maskAny(err)
	}
	// Start arangodb slave several times
	if agencySize > 1 {
		g := errgroup.Group{}
		for i := 1; i < agencySize; i++ {
			g.Go(func() error {
				// Add machine
				if _, err := c.add(); err != nil {
					return maskAny(err)
				}
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			return nil, maskAny(err)
		}
	}

	return c, nil
}

// Block until all servers on all machines are ready
func (c *arangodbCluster) WaitUntilReady() error {
	machines, err := c.Machines()
	if err != nil {
		return maskAny(err)
	}
	g := errgroup.Group{}
	for _, m := range machines {
		m := m.(*arangodb)
		g.Go(func() error {
			// Wait until all servers are reachable
			if err := m.waitUntilServersReady(c.log, serverReadyTimeout); err != nil {
				return maskAny(err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return maskAny(err)
	}
	return nil
}

// ID returns a unique identifier for this cluster
func (c *arangodbCluster) ID() string {
	return c.id
}

// ArangoImage returns the arango (database) docker image used on this cluster
func (c *arangodbCluster) ArangoImage() string {
	return c.ArangodbConfig.ArangoImage
}

// Machines returns all machines in the cluster
func (c *arangodbCluster) Machines() ([]cluster.Machine, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	result := []cluster.Machine{}
	for _, m := range c.machines {
		result = append(result, m)
	}
	return result, nil
}

// Add adds a single machine to the cluster
func (c *arangodbCluster) Add() (cluster.Machine, error) {
	// Create & start machine
	m, err := c.add()
	if err != nil {
		return nil, maskAny(err)
	}

	// Wait until all servers are reachable
	ma := m.(*arangodb)
	if err := ma.waitUntilServersReady(c.log, serverReadyTimeout); err != nil {
		return nil, maskAny(err)
	}

	return m, nil
}

// Remove the entire cluster
func (c *arangodbCluster) Destroy() error {
	machines, err := c.Machines()
	if err != nil {
		return maskAny(err)
	}
	c.log.Infof("Destroying %d machines", len(machines))

	g := errgroup.Group{}
	for _, m := range machines {
		g.Go(m.Destroy)
	}
	if err := g.Wait(); err != nil {
		return maskAny(err)
	}
	return nil
}

func (c *arangodbCluster) add() (cluster.Machine, error) {
	// Create new index
	index := int(atomic.AddInt32(&c.lastMachineIndex, 1) - 1)

	// Create machine
	m, err := c.createMachine(index)
	if err != nil {
		return nil, maskAny(err)
	}

	// Register machine
	c.mutex.Lock()
	c.machines = append(c.machines, m)
	c.mutex.Unlock()

	// Pull arangodb image
	if err := m.pullImageIfNeeded(c.ArangodbConfig.ArangodbImage); err != nil {
		return nil, maskAny(err)
	}

	// Pull network-block image
	if err := m.pullImageIfNeeded(c.ArangodbConfig.NetworkBlockerImage); err != nil {
		return nil, maskAny(err)
	}

	// Start network blocker
	if err := m.startNetworkBlocker(c.ArangodbConfig.NetworkBlockerImage); err != nil {
		return nil, maskAny(err)
	}

	// Start machine
	if err := m.start(); err != nil {
		return nil, maskAny(err)
	}

	// Start watchdog
	m.watchdog()

	return m, nil
}

// masterArangodbClient creates a client for the master arangodb starter.
func (c *arangodbCluster) masterArangodbClient() (arangostarter.API, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(c.machines) == 0 {
		return nil, maskAny(fmt.Errorf("No master found"))
	}
	master := c.machines[0]
	masterIP := master.dockerHost.IP
	masterPort := master.arangodbPort

	client, err := arangostarter.NewArangoStarterClient(masterIP, masterPort)
	if err != nil {
		return nil, maskAny(err)
	}
	return client, nil
}
