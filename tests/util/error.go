package util

import "github.com/juju/errgo"

var (
	maskAny      = errgo.MaskFunc(errgo.Any)
	failureError = errgo.New("failure")
)

func isFailure(err error) bool {
	return errgo.Cause(err) == failureError
}
