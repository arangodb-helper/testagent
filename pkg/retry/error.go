package retry

import "github.com/pkg/errors"

var (
	maskAny      = errors.WithStack
	FailureError = errors.New("failure")
)

func isFailure(err error) bool {
	return errors.Cause(err) == FailureError
}
