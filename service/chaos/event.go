package chaos

import (
	"fmt"
	"time"
)

type Event struct {
	Time   time.Time // When did the event start
	Action string    // What happened
}

func (e Event) String() string {
	return fmt.Sprintf("[%s] %s", e.Time, e.Action)
}

// Get a list of recent events
func (c *chaosMonkey) GetRecentEvents() []Event {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var result []Event
	return append(result, c.recentEvents...)
}

// newEvent creates a new event with current time and given action.
func newEvent(action string, args ...interface{}) Event {
	return Event{
		Time:   time.Now(),
		Action: fmt.Sprintf(action, args...),
	}
}

// recordEvent adds an event to the front of the recentEvents list.
func (c *chaosMonkey) recordEvent(evt Event) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.recentEvents = append([]Event{evt}, c.recentEvents...)
	c.log.Infof("Recorded event %s", evt)
}
