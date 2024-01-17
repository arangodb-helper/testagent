package chaos

import (
	"context"
	"crypto/sha1"
	"fmt"
)

type Action interface {
	ID() string
	Name() string
	Succeeded() int
	Failed() int
	Skipped() int
	Enabled() bool

	Enable()
	Disable()
}

type chaosAction struct {
	action       func(context.Context, *chaosAction) bool
	name         string
	succeeded    int
	failures     int
	skipped      int
	disabled     bool
	minimumLevel int
}

func (a *chaosAction) ID() string {
	hash := sha1.Sum([]byte(a.name))
	return fmt.Sprintf("%x", hash[:6])
}

func (a *chaosAction) Name() string {
	return a.name
}

func (a *chaosAction) Succeeded() int {
	return a.succeeded
}

func (a *chaosAction) Failed() int {
	return a.failures
}

func (a *chaosAction) Skipped() int {
	return a.skipped
}

func (a *chaosAction) Enabled() bool {
	return !a.disabled
}

func (a *chaosAction) Enable() {
	a.disabled = false
}

func (a *chaosAction) Disable() {
	a.disabled = true
}
