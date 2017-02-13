package chaos

import "context"

// addMachine adds a new machine.
// Before doing so, it first checks there are no too much machines already.
func (c *chaosMonkey) addMachine(ctx context.Context, action *chaosAction) bool {
	machines, err := c.cluster.Machines()
	if err != nil {
		c.log.Errorf("Failed to get machines: %v", err)
		return false
	}
	if len(machines) >= c.MaxMachines {
		c.log.Infof("Cannot have more than %d machines")
	}

	// Add a machine
	c.recordEvent(newEvent("Adding %d' machine...", len(machines)+1))
	if m, err := c.cluster.Add(); err != nil {
		c.log.Errorf("Failed to add machine: %v", err)
		action.failures++
		c.recordEvent(newEvent("Add new machine failed: %v", err))
	} else {
		action.succeeded++
		c.recordEvent(newEvent("Add new machine succeeded (added %s)", m.ID()))
	}
	return true
}
