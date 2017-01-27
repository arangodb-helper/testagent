package chaos

import (
	cluster "github.com/arangodb/testAgent/service/cluster"
)

type ChaosMonkey interface {
	// Start introducing chaos
	Start()
}

// NewChaosMonkey creates a new chaos monkey for the given cluster
func NewChaosMonkey(cluster cluster.Cluster) ChaosMonkey {
	return &chaosMonkey{
		cluster: cluster,
	}
}

type chaosMonkey struct {
	cluster cluster.Cluster
}

func (c *chaosMonkey) Start() {
	// TODO
}
