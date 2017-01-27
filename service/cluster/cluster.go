package cluster

import "net/url"

type ClusterBuilder interface {
	// Create creates and starts a new cluster.
	// The number of "machines" created equals the given agency size.
	// This function returns when the cluster is operational (or an error occurs)
	Create(agencySize int) (Cluster, error)
}

type Cluster interface {
	// Machines returns all current machines in the cluster.
	Machines() ([]Machine, error)

	// Remove the entire cluster
	Destroy() error
}

// Machine represents a single "computer" on which an optional agent, a coordinator and a dbserver runs.
type Machine interface {
	// HasAgent returns true if there is an agent on this machine
	HasAgent() bool
	// AgentURL returns the URL of the agent on this machine.
	AgentURL() url.URL
	// DBServerURL returns the URL of the DBServer on this machine.
	DBServerURL() url.URL
	// CoordinatorURL returns the URL of the Coordinator on this machine.
	CoordinatorURL() url.URL

	// Reboot performs a graceful reboot of the machine
	Reboot() error

	// Remove the machine without the ability to recover it
	Destroy() error
}
