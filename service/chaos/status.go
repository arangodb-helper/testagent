package chaos

import (
	"sync/atomic"

	"github.com/arangodb/testAgent/service/cluster"
	"golang.org/x/sync/errgroup"
)

// checkAgencyReadyStatus checks that all agents in the cluster are ready.
// It returns: #readyAgents, agentMachines, error
func (c *chaosMonkey) checkAgencyReadyStatus() (int, []cluster.Machine, error) {
	machines, err := c.cluster.Machines()
	if err != nil {
		return 0, nil, maskAny(err)
	}
	readyAgents := int32(0)
	var agentMachines []cluster.Machine
	g := errgroup.Group{}
	for _, m := range machines {
		m := m
		if m.HasAgent() {
			agentMachines = append(agentMachines, m)
			g.Go(func() error {
				if err := m.TestAgentStatus(); err != nil {
					return maskAny(err)
				}
				atomic.AddInt32(&readyAgents, 1)
				return nil
			})
		}
	}
	if err := g.Wait(); err != nil {
		return 0, nil, maskAny(err)
	}

	return int(readyAgents), agentMachines, nil
}
