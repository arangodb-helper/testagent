package simple

import (
	"fmt"
	"net/url"
	"time"

	"github.com/arangodb/testAgent/service/test"
)

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
		t.reportFailure(test.NewFailure("Failed to create document with key '%s' in collection '%s': %v", key, collectionName, err))
		return "", maskAny(err)
	}
	t.createCounter.succeeded++
	t.log.Infof("Creating document '%s' in '%s' succeeded", key, collectionName)
	return update.Rev, nil
}
