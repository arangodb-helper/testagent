package chaos

import "math/rand"

// killAgent randomly picks an agent and kills it (the hard way).
// Before doing so, it first checks if killing an agent is allowed on the current cluster state.
func (c *chaosMonkey) killAgent() bool {
	agentMachines, _, err := c.checkAgencyReadyStatus()
	if err != nil {
		c.log.Infof("Not all agents are ready (%s), so I cannot kill one now", err.Error())
		return false
	}
	if len(agentMachines) < 3 {
		c.log.Infof("There are too few (%d) agents in the cluster, so I cannot kill one now", len(agentMachines))
		return false
	}

	// Pick a random agent machine
	m := agentMachines[rand.Intn(len(agentMachines))]
	c.recordEvent(newEvent("Killing agent on %s...", m.ID()))
	if err := m.KillAgent(); err != nil {
		c.log.Errorf("Failed to kill agent: %v", err)
		c.recordEvent(newEvent("Kill agent on %s failed: %v", m.ID(), err))
	} else {
		c.recordEvent(newEvent("Kill agent on %s succeeded", m.ID()))
	}
	return true
}

// killDBServer randomly picks a dbserver and kills it.
// Before doing so, it first checks if killing a dbserver is allowed on the current cluster state.
func (c *chaosMonkey) killDBServer() bool {
	readyMachines, notReadyServers, err := c.checkDBServerReadyStatus()
	if err != nil {
		c.log.Infof("Failed to check dbserver ready status (%s), so I cannot kill one now", err.Error())
		return false
	}
	if notReadyServers > 0 {
		c.log.Infof("At least 1 dbserver is already down (%d down), so I cannot kill one now", notReadyServers)
		return false
	}
	if len(readyMachines) == 0 {
		c.log.Infof("There are no ready dbservers in the cluster, so I cannot kill one now")
		return false
	}

	// Pick a random dbserver machine
	m := readyMachines[rand.Intn(len(readyMachines))]
	c.recordEvent(newEvent("Killing dbserver on %s...", m.ID()))
	if err := m.KillDBServer(); err != nil {
		c.log.Errorf("Failed to kill dbserver: %v", err)
		c.recordEvent(newEvent("Kill dbserver on %s failed: %v", m.ID(), err))
	} else {
		c.recordEvent(newEvent("Kill dbserver on %s succeeded", m.ID()))
	}
	return true
}

// killCoordinator randomly picks a coordinator and kills it.
// Before doing so, it first checks if killing a coordinator is allowed on the current cluster state.
func (c *chaosMonkey) killCoordinator() bool {
	readyMachines, _, err := c.checkCoordinatorReadyStatus()
	if err != nil {
		c.log.Infof("Failed to check coordinator ready status (%s), so I cannot kill one now", err.Error())
		return false
	}
	if len(readyMachines) == 0 {
		c.log.Infof("There are no ready coordinators in the cluster, so I cannot kill one now")
		return false
	}

	// Pick a random coordinator machine
	m := readyMachines[rand.Intn(len(readyMachines))]
	c.recordEvent(newEvent("Killing coordinator on %s...", m.ID()))
	if err := m.KillCoordinator(); err != nil {
		c.log.Errorf("Failed to kill coordinator: %v", err)
		c.recordEvent(newEvent("Kill coordinator on %s failed: %v", m.ID(), err))
	} else {
		c.recordEvent(newEvent("Kill coordinator on %s succeeded", m.ID()))
	}
	return true
}
