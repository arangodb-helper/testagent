package util

import "github.com/pkg/errors"

var (
	maskAny      = errors.WithStack
	failureError = errors.New("failure")
)

func isFailure(err error) bool {
	return errors.Cause(err) == failureError
}
