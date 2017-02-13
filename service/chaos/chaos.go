package chaos

import (
	"context"
	"math/rand"
	"sync"
	"time"

	cluster "github.com/arangodb/testAgent/service/cluster"
	logging "github.com/op/go-logging"
)

type ChaosMonkey interface {
	// Active returns true when chaos is being introduced.
	Active() bool

	// Describe the state of the chaosmonkey
	State() string

	// Start introducing chaos
	Start()

	// Stop introducing chaos
	Stop()

	// WaitUntilInactive blocks until Active return false
	WaitUntilInactive()

	// Get a list of recent events
	GetRecentEvents(maxEvents int) []Event

	// Return all actions possible by the chaos monkey
	Actions() []Action
}

type ChaosMonkeyConfig struct {
	MaxMachines int // Maximum number of machines to allow in a cluster.
}

// NewChaosMonkey creates a new chaos monkey for the given cluster
func NewChaosMonkey(log *logging.Logger, cluster cluster.Cluster, config ChaosMonkeyConfig) ChaosMonkey {
	c := &chaosMonkey{
		ChaosMonkeyConfig: config,
		log:               log,
		cluster:           cluster,
	}
	c.actions = []*chaosAction{
		&chaosAction{c.restartAgent, "Restart Agent", 0, 0, 0, false},
		&chaosAction{c.restartDBServer, "Restart DBServer", 0, 0, 0, false},
		&chaosAction{c.restartCoordinator, "Restart Coordinator", 0, 0, 0, false},
		&chaosAction{c.killAgent, "Kill Agent", 0, 0, 0, false},
		&chaosAction{c.killDBServer, "Kill DBServer", 0, 0, 0, false},
		&chaosAction{c.killCoordinator, "Kill Coordinator", 0, 0, 0, false},
		&chaosAction{c.rebootMachine, "Reboot Machine", 0, 0, 0, false},
		&chaosAction{c.rejectAgentTraffic, "Reject Agent Traffic", 0, 0, 0, false},
		&chaosAction{c.rejectDBServerTraffic, "Reject DBServer Traffic", 0, 0, 0, false},
		&chaosAction{c.rejectCoordinatorTraffic, "Reject Coordinator Traffic", 0, 0, 0, false},
		&chaosAction{c.dropAgentTraffic, "Drop Agent Traffic", 0, 0, 0, false},
		&chaosAction{c.dropDBServerTraffic, "Drop DBServer Traffic", 0, 0, 0, false},
		&chaosAction{c.dropCoordinatorTraffic, "Drop Coordinator Traffic", 0, 0, 0, false},
		&chaosAction{c.addMachine, "Add New Machine", 0, 0, 0, true},
	}
	return c
}

type chaosMonkey struct {
	ChaosMonkeyConfig
	mutex        sync.Mutex
	log          *logging.Logger
	cluster      cluster.Cluster
	active       bool
	cancel       context.CancelFunc
	cancelled    bool
	recentEvents []Event // Limit list of events (last event first)
	actions      []*chaosAction
}

// Active returns true when chaos is being introduced.
func (c *chaosMonkey) Active() bool {
	return c.active
}

// Describe the state of the chaosmonkey
func (c *chaosMonkey) State() string {
	if c.active {
		if c.cancelled {
			return "stopping"
		}
		return "active"
	}
	return "inactive"
}

// Start introducing chaos
func (c *chaosMonkey) Start() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.active {
		ctx, cancel := context.WithCancel(context.Background())
		c.active = true
		c.cancelled = false
		c.cancel = cancel
		go c.chaosLoop(ctx)
	}
}

// Stop introducing chaos
func (c *chaosMonkey) Stop() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if cancel := c.cancel; cancel != nil {
		c.cancelled = true
		cancel()
	}
}

// WaitUntilInactive blocks until Active return false
func (c *chaosMonkey) WaitUntilInactive() {
	for {
		if !c.active {
			return
		}
		time.Sleep(time.Millisecond * 200)
	}
}

// Return all actions possible by the chaos monkey
func (c *chaosMonkey) Actions() []Action {
	var result []Action
	for _, a := range c.actions {
		result = append(result, a)
	}
	return result
}

// chaosLoop runs the process to actually introduce chaos
func (c *chaosMonkey) chaosLoop(ctx context.Context) {
	for {
		// Pick a random chaos action
		action := c.actions[rand.Intn(len(c.actions))]
		var delay time.Duration
		if action.Enabled() {
			if action.action(ctx, action) {
				// Chaos was introduced
				delay = time.Second * 30
			} else {
				// Chaos was not introduced, wait a bit shorter
				delay = time.Second * 2
			}
		} else {
			// Action disabled, pick a new one after a very short wait
			delay = time.Millisecond * 50
		}
		select {
		case <-ctx.Done():
			c.log.Debugf("stop signaled, terminating from chaosLoop")
			c.mutex.Lock()
			defer c.mutex.Unlock()
			c.active = false
			c.cancel = nil
			return
		case <-time.After(delay):
			// Continue looping
		}
	}
}
