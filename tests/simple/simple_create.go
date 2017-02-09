package simple

import (
	"fmt"
	"net/url"
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

// createDocument creates a new document.
// The operation is expected to succeed.
func (t *simpleTest) createDocument(collectionName string, document interface{}, key string) (string, error) {
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
	q := url.Values{}
	q.Set("waitForSync", "true")
	t.log.Infof("Creating document '%s' in '%s'...", key, collectionName)
	update, err := t.client.Post(fmt.Sprintf("/_api/document/%s", collectionName), q, nil, document, "", nil, []int{200, 201, 202}, []int{400, 404, 409, 307}, operationTimeout, retryTimeout)
	if err != nil {
		// This is a failure
		t.createCounter.failed++
		t.reportFailure(test.NewFailure("Failed to create document in collection '%s': %v", collectionName, err))
		return "", maskAny(err)
	}
	t.createCounter.succeeded++
	t.log.Infof("Creating document '%s' in '%s' succeeded", key, collectionName)
	return update.Rev, nil
}
