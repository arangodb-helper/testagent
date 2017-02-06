package chaos

import (
	"context"
	"math/rand"
)

// restartAgent randomly picks an agent and restarts it.
// Before doing so, it first checks if restarting an agent is allowed on the current cluster state.
func (c *chaosMonkey) restartAgent(ctx context.Context, action *chaosAction) bool {
	agentMachines, _, err := c.checkAgencyReadyStatus()
	if err != nil {
		c.log.Infof("Not all agents are ready (%s), so I cannot restart one now", err.Error())
		action.skipped++
		return false
	}
	if len(agentMachines) < 3 {
		c.log.Infof("There are too few (%d) agents in the cluster, so I cannot restart one now", len(agentMachines))
		action.skipped++
		return false
	}

	// Pick a random agent machine
	m := agentMachines[rand.Intn(len(agentMachines))]
	c.recordEvent(newEvent("Restarting agent on %s...", m.ID()))
	if err := m.RestartAgent(); err != nil {
		c.log.Errorf("Failed to restart agent: %v", err)
		action.failures++
		c.recordEvent(newEvent("Restart agent on %s failed: %v", m.ID(), err))
	} else {
		action.succeeded++
		c.recordEvent(newEvent("Restart agent on %s succeeded", m.ID()))
	}
	return true
}

// restartDBServer randomly picks a dbserver and restarts it.
// Before doing so, it first checks if restarting a dbserver is allowed on the current cluster state.
func (c *chaosMonkey) restartDBServer(ctx context.Context, action *chaosAction) bool {
	readyMachines, notReadyServers, err := c.checkDBServerReadyStatus()
	if err != nil {
		c.log.Infof("Failed to check dbserver ready status (%s), so I cannot restart one now", err.Error())
		action.skipped++
		return false
	}
	if notReadyServers > 0 {
		c.log.Infof("At least 1 dbserver is already down (%d down), so I cannot restart one now", notReadyServers)
		action.skipped++
		return false
	}
	if len(readyMachines) == 0 {
		c.log.Infof("There are no ready dbservers in the cluster, so I cannot restart one now")
		action.skipped++
		return false
	}

	// Pick a random dbserver machine
	m := readyMachines[rand.Intn(len(readyMachines))]
	c.recordEvent(newEvent("Restarting dbserver on %s...", m.ID()))
	if err := m.RestartDBServer(); err != nil {
		c.log.Errorf("Failed to restart dbserver: %v", err)
		action.failures++
		c.recordEvent(newEvent("Restart dbserver on %s failed: %v", m.ID(), err))
	} else {
		action.succeeded++
		c.recordEvent(newEvent("Restart dbserver on %s succeeded", m.ID()))
	}
	return true
}

// restartCoordinator randomly picks a coordinator and restarts it.
// Before doing so, it first checks if restarting a coordinator is allowed on the current cluster state.
func (c *chaosMonkey) restartCoordinator(ctx context.Context, action *chaosAction) bool {
	readyMachines, _, err := c.checkCoordinatorReadyStatus()
	if err != nil {
		c.log.Infof("Failed to check coordinator ready status (%s), so I cannot restart one now", err.Error())
		action.skipped++
		return false
	}
	if len(readyMachines) == 0 {
		c.log.Infof("There are no ready coordinators in the cluster, so I cannot restart one now")
		action.skipped++
		return false
	}

	// Pick a random coordinator machine
	m := readyMachines[rand.Intn(len(readyMachines))]
	c.recordEvent(newEvent("Restarting coordinator on %s...", m.ID()))
	if err := m.RestartCoordinator(); err != nil {
		c.log.Errorf("Failed to restart coordinator: %v", err)
		action.failures++
		c.recordEvent(newEvent("Restart coordinator on %s failed: %v", m.ID(), err))
	} else {
		action.succeeded++
		c.recordEvent(newEvent("Restart coordinator on %s succeeded", m.ID()))
	}
	return true
}
