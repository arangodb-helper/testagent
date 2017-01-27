package util

import "github.com/juju/errgo"

var (
	maskAny      = errgo.MaskFunc(errgo.Any)
	FailureError = errgo.New("failure")
)

func IsFailure(err error) bool {
	return errgo.Cause(err) == FailureError
}
