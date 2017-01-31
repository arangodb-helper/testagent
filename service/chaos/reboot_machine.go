package chaos

import "math/rand"

// rebootMachine randomly picks a machine and restarts it gracefully.
// Before doing so, it first checks if reboot a machine is allowed on the current cluster state.
func (c *chaosMonkey) rebootMachine() bool {
	readyAgentMachines, _, _ := c.checkAgencyReadyStatus()
	readyDBServerMachines, notReadyDBServers, _ := c.checkDBServerReadyStatus()
	readyCoordinatorMachines, _, _ := c.checkCoordinatorReadyStatus()

	var rebootCandidates MachineList
	if len(readyAgentMachines) < 3 {
		// Find machines that do not have an agent
		rebootCandidates = readyDBServerMachines.Union(readyCoordinatorMachines).ExceptAgents()
	} else {
		// Just pick any machine that has at least 1 ready server
		rebootCandidates = readyDBServerMachines.Union(readyCoordinatorMachines).Union(readyAgentMachines)
	}

	if len(rebootCandidates) == 0 {
		c.log.Infof("There are 0 machines that can be rebooted")
		return false
	}
	if notReadyDBServers > 0 {
		c.log.Infof("At least 1 dbserver is already down (%d down), so I cannot reboot a machine now", notReadyDBServers)
		return false
	}
	if len(readyCoordinatorMachines) <= 1 {
		c.log.Infof("Only %d coordinator is ready, so I cannot reboot a machine now", len(readyCoordinatorMachines))
		return false
	}

	// Pick a random machine
	m := rebootCandidates[rand.Intn(len(rebootCandidates))]
	c.recordEvent(newEvent("Rebooting machine %s...", m.ID()))
	if err := m.Reboot(); err != nil {
		c.log.Errorf("Failed to reboot machine: %v", err)
		c.recordEvent(newEvent("Reboot machine %s failed: %v", m.ID(), err))
	} else {
		c.recordEvent(newEvent("Reboot machine %s succeeded", m.ID()))
	}
	return true
}
