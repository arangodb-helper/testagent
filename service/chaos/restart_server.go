package chaos

import "math/rand"

// restartAgent randomly picks an agent and restarts it.
// Before doing so, it first checks if restarting an agent is allowed on the current cluster state.
func (c *chaosMonkey) restartAgent() bool {
	_, agentMachines, err := c.checkAgencyReadyStatus()
	if err != nil {
		c.log.Infof("Not all agents are ready (%s), so I cannot restart one now", err.Error())
		return false
	}
	if len(agentMachines) < 3 {
		c.log.Infof("There are too few (%d) agents in the cluster, so I cannot restart one now", len(agentMachines))
		return false
	}

	// Pick a random agent machine
	m := agentMachines[rand.Intn(len(agentMachines))]
	c.recordEvent(newEvent("Restarting agent on %s...", m.ID()))
	if err := m.RestartAgent(); err != nil {
		c.log.Errorf("Failed to restart agent: %v", err)
		c.recordEvent(newEvent("Restart agent on %s failed: %v", m.ID(), err))
	} else {
		c.recordEvent(newEvent("Restart agent on %s succeeded", m.ID()))
	}
	return true
}
