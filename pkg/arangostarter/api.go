package arangostarter

type API interface {
	// PrepareSlave prepare the launch of an arangodb slave.
	// Given an ID, host, port & dataDir will it reserve a peer slot and return a stable port offset.
	PrepareSlave(id, hostIP string, port int, dataDir string) (Peer, error)

	// GetProcesses loads information of all the server processes launched by a specific arangodb.
	GetProcesses() (ProcessListResponse, error)

	// Shutdown will shutdown a specific arangodb (and all its started servers).
	// With goodbye set, it will remove the peer slot for this arangodb instance.
	Shutdown(goodbye bool) error

	// WaitUntilGone will block until the specific arangodb instance can no longer be reached.
	WaitUntilGone() error
}

type Peer struct {
	ID         string // Unique of of the peer
	Address    string // IP address of arangodb peer server
	Port       int    // Port number of arangodb peer server
	PortOffset int    // Offset to add to base ports for the various servers (agent, coordinator, dbserver)
	DataDir    string // Directory holding my data
	HasAgent   bool   // If set, this peer is running an agent
}

// Peer information.
// When this type (or any of the types used in here) is changed, increase `SetupConfigVersion`.
type Peers struct {
	Peers      []Peer // All peers (index 0 is reserver for the master)
	AgencySize int    // Number of agents
}

// PeerByID returns a peer with given id & true, or false if not found.
func (p Peers) PeerByID(id string) (Peer, bool) {
	for _, x := range p.Peers {
		if x.ID == id {
			return x, true
		}
	}
	return Peer{}, false
}

type ProcessListResponse struct {
	ServersStarted bool            `json:"servers-started,omitempty"` // True if the server have all been started
	Servers        []ServerProcess `json:"servers,omitempty"`         // List of servers started by ArangoDB
}

type ServerProcess struct {
	Type        string `json:"type"`                   // agent | coordinator | dbserver
	IP          string `json:"ip"`                     // IP address needed to reach the server
	Port        int    `json:"port"`                   // Port needed to reach the server
	ProcessID   int    `json:"pid,omitempty"`          // PID of the process (0 when running in docker)
	ContainerID string `json:"container-id,omitempty"` // ID of docker container running the server
}
