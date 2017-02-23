package chaos

import (
	"context"
	"math/rand"
	"time"
)

// dropAgentTraffic randomly picks an agent and silently drops all network traffic to it it.
func (c *chaosMonkey) dropAgentTraffic(ctx context.Context, action *chaosAction) bool {
	if c.DisableNetworkChaos {
		return false
	}
	agentMachines, _, err := c.checkAgencyReadyStatus()
	if err != nil {
		c.log.Infof("Not all agents are ready (%s), so I cannot drop network traffic to one now", err.Error())
		action.skipped++
		return false
	}
	if len(agentMachines) < 3 {
		c.log.Infof("There are too few (%d) agents in the cluster, so I cannot drop network traffic to one now", len(agentMachines))
		action.skipped++
		return false
	}

	// Pick a random agent machine
	m := agentMachines[rand.Intn(len(agentMachines))]
	timeout := createNetworkTimeout()
	c.recordEvent(newEvent("Dropping network traffic to agent on %s for %s", m.ID(), timeout))
	if err := m.DropAgentTraffic(); err != nil {
		c.log.Errorf("Failed to drop network traffic to agent: %v", err)
		action.failures++
		c.recordEvent(newEvent("Dropping network traffic to agent on %s failed: %v", m.ID(), err))
		return false
	}

	// Wait a while before restoring network traffic
	select {
	case <-ctx.Done():
	case <-time.After(timeout):
	}

	// Cleanup
	action.succeeded++
	if err := m.AcceptAgentTraffic(); err != nil {
		c.recordEvent(newEvent("Restoring network traffic to agent on %s failed: %v", m.ID(), err))
	} else {
		c.recordEvent(newEvent("Restoring network traffic to agent on %s succeeded", m.ID()))
	}

	return true
}

// dropDBServerTraffic randomly picks an dbserver and silently drops all network traffic to it it.
func (c *chaosMonkey) dropDBServerTraffic(ctx context.Context, action *chaosAction) bool {
	if c.DisableNetworkChaos {
		return false
	}
	readyMachines, notReadyServers, err := c.checkDBServerReadyStatus()
	if err != nil {
		c.log.Infof("Failed to check dbserver ready status (%s), so I cannot drop network traffic to one now", err.Error())
		action.skipped++
		return false
	}
	if notReadyServers > 0 {
		c.log.Infof("At least 1 dbserver is already down (%d down), so I cannot drop network traffic to one now", notReadyServers)
		action.skipped++
		return false
	}
	if len(readyMachines) == 0 {
		c.log.Infof("There are no ready dbservers in the cluster, so I cannot drop network traffic to one now")
		action.skipped++
		return false
	}

	// Pick a random dbserver machine
	m := readyMachines[rand.Intn(len(readyMachines))]
	timeout := createNetworkTimeout()
	c.recordEvent(newEvent("Dropping network traffic to dbserver on %s for %s", m.ID(), timeout))
	if err := m.DropDBServerTraffic(); err != nil {
		c.log.Errorf("Failed to drop network traffic to dbserver: %v", err)
		action.failures++
		c.recordEvent(newEvent("Dropping network traffic to dbserver on %s failed: %v", m.ID(), err))
		return false
	}

	// Wait a while before restoring network traffic
	select {
	case <-ctx.Done():
	case <-time.After(createNetworkTimeout()):
	}

	// Cleanup
	action.succeeded++
	if err := m.AcceptDBServerTraffic(); err != nil {
		c.recordEvent(newEvent("Restoring network traffic to dbserver on %s failed: %v", m.ID(), err))
	} else {
		c.recordEvent(newEvent("Restoring network traffic to dbserver on %s succeeded", m.ID()))
	}

	return true
}

// dropCoordinatorTraffic randomly picks an coordinator and silently drops all network traffic to it it.
func (c *chaosMonkey) dropCoordinatorTraffic(ctx context.Context, action *chaosAction) bool {
	if c.DisableNetworkChaos {
		return false
	}
	readyMachines, _, err := c.checkCoordinatorReadyStatus()
	if err != nil {
		c.log.Infof("Failed to check coordinator ready status (%s), so I cannot drop network traffic to one now", err.Error())
		action.skipped++
		return false
	}
	if len(readyMachines) == 0 {
		c.log.Infof("There are no ready coordinators in the cluster, so I cannot drop network traffic to one now")
		action.skipped++
		return false
	}

	// Pick a random coordinator machine
	m := readyMachines[rand.Intn(len(readyMachines))]
	timeout := createNetworkTimeout()
	c.recordEvent(newEvent("Dropping network traffic to coordinator on %s for %s", m.ID(), timeout))
	if err := m.DropCoordinatorTraffic(); err != nil {
		c.log.Errorf("Failed to drop network traffic to coordinator: %v", err)
		action.failures++
		c.recordEvent(newEvent("Dropping network traffic to coordinator on %s failed: %v", m.ID(), err))
		return false
	}

	// Wait a while before restoring network traffic
	select {
	case <-ctx.Done():
	case <-time.After(timeout):
	}

	// Cleanup
	action.succeeded++
	if err := m.AcceptCoordinatorTraffic(); err != nil {
		c.recordEvent(newEvent("Restoring network traffic to coordinator on %s failed: %v", m.ID(), err))
	} else {
		c.recordEvent(newEvent("Restoring network traffic to coordinator on %s succeeded", m.ID()))
	}

	return true
}
