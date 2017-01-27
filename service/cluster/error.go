package cluster

import (
	"github.com/juju/errgo"
)

var (
	maskAny      = errgo.MaskFunc(errgo.Any)
	TimeoutError = errgo.New("timeout")
)
