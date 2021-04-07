package simple

import (
	"fmt"
	"github.com/arangodb-helper/testagent/service/test"
	"time"
)

// createCollection creates a new collection.
// The operation is expected to succeed.
func (t *simpleTest) createCollection(c *collection, numberOfShards, replicationFactor int) error {
	opts := struct {
		Name              string `json:"name"`
		NumberOfShards    int    `json:"numberOfShards"`
		ReplicationFactor int    `json:"replicationFactor"`
	}{
		Name:              c.name,
		NumberOfShards:    numberOfShards,
		ReplicationFactor: replicationFactor,
	}
	//operationTimeout, retryTimeout := t.OperationTimeout, t.RetryTimeout
	// For now, we increase the timeout to 5 minutes, since the cluster-internal
	// timeout is 4 minutes:
	operationTimeout := time.Minute
	testTimeout := time.Now().Add(operationTimeout * 5)

	backoff := time.Millisecond * 250
	i := 0

	for {

		i++
		if time.Now().After(testTimeout) {
			break;
		}

		checkRetry := false
		success := false
		shouldNotBeThere := false

		t.log.Infof("Creating (%d) collection '%s' with numberOfShards=%d, replicationFactor=%d...",
			i, c.name, numberOfShards, replicationFactor)
		resp, err := t.client.Post(
			"/_api/collection", nil, nil, opts, "", nil, []int{0, 1, 200, 409, 500, 503},
			[]int{400, 404, 307}, operationTimeout, 1)

		// 200: good
		// 500: collection couldn't be finished. most likely, because

		if err[0] == nil {
			if resp[0].StatusCode == 503 || resp[0].StatusCode == 409 || resp[0].StatusCode == 0 {
				checkRetry = true
			} else if resp[0].StatusCode == 1 || resp[0].StatusCode == 500 { // connection refused or not created
				shouldNotBeThere = true
			} else {
				success = true
			}
		} else {
			// This is a failure
			t.createCollectionCounter.failed++
			t.reportFailure(test.NewFailure("Failed to create collection '%s': %v", c.name, err[0]))
			return maskAny(err[0])
		}

		if checkRetry {
			if exists, checkErr := t.collectionExists(c); checkErr == nil { // TODO collectionExists retries
				if shouldNotBeThere {
					// Collection has been created, although it shouldn't
					t.createCollectionCounter.failed++
					t.reportFailure(
						test.NewFailure(
							"Failure while creating collection '%s': %d reported but collection exists",
							c.name, resp[0].StatusCode))
					return maskAny(
						fmt.Errorf("Failure while creating collection '%s': %d reported but collection exists",
							c.name, resp[0].StatusCode))
				} else {
					success = true
				}
			} else if !exists {
				if resp[0].StatusCode == 409 {
					// Collection has not been created, so 409 status is really wrong
					t.createCollectionCounter.failed++
					t.reportFailure(
						test.NewFailure(
							"Failed to create collection '%s': 409 reported but collection does not exist", c.name))
					return maskAny(
						fmt.Errorf("Create collection for '%s' reported 409, but collection does not exist", c.name))
				}
			}
		}

		if success {
			t.createCollectionCounter.succeeded++
			t.log.Infof(
				"Creating collection '%s' with numberOfShards=%d, replicationFactor=%d succeeded",
				c.name, numberOfShards, replicationFactor)
			return nil
		}

		t.log.Infof("Creating (%d) collection '%s' with numberOfShards=%d, replicationFactor=%d got %d",
			i, c.name, numberOfShards, replicationFactor, resp[0].StatusCode)
		time.Sleep(backoff)
		if backoff < time.Second * 5 {
			backoff += backoff
		}

	}

	// Overall timeout :(
	t.reportFailure(
		test.NewFailure("Timed out while trying to create (%d) collection %s.", i, c.name))
	return maskAny(fmt.Errorf("Timed out while trying to create (%d) collection %s.", i, c.name))

}

// removeCollection remove an existing collection.
// The operation is expected to succeed.
func (t *simpleTest) removeExistingCollection(c *collection) error {

	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(t.OperationTimeout * 5)

	url := fmt.Sprintf("/_api/collection/%s", c.name)
	backoff := time.Millisecond * 250
	i := 0

	for {

		i++
		if time.Now().After(testTimeout) {
			break
		}

		t.log.Infof("Removing (%d) collection '%s'...", i, c.name)
		resp, err := t.client.Delete(
			url, nil, nil, []int{0, 1, 200, 404, 503}, []int{400, 409, 307}, operationTimeout, 1)

		if err[0] != nil {
			// This is a failure
			t.removeExistingCollectionCounter.failed++
			t.reportFailure(test.NewFailure("Failed to remove collection '%s': %v", c.name, err[0]))
			return maskAny(err[0])
		} else if resp[0].StatusCode == 404 {
			// Collection not found.
			// This can happen if the first attempt timed out, but did actually succeed.
			// So we accept this if there are multiple attempts.
			if i == 1 { // this is a failure in first run
				// Not enough attempts, this is a failure
				t.removeExistingCollectionCounter.failed++
				t.reportFailure(
					test.NewFailure("Failed to remove collection '%s': got 404 after only 1 attempt", c.name))
				return maskAny(fmt.Errorf("Failed to remove collection '%s': got 404 after only 1 attempt", c.name))
			}
		} else if resp[0].StatusCode == 200 {
			t.removeExistingCollectionCounter.succeeded++
			t.log.Infof("Removing collection '%s' succeeded", c.name)
			t.unregisterCollection(c)
			return nil
		}

		t.log.Infof("Removing (%d) collection '%s' got %d", i, c.name, resp[0].StatusCode)
		time.Sleep(backoff)
		if backoff < time.Second * 5 {
			backoff += backoff
		}

	}

	t.removeExistingCollectionCounter.failed++
	t.reportFailure(test.NewFailure("Timed out (%d) while removing collection '%s'", i, c.name))
	return maskAny(fmt.Errorf("Timed out (%d) while removing collection '%s'", i, c.name))

}

// collectionExists tries to fetch information about the collection to see if it exists.
func (t *simpleTest) collectionExists(c *collection) (bool, error) {

	operationTimeout := t.OperationTimeout / 4
	timeout := time.Now().Add(operationTimeout)

	i := 0
	backoff := time.Millisecond * 250
	url := fmt.Sprintf("/_api/collection/%s", c.name)

	for  {

		i++
		if time.Now().After(timeout) {
			break;
		}

		t.log.Infof("Checking (%d) collection '%s'...", i, c.name)
		resp, err := t.client.Get(
			url , nil, nil, nil, []int{0, 1, 200, 404, 503}, []int{400, 409, 307}, operationTimeout, 1)

		if err[0] != nil {
			// This is a failure
			t.log.Infof("Failed checking for collection '%s': %v", c.name, err[0])
			return false, maskAny(err[0])
		} else if resp[0].StatusCode == 404 {
			return false, nil
		} else if resp[0].StatusCode == 200 {
			return true, nil
		}

		// 0, 1, 503 retry

		t.log.Infof("Checking (%d) collection '%s' got %d", i, c.name, resp[0].StatusCode)
		time.Sleep(backoff)
		if backoff < time.Second * 5 {
			backoff += backoff
		}

	}

	// This is a failure
	out := fmt.Errorf("Timed out checking for collection '%s'", c.name)
	t.log.Error(out)
	return false, maskAny(out)

}
