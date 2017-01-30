package cluster

import (
	"fmt"
	"io"
	"net/url"
)

type ClusterBuilder interface {
	// Create creates and starts a new cluster.
	// The number of "machines" created equals the given agency size.
	// This function returns when the cluster is operational (or an error occurs)
	Create(agencySize int) (Cluster, error)
}

type Cluster interface {
	// ID returns a unique identifier for this cluster
	ID() string

	// Machines returns all current machines in the cluster.
	Machines() ([]Machine, error)

	// Remove the entire cluster
	Destroy() error
}

type MachineState int

const (
	MachineStateNew       = MachineState(0) // Initial state
	MachineStateStarted   = MachineState(1) // Machine has been started but is not yet usable
	MachineStateReady     = MachineState(2) // Machine is running & servers are available
	MachineStateShutdown  = MachineState(3) // Machine is going down
	MachineStateDestroyed = MachineState(4) // Machine destroyed beyond repair
)

func (s MachineState) String() string {
	switch s {
	case MachineStateNew:
		return "new"
	case MachineStateStarted:
		return "started"
	case MachineStateReady:
		return "ready"
	case MachineStateShutdown:
		return "shutdown"
	case MachineStateDestroyed:
		return "destroyed"
	default:
		return fmt.Sprintf("Unknown state %d", int(s))
	}
}

// Machine represents a single "computer" on which an optional agent, a coordinator and a dbserver runs.
type Machine interface {
	// ID returns a unique identifier for this machine
	ID() string
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

	// LastAgentReadyStatus returns true if the last known agent ready check succeeded.
	LastAgentReadyStatus() bool
	// LastDBServerReadyStatus returns true if the last known dbserver ready check succeeded.
	LastDBServerReadyStatus() bool
	// LastCoordinatorReadyStatus returns true if the last known coordinator ready check succeeded.
	LastCoordinatorReadyStatus() bool

	// TestAgentStatus checks if the agent on this machine is ready (with a reasonable timeout). If returns nil on ready, error on not ready.
	TestAgentStatus() error
	// TestDBServerStatus checks if the dbserver on this machine is ready (with a reasonable timeout). If returns nil on ready, error on not ready.
	TestDBServerStatus() error
	// TestCoordinatorStatus checks if the coordinator on this machine is ready (with a reasonable timeout). If returns nil on ready, error on not ready.
	TestCoordinatorStatus() error

	// Perform a graceful restart of the agent. This function does NOT wait until the agent is ready again.
	RestartAgent() error
	// Perform a graceful restart of the dbserver. This function does NOT wait until the dbserver is ready again.
	RestartDBServer() error
	// Perform a graceful restart of the coordinator. This function does NOT wait until the coordinator is ready again.
	RestartCoordinator() error

	// CollectAgentLogs collects recent logs from the agent and writes them to the given writer.
	CollectAgentLogs(w io.Writer) error
	// CollectDBServerLogs collects recent logs from the dbserver and writes them to the given writer.
	CollectDBServerLogs(w io.Writer) error
	// CollectCoordinatorLogs collects recent logs from the coordinator and writes them to the given writer.
	CollectCoordinatorLogs(w io.Writer) error

	// Reboot performs a graceful reboot of the machine
	Reboot() error

	// Remove the machine without the ability to recover it
	Destroy() error
}
