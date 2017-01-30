package chaos

import (
	"sync"

	"github.com/arangodb/testAgent/service/cluster"
	"golang.org/x/sync/errgroup"
)

// checkAgencyReadyStatus checks that all agents in the cluster are ready.
// It returns: readyAgentMachines, #notReadyAgents, error
func (c *chaosMonkey) checkAgencyReadyStatus() ([]cluster.Machine, int, error) {
	machines, err := c.cluster.Machines()
	if err != nil {
		return nil, 0, maskAny(err)
	}
	var agentMachines []cluster.Machine
	g := errgroup.Group{}
	for _, m := range machines {
		m := m // Used in nested func
		if m.HasAgent() {
			agentMachines = append(agentMachines, m)
			g.Go(func() error {
				if err := m.TestAgentStatus(); err != nil {
					return maskAny(err)
				}
				return nil
			})
		}
	}
	if err := g.Wait(); err != nil {
		return nil, 0, maskAny(err)
	}

	return agentMachines, len(machines) - len(agentMachines), nil
}

// checkDBServerReadyStatus checks that all DBServers in the cluster are ready.
// It returns: readyDBServerMachines, #notReadyDBServerMachines error
func (c *chaosMonkey) checkDBServerReadyStatus() ([]cluster.Machine, int, error) {
	machines, err := c.cluster.Machines()
	if err != nil {
		return nil, 0, maskAny(err)
	}
	var mutex sync.Mutex
	var readyMachines []cluster.Machine
	g := errgroup.Group{}
	for _, m := range machines {
		m := m // Used in nested func
		g.Go(func() error {
			if err := m.TestDBServerStatus(); err == nil {
				mutex.Lock()
				defer mutex.Unlock()
				readyMachines = append(readyMachines, m)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, 0, maskAny(err)
	}

	return readyMachines, len(machines) - len(readyMachines), nil
}

// checkCoordinatorReadyStatus checks that all Coordinators in the cluster are ready.
// It returns: readyCoordinatorMachines, #notReadyCoordinatorMachines error
func (c *chaosMonkey) checkCoordinatorReadyStatus() ([]cluster.Machine, int, error) {
	machines, err := c.cluster.Machines()
	if err != nil {
		return nil, 0, maskAny(err)
	}
	var mutex sync.Mutex
	var readyMachines []cluster.Machine
	g := errgroup.Group{}
	for _, m := range machines {
		m := m // Used in nested func
		g.Go(func() error {
			if err := m.TestCoordinatorStatus(); err == nil {
				mutex.Lock()
				defer mutex.Unlock()
				readyMachines = append(readyMachines, m)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, 0, maskAny(err)
	}

	return readyMachines, len(machines) - len(readyMachines), nil
}
