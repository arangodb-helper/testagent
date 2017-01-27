package service

import "github.com/arangodb/testAgent/service/test"

// Report the given failure
func (s *Service) ReportFailure(f test.Failure) {
	// Collect cluster state
	// TODO

	// Collect recent chaos
	// TODO

	// Notify about failure
}
