package simple

import (
	"fmt"

	"github.com/arangodb/testAgent/service/test"
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
	operationTimeout, retryTimeout := t.OperationTimeout, t.RetryTimeout
	t.log.Infof("Creating collection '%s' with numberOfShards=%d, replicationFactor=%d...", c.name, numberOfShards, replicationFactor)
	if resp, err := t.client.Post("/_api/collection", nil, nil, opts, "", nil, []int{200, 409}, []int{400, 404, 307}, operationTimeout, retryTimeout); err != nil {
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
	t.log.Infof("Creating collection '%s' with numberOfShards=%d, replicationFactor=%d succeeded", c.name, numberOfShards, replicationFactor)
	return nil
}

// removeCollection remove an existing collection.
// The operation is expected to succeed.
func (t *simpleTest) removeExistingCollection(c *collection) error {
	operationTimeout, retryTimeout := t.OperationTimeout, t.RetryTimeout
	t.log.Infof("Removing collection '%s'...", c.name)
	if _, err := t.client.Delete("/_api/collection/"+c.name, nil, nil, []int{200}, []int{400, 404, 409, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.removeExistingCollectionCounter.failed++
		t.reportFailure(test.NewFailure("Failed to remove collection '%s': %v", c.name, err))
		return maskAny(err)
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
