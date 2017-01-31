package util

import (
	"time"

	"github.com/cenkalti/backoff"
)

type aggregateError struct {
	errors []error
}

func (aerr *aggregateError) Error() string {
	l := len(aerr.errors)
	if l == 0 {
		return "no error"
	}
	return aerr.errors[l-1].Error()
}

func (aerr *aggregateError) Errors() []error {
	return aerr.errors
}

func retry(op func() error, timeout time.Duration) error {
	var failure error
	aerr := &aggregateError{}
	wrappedOp := func() error {
		if err := op(); err == nil {
			return nil
		} else {
			aerr.errors = append(aerr.errors, err)
			if isFailure(err) {
				// Detected failure
				failure = err
				return nil
			} else {
				return err
			}
		}
	}
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = timeout
	b.MaxInterval = timeout / 3
	if err := backoff.Retry(wrappedOp, b); err != nil {
		return maskAny(aerr) // Note that we return the aggregated error!
	}
	if failure != nil {
		return maskAny(aerr) // Note that we return the aggregated error!
	}
	return nil
}
