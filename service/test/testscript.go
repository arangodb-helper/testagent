package test

import (
	"fmt"
	"time"

	"github.com/arangodb/testAgent/service/cluster"
)

type TestListener interface {
	ReportFailure(f Failure)
}

type TestScript interface {
	// Name returns the name of the script
	Name() string

	// Start triggers the test script to start.
	// It should spwan actions in a go routine.
	Start(cluster cluster.Cluster, listener TestListener) error

	// Stop any running test. This should not return until tests are actually stopped.
	Stop() error
}

type Failure struct {
	Timestamp time.Time
	Message   string
}

func NewFailure(message string, args ...interface{}) Failure {
	return Failure{
		Timestamp: time.Now(),
		Message:   fmt.Sprintf(message, args...),
	}
}
