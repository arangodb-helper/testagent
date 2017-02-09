package simple

import (
	"fmt"
	"net/url"
	"time"

	"github.com/arangodb/testAgent/service/test"
)

// removeExistingDocument removes an existing document with an optional explicit revision.
// The operation is expected to succeed.
func (t *simpleTest) removeExistingDocument(collectionName string, key, rev string) error {
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
	q := url.Values{}
	q.Set("waitForSync", "true")
	hdr, ifMatchStatus := createRandomIfMatchHeader(nil, rev)
	t.log.Infof("Removing existing document '%s' (%s) from '%s'...", key, ifMatchStatus, collectionName)
	if _, err := t.client.Delete(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, hdr, []int{200, 201, 202}, []int{400, 404, 412, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.deleteExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to delete existing document '%s' (%s) in collection '%s': %v", key, ifMatchStatus, collectionName, err))
		return maskAny(err)
	}
	t.deleteExistingCounter.succeeded++
	t.log.Infof("Removing existing document '%s' (%s) from '%s' succeeded", key, ifMatchStatus, collectionName)
	return nil
}

// removeExistingDocumentWrongRevision removes an existing document with an explicit wrong revision.
// The operation is expected to fail.
func (t *simpleTest) removeExistingDocumentWrongRevision(collectionName string, key, rev string) error {
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
	q := url.Values{}
	q.Set("waitForSync", "true")
	hdr := ifMatchHeader(nil, rev)
	t.log.Infof("Removing existing document '%s' wrong revision from '%s'...", key, collectionName)
	if _, err := t.client.Delete(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, hdr, []int{412}, []int{200, 201, 202, 400, 404, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.deleteExistingWrongRevisionCounter.failed++
		t.reportFailure(test.NewFailure("Failed to delete existing document '%s' wrong revision in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	t.deleteExistingWrongRevisionCounter.succeeded++
	t.log.Infof("Removing existing document '%s' wrong revision from '%s' succeeded", key, collectionName)
	return nil
}

// removeNonExistingDocument removes a non-existing document.
// The operation is expected to fail.
func (t *simpleTest) removeNonExistingDocument(collectionName string, key string) error {
	operationTimeout, retryTimeout := time.Minute/4, time.Minute
	q := url.Values{}
	q.Set("waitForSync", "true")
	t.log.Infof("Removing non-existing document '%s' from '%s'...", key, collectionName)
	if _, err := t.client.Delete(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, nil, []int{404}, []int{200, 201, 202, 400, 412, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.deleteNonExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to delete non-existing document '%s' in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	t.deleteNonExistingCounter.succeeded++
	t.log.Infof("Removing non-existing document '%s' from '%s' succeeded", key, collectionName)
	return nil
}
