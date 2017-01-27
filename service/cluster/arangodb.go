package cluster

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	logging "github.com/op/go-logging"
	"golang.org/x/sync/errgroup"
)

const (
	serverReadyTimeout = time.Minute * 5
)

type ArangodbConfig struct {
	MasterPort     int
	ArangodbImage  string
	DockerHostIP   string
	DockerEndpoint string
}

// arangodbClusterBuilder implements a ClusterBuilder using arangodb.
type arangodbClusterBuilder struct {
	log *logging.Logger
	ArangodbConfig
}

type arangodbCluster struct {
	ArangodbConfig

	mutex      sync.Mutex
	log        *logging.Logger
	id         string
	agencySize int
	client     *docker.Client
	machines   []arangodb
}

// NewArangodbClusterBuilder creates a new ClusterBuilder using arangodb.
func NewArangodbClusterBuilder(log *logging.Logger, config ArangodbConfig) (ClusterBuilder, error) {
	if config.MasterPort == 0 {
		return nil, maskAny(fmt.Errorf("MasterPort missing"))
	}
	if config.ArangodbImage == "" {
		return nil, maskAny(fmt.Errorf("ArangodbImage missing"))
	}
	if config.DockerHostIP == "" {
		return nil, maskAny(fmt.Errorf("DockerHostIP missing"))
	}
	if config.DockerEndpoint == "" {
		return nil, maskAny(fmt.Errorf("DockerEndpoint missing"))
	}
	return &arangodbClusterBuilder{
		log:            log,
		ArangodbConfig: config,
	}, nil
}

// Create creates and starts a new cluster.
// The number of "machines" created equals the given agency size.
// This function returns when the cluster is operational (or an error occurs)
func (cb *arangodbClusterBuilder) Create(agencySize int) (Cluster, error) {
	// Create docker client
	client, err := docker.NewClient(cb.ArangodbConfig.DockerEndpoint)
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
		log:            cb.log,
		ArangodbConfig: cb.ArangodbConfig,
		client:         client,
		agencySize:     agencySize,
		id:             id,
	}

	// Start arangodb several times
	g := errgroup.Group{}
	for i := 0; i < agencySize; i++ {
		index := i // using index in goroutine
		g.Go(func() error {
			// Launch machine
			m, err := c.launchArangodb(index)
			if err != nil {
				return maskAny(err)
			}
			// Wait until all servers are reachable
			if err := m.waitUntilServersReady(c.log, serverReadyTimeout); err != nil {
				return maskAny(err)
			}

			// Register machine
			c.mutex.Lock()
			defer c.mutex.Unlock()
			c.machines = append(c.machines, *m)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, maskAny(err)
	}

	return c, nil
}

// RestartMachine restarts a single machine in the cluster
func (c *arangodbCluster) RestartMachine() error {
	//TODO
	return nil
}

// Remove the entire cluster
func (c *arangodbCluster) Destroy() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	g := errgroup.Group{}
	for _, m := range c.machines {
		g.Go(m.destroy)
	}
	if err := g.Wait(); err != nil {
		return maskAny(err)
	}
	return nil
}
