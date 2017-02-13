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
	return fmt.Sprintf("[%s] %s", e.Time.Format("2006-01-02 15:04:05"), e.Action)
}

// Get a list of recent events
func (c *chaosMonkey) GetRecentEvents(maxEvents int) []Event {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	source := c.recentEvents
	if maxEvents < len(source) {
		source = source[:maxEvents]
	}
	result := make([]Event, 0, maxEvents)
	return append(result, source...)
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
