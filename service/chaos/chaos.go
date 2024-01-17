package chaos

import (
	"context"
	"math/rand"
	"sync"
	"time"

	cluster "github.com/arangodb-helper/testagent/service/cluster"
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

	// Set chaos level
	SetChaosLevel(level int)

	// Get chaos level
	Level() int

	// WaitUntilInactive blocks until Active return false
	WaitUntilInactive()

	// Get a list of recent events
	GetRecentEvents(maxEvents int) []Event

	// Return all actions possible by the chaos monkey
	Actions() []Action
}

type ChaosMonkeyConfig struct {
	MaxMachines         int  // Maximum number of machines to allow in a cluster.
	DisableNetworkChaos bool // If set to true, no network chaos is ever introduced
	ChaosLevel          int  // Chaos level
}

// NewChaosMonkey creates a new chaos monkey for the given cluster
func NewChaosMonkey(log *logging.Logger, cluster cluster.Cluster, config ChaosMonkeyConfig) ChaosMonkey {
	c := &chaosMonkey{
		ChaosMonkeyConfig: config,
		log:               log,
		cluster:           cluster,
	}
	c.actions = []*chaosAction{
		&chaosAction{c.restartAgent, "Restart Agent", 0, 0, 0, true, 1},
		&chaosAction{c.restartDBServer, "Restart DBServer", 0, 0, 0, true, 1},
		&chaosAction{c.restartCoordinator, "Restart Coordinator", 0, 0, 0, true, 1},
		&chaosAction{c.killAgent, "Kill Agent", 0, 0, 0, true, 2},
		&chaosAction{c.killDBServer, "Kill DBServer", 0, 0, 0, true, 2},
		&chaosAction{c.killCoordinator, "Kill Coordinator", 0, 0, 0, true, 2},
		&chaosAction{c.rebootMachine, "Reboot Machine", 0, 0, 0, true, 3},
		&chaosAction{c.addMachine, "Add New Machine", 0, 0, 0, true, 3},
		&chaosAction{c.removeMachine, "Remove Machine", 0, 0, 0, true, 3},
		&chaosAction{c.rejectAgentTraffic, "Reject Agent Traffic", 0, 0, 0, true, 4},
		&chaosAction{c.rejectDBServerTraffic, "Reject DBServer Traffic", 0, 0, 0, true, 4},
		&chaosAction{c.rejectCoordinatorTraffic, "Reject Coordinator Traffic", 0, 0, 0, true, 4},
		&chaosAction{c.dropAgentTraffic, "Drop Agent Traffic", 0, 0, 0, true, 4},
		&chaosAction{c.dropDBServerTraffic, "Drop DBServer Traffic", 0, 0, 0, true, 4},
		&chaosAction{c.dropCoordinatorTraffic, "Drop Coordinator Traffic", 0, 0, 0, true, 4},
	}
	c.applyChaosLevel()
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

func (c *chaosMonkey) applyChaosLevel() {
	for i := 0; i < len(c.actions); i++ {
		c.actions[i].disabled = c.actions[i].minimumLevel > c.ChaosLevel
	}
}

// Set chaos level
func (c *chaosMonkey) SetChaosLevel(level int) {
	c.log.Debugf("Setting chaos level: %d", level)
	c.ChaosLevel = level
	c.applyChaosLevel()
}

// Get chaos level
func (c *chaosMonkey) Level() int {
	return c.ChaosLevel
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
