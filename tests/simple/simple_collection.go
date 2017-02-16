package simple

import (
	"time"

	"github.com/arangodb/testAgent/service/test"
)

// createCollection creates a new collection.
// The operation is expected to succeed.
func (t *simpleTest) createCollection(name string, numberOfShards, replicationFactor int) error {
	opts := struct {
		Name              string `json:"name"`
		NumberOfShards    int    `json:"numberOfShards"`
		ReplicationFactor int    `json:"replicationFactor"`
	}{
		Name:              name,
		NumberOfShards:    numberOfShards,
		ReplicationFactor: replicationFactor,
	}
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
	t.log.Infof("Creating collection '%s' with numberOfShards=%d, replicationFactor=%d...", name, numberOfShards, replicationFactor)
	if _, err := t.client.Post("/_api/collection", nil, nil, opts, "", nil, []int{200}, []int{400, 404, 409, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.reportFailure(test.NewFailure("Failed to create collection '%s': %v", name, err))
		return maskAny(err)
	}
	t.log.Infof("Creating collection '%s' with numberOfShards=%d, replicationFactor=%d succeeded", name, numberOfShards, replicationFactor)
	return nil
}

// removeCollection remove an existing collection.
// The operation is expected to succeed.
func (t *simpleTest) removeExistingCollection(c *collection) error {
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
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
