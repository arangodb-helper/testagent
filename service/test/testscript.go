package test

import (
	"fmt"
	"io"
	"time"

	"github.com/arangodb-helper/testagent/service/cluster"
	"github.com/pkg/errors"
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
	// It should spawn actions in a go routine.
	Start(cluster cluster.Cluster, listener TestListener) error

	// Stop any running test. This should not return until tests are actually stopped.
	Stop() error

	// Interrupt the tests, but be prepared to continue.
	Pause() error

	// Resume running the tests, where Pause interrupted it.
	Resume() error

	// CollectLogs copies all logging info to the given writer.
	CollectLogs(io.Writer) error
}

type Failure struct {
	Timestamp time.Time
	Message   string
	Test      string
	Errors    []error
}

type AggregateError interface {
	Errors() []error
}

func NewFailure(testName string, message string, args ...interface{}) Failure {
	var errorList []error
	for _, x := range args {
		if err, ok := x.(error); ok {
			errorList = append(errorList, err)
			if aerr, ok := errors.Cause(err).(AggregateError); ok {
				errorList = append(errorList, aerr.Errors()...)
			}
		}
	}
	return Failure{
		Timestamp: time.Now(),
		Test:      testName,
		Message:   fmt.Sprintf(message, args...),
		Errors:    errorList,
	}
}

type TestStatus struct {
	Active   bool
	Pausing  bool
	Failures int
	Actions  int
	Messages []string
	Counters []Counter
}

type Counter struct {
	Name      string
	Succeeded int
	Failed    int
}
