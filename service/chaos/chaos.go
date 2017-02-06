package chaos

import (
	"context"
	"fmt"
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

	// Return all current statistics
	Statistics() []Statistic
}

type Statistic struct {
	Name  string
	Value int
}

// NewChaosMonkey creates a new chaos monkey for the given cluster
func NewChaosMonkey(log *logging.Logger, cluster cluster.Cluster) ChaosMonkey {
	c := &chaosMonkey{
		log:     log,
		cluster: cluster,
	}
	c.actions = []*chaosAction{
		&chaosAction{c.restartAgent, "Restart Agent", 0, 0, 0},
		&chaosAction{c.restartDBServer, "Restart DBServer", 0, 0, 0},
		&chaosAction{c.restartCoordinator, "Restart Coordinator", 0, 0, 0},
		&chaosAction{c.killAgent, "Kill Agent", 0, 0, 0},
		&chaosAction{c.killDBServer, "Kill DBServer", 0, 0, 0},
		&chaosAction{c.killCoordinator, "Kill Coordinator", 0, 0, 0},
		&chaosAction{c.rebootMachine, "Reboot Machine", 0, 0, 0},
		&chaosAction{c.rejectAgentTraffic, "Reject Agent Traffic", 0, 0, 0},
		&chaosAction{c.rejectDBServerTraffic, "Reject DBServer Traffic", 0, 0, 0},
		&chaosAction{c.rejectCoordinatorTraffic, "Reject Coordinator Traffic", 0, 0, 0},
		&chaosAction{c.dropAgentTraffic, "Drop Agent Traffic", 0, 0, 0},
		&chaosAction{c.dropDBServerTraffic, "Drop DBServer Traffic", 0, 0, 0},
		&chaosAction{c.dropCoordinatorTraffic, "Drop Coordinator Traffic", 0, 0, 0},
	}
	return c
}

type chaosMonkey struct {
	mutex        sync.Mutex
	log          *logging.Logger
	cluster      cluster.Cluster
	active       bool
	cancel       context.CancelFunc
	cancelled    bool
	recentEvents []Event // Limit list of events (last event first)
	actions      []*chaosAction
}

type chaosAction struct {
	action    func(context.Context, *chaosAction) bool
	name      string
	succeeded int
	failures  int
	skipped   int
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

// Return all current statistics
func (c *chaosMonkey) Statistics() []Statistic {
	result := make([]Statistic, 0, len(c.actions)*3)
	for _, action := range c.actions {
		result = append(result, Statistic{fmt.Sprintf("%s succeeded", action.name), action.succeeded})
	}
	for _, action := range c.actions {
		result = append(result, Statistic{fmt.Sprintf("%s failed", action.name), action.failures})
	}
	for _, action := range c.actions {
		result = append(result, Statistic{fmt.Sprintf("%s skipped", action.name), action.skipped})
	}
	return result
}

// chaosLoop runs the process to actually introduce chaos
func (c *chaosMonkey) chaosLoop(ctx context.Context) {
	for {
		// Pick a random chaos action
		action := c.actions[rand.Intn(len(c.actions))]
		var delay time.Duration
		if action.action(ctx, action) {
			// Chaos was introduced
			delay = time.Second * 30
		} else {
			// Chaos was not introduced, wait a bit shorter
			delay = time.Second * 2
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
