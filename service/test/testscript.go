package test

import (
	"fmt"
	"io"
	"time"

	"github.com/arangodb/testAgent/service/cluster"
	"github.com/juju/errgo"
)

type TestListener interface {
	ReportFailure(f Failure)
}

type TestScript interface {
	// Name returns the name of the script
	Name() string

	// Status returns the current status of the test
	Status() TestStatus

	// Start triggers the test script to start.
	// It should spwan actions in a go routine.
	Start(cluster cluster.Cluster, listener TestListener) error

	// Stop any running test. This should not return until tests are actually stopped.
	Stop() error

	// CollectLogs copies all logging info to the given writer.
	CollectLogs(io.Writer) error
}

type Failure struct {
	Timestamp time.Time
	Message   string
	Errors    []error
}

type AggregateError interface {
	Errors() []error
}

func NewFailure(message string, args ...interface{}) Failure {
	var errors []error
	for _, x := range args {
		if err, ok := x.(error); ok {
			errors = append(errors, err)
			if aerr, ok := errgo.Cause(err).(AggregateError); ok {
				errors = append(errors, aerr.Errors()...)
			}
		}
	}
	return Failure{
		Timestamp: time.Now(),
		Message:   fmt.Sprintf(message, args...),
		Errors:    errors,
	}
}

type TestStatus struct {
	Failures int
	Actions  int
	Messages []string
}
