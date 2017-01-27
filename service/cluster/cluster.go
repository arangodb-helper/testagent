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

type MachineState int

const (
	MachineStateNew      = MachineState(0) // Initial state
	MachineStateStarted  = MachineState(1) // Machine has been started but is not yet usable
	MachineStateReady    = MachineState(2) // Machine is running & servers are available
	MachineStateShutdown = MachineState(3) // Machine is going down
)

// Machine represents a single "computer" on which an optional agent, a coordinator and a dbserver runs.
type Machine interface {
	// State returns the current state of the machine
	State() MachineState

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
