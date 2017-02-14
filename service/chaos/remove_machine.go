package chaos

import (
	"context"
	"math/rand"
)

// removeMachine randomly picks a machine and removes it gracefully.
// Before doing so, it first checks if removing a machine is allowed on the current cluster state.
func (c *chaosMonkey) removeMachine(ctx context.Context, action *chaosAction) bool {
	readyDBServerMachines, notReadyDBServers, _ := c.checkDBServerReadyStatus()
	readyCoordinatorMachines, _, _ := c.checkCoordinatorReadyStatus()

	// Find machines that do not have an agent
	removeCandidates := readyDBServerMachines.Intersection(readyCoordinatorMachines).ExceptAgents()

	if len(removeCandidates) == 0 {
		c.log.Infof("There are 0 machines that can be removed")
		action.skipped++
		return false
	}
	if notReadyDBServers > 0 {
		c.log.Infof("At least 1 dbserver is already down (%d down), so I cannot remove a machine now", notReadyDBServers)
		action.skipped++
		return false
	}
	if len(readyCoordinatorMachines) <= 1 {
		c.log.Infof("Only %d coordinator is ready, so I cannot remove a machine now", len(readyCoordinatorMachines))
		action.skipped++
		return false
	}

	// Pick a random machine
	m := removeCandidates[rand.Intn(len(removeCandidates))]
	c.recordEvent(newEvent("Removing machine %s...", m.ID()))
	if err := m.Destroy(); err != nil {
		c.log.Errorf("Failed to remove machine: %v", err)
		action.failures++
		c.recordEvent(newEvent("Remove machine %s failed: %v", m.ID(), err))
	} else {
		action.succeeded++
		c.recordEvent(newEvent("Remove machine %s succeeded", m.ID()))
	}
	return true
}
