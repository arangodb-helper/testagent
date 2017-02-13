package cluster

import (
	"github.com/pkg/errors"
)

var (
	maskAny      = errors.WithStack
	TimeoutError = errors.New("timeout")
)
