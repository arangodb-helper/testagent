package server

import (
	"github.com/arangodb/testAgent/service/chaos"
	"github.com/arangodb/testAgent/service/cluster"
	"github.com/arangodb/testAgent/service/test"
)

type Machine struct {
	ID                         string
	AgentURL                   string
	DBServerURL                string
	CoordinatorURL             string
	HasAgent                   bool
	LastAgentReadyStatus       bool
	LastDBServerReadyStatus    bool
	LastCoordinatorReadyStatus bool
}

type Test struct {
	Name     string
	Failures int
	Messages []string
}

type Chaos struct {
	Active bool
	Events []chaos.Event
}

const (
	maxChaosEvents = 20
)

func machineFromCluster(cm cluster.Machine) Machine {
	aURL := cm.AgentURL()
	dURL := cm.DBServerURL()
	cURL := cm.CoordinatorURL()
	return Machine{
		ID:                         cm.ID(),
		HasAgent:                   cm.HasAgent(),
		AgentURL:                   aURL.String(),
		DBServerURL:                dURL.String(),
		CoordinatorURL:             cURL.String(),
		LastAgentReadyStatus:       cm.LastAgentReadyStatus(),
		LastDBServerReadyStatus:    cm.LastDBServerReadyStatus(),
		LastCoordinatorReadyStatus: cm.LastCoordinatorReadyStatus(),
	}
}

type machineByID []Machine

// Len is the number of elements in the collection.
func (l machineByID) Len() int { return len(l) }

// Less reports whether the element with
// index i should sort before the element with index j.
func (l machineByID) Less(i, j int) bool { return l[i].ID < l[j].ID }

// Swap swaps the elements with indexes i and j.
func (l machineByID) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

func testFromTestScript(ct test.TestScript) Test {
	status := ct.Status()
	return Test{
		Name:     ct.Name(),
		Failures: status.Failures,
		Messages: status.Messages,
	}
}
