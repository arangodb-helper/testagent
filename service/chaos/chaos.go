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
	GetRecentEvents() []Event
}

// NewChaosMonkey creates a new chaos monkey for the given cluster
func NewChaosMonkey(log *logging.Logger, cluster cluster.Cluster) ChaosMonkey {
	return &chaosMonkey{
		log:     log,
		cluster: cluster,
	}
}

type chaosMonkey struct {
	mutex        sync.Mutex
	log          *logging.Logger
	cluster      cluster.Cluster
	active       bool
	cancel       context.CancelFunc
	cancelled    bool
	recentEvents []Event // Limit list of events (last event first)
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

// chaosLoop runs the process to actually introduce chaos
func (c *chaosMonkey) chaosLoop(ctx context.Context) {
	chaosActions := []func() bool{
		c.restartAgent,
		c.restartDBServer,
		c.restartCoordinator,
		c.killAgent,
		c.killDBServer,
		c.killCoordinator,
		c.rebootMachine,
	}
	for {
		// Pick a random chaos action
		action := chaosActions[rand.Intn(len(chaosActions))]
		var delay time.Duration
		if action() {
			// Chaos was introduced
			delay = time.Second * 30
		} else {
			// Chaos was not introducted, wait a bit shorter
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
