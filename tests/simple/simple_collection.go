package simple

import (
	"fmt"
	"time"
	"github.com/arangodb-helper/testagent/service/test"
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
	operationTimeout := time.Minute * 5
	retryTimeout := time.Minute * 5
	t.log.Infof("Creating collection '%s' with numberOfShards=%d, replicationFactor=%d...", c.name, numberOfShards, replicationFactor)
	for i := 0; i < 3; i++ {
		if resp, err := t.client.Post("/_api/collection", nil, nil, opts, "", nil, []int{200, 409}, []int{400, 404, 307, 500}, operationTimeout, retryTimeout); err != nil {
			// This is a failure
			t.reportFailure(test.NewFailure("Failed to create collection '%s': %v", c.name, err))
			return maskAny(err)
		} else if resp.StatusCode == 409 {
			// Duplicate name, check if that is correct
			if exists, checkErr := t.collectionExists(c); checkErr != nil {
				t.log.Errorf("Failed to check if collection exists: %v", checkErr)
				t.reportFailure(test.NewFailure("Failed to create collection '%s': %v and cannot check existance: %v", c.name, err, checkErr))
				return maskAny(err)
			} else if !exists {
				// Collection has not been created, so 409 status is really wrong
				t.reportFailure(test.NewFailure("Failed to create collection '%s': 409 reported but collection does not exist", c.name))
				return maskAny(fmt.Errorf("Create collection reported 409, but collection does not exist"))
			}
		}
	}
	t.log.Infof("Creating collection '%s' with numberOfShards=%d, replicationFactor=%d succeeded", c.name, numberOfShards, replicationFactor)
	return nil
}

// removeCollection remove an existing collection.
// The operation is expected to succeed.
func (t *simpleTest) removeExistingCollection(c *collection) error {
	operationTimeout, retryTimeout := t.OperationTimeout, t.RetryTimeout
	t.log.Infof("Removing collection '%s'...", c.name)
	resp, err := t.client.Delete("/_api/collection/"+c.name, nil, nil, []int{200, 404}, []int{400, 409, 307}, operationTimeout, retryTimeout)
	if err != nil {
		// This is a failure
		t.removeExistingCollectionCounter.failed++
		t.reportFailure(test.NewFailure("Failed to remove collection '%s': %v", c.name, err))
		return maskAny(err)
	} else if resp.StatusCode == 404 {
		// Collection not found.
		// This can happen if the first attempt timed out, but did actually succeed.
		// So we accept this is there are multiple attempts.
		if resp.Attempts <= 1 {
			// Not enough attempts, this is a failure
			t.removeExistingCollectionCounter.failed++
			t.reportFailure(test.NewFailure("Failed to remove collection '%s': got 404 after only 1 attempt", c.name))
			return maskAny(fmt.Errorf("Failed to remove collection '%s': got 404 after only 1 attempt", c.name))
		}
	}
	t.removeExistingCollectionCounter.succeeded++
	t.log.Infof("Removing collection '%s' succeeded", c.name)
	t.unregisterCollection(c)
	return nil
}

// collectionExists tries to fetch information about the collection to see if it exists.
func (t *simpleTest) collectionExists(c *collection) (bool, error) {
	operationTimeout, retryTimeout := t.OperationTimeout, t.RetryTimeout
	t.log.Infof("Checking collection '%s'...", c.name)
	if resp, err := t.client.Get("/_api/collection/"+c.name, nil, nil, nil, []int{200, 404}, []int{400, 409, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		return false, maskAny(err)
	} else if resp.StatusCode == 404 {
		// Not found
		return false, nil
	} else {
		// Found
		return true, nil
	}
}
