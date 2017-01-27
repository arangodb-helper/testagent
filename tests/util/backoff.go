package util

import (
	"time"

	"github.com/cenkalti/backoff"
)

func retry(op func() error, timeout time.Duration) error {
	var failure error
	wrappedOp := func() error {
		if err := op(); err == nil {
			return nil
		} else if IsFailure(err) {
			// Detected failure
			failure = err
			return nil
		} else {
			return err
		}
	}
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = timeout
	b.MaxInterval = timeout / 3
	if err := backoff.Retry(wrappedOp, b); err != nil {
		return maskAny(err)
	}
	if failure != nil {
		return maskAny(failure)
	}
	return nil
}
