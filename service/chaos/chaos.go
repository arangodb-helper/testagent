package chaos

import (
	"math/rand"
	"sync"
	"time"

	cluster "github.com/arangodb/testAgent/service/cluster"
	logging "github.com/op/go-logging"
)

type ChaosMonkey interface {
	// Active returns true when chaos is being introduced.
	Active() bool

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
	stop         bool
	recentEvents []Event // Limit list of events (last event first)
}

// Active returns true when chaos is being introduced.
func (c *chaosMonkey) Active() bool {
	return c.active
}

// Start introducing chaos
func (c *chaosMonkey) Start() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.stop = false
	if !c.active {
		c.active = true
		go c.chaosLoop()
	}
}

// Stop introducing chaos
func (c *chaosMonkey) Stop() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.stop = true
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
func (c *chaosMonkey) chaosLoop() {
	chaosActions := []func() bool{
		c.restartAgent,
	}
	for {
		if c.stop {
			c.log.Debugf("stop signaled, terminating from chaosLoop")
			c.active = false
			return
		}

		// Pick a random chaos action
		action := chaosActions[rand.Intn(len(chaosActions))]
		if action() {
			// Chaos was introduced
			time.Sleep(time.Second * 30)
		} else {
			// Chaos was not introducted, wait a bit shorter
			time.Sleep(time.Second * 2)
		}
	}
}
