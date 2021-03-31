package simple

import (
	"fmt"
	"net/url"
	"time"

	"github.com/arangodb-helper/testagent/service/test"
)

// removeExistingDocument removes an existing document with an optional explicit revision.
// The operation is expected to succeed.
func (t *simpleTest) removeExistingDocument(collectionName string, key, rev string) error {

	operationTimeout := t.OperationTimeout
	testTimeout := time.Now().Add(operationTimeout * 4)

	q := url.Values{}
	q.Set("waitForSync", "true")
	hdr, ifMatchStatus, _ := createRandomIfMatchHeader(nil, rev)

	t.log.Infof("Removing existing document '%s' (%s) from '%s'...", key, ifMatchStatus, collectionName)

	backoff := time.Millisecond * 250
	i := 0

	for true {

		i++
		if time.Now().After(testTimeout) {
			break;
		}

		checkRetry := false
		success := false
		resp, err := t.client.Delete(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, hdr, []int{0, 200, 201, 202, 404, 503}, []int{400, 412, 307}, operationTimeout, 1)

		if err[0] == nil { // we have a response
			if resp[0].StatusCode == 503 || resp[0].StatusCode == 0 {
				// 0, 503 and 409 -> check if accidentally successful
				checkRetry = true
			} else if resp[0].StatusCode == 404 {
				if i == 1 {
					// Not enough attempts, this is a failure
					t.deleteExistingCounter.failed++
					t.reportFailure(
						test.NewFailure(
							"Failed to delete existing document '%s' (%s) in collection '%s': got 404 after only 1 attempt",
							key, ifMatchStatus, collectionName))
					return maskAny(
						fmt.Errorf(
							"Failed to delete existing document '%s' (%s) in collection '%s': got 404 after only 1 attempt",
							key, ifMatchStatus, collectionName))
				} else {
					success = true
				}
			}
		} else {
			t.deleteExistingCounter.failed++
			t.reportFailure(
				test.NewFailure(
					"Failed to delete existing document '%s' (%s) in collection '%s': %v",
					key, ifMatchStatus, collectionName, err[0]))
			return maskAny(err[0])
		}

		if checkRetry {
			d, e := readDocument(t, collectionName, key, "", 120, false)
			if e == nil {
				if d == nil {
					success = true
				}
			}
		}

		if success {
			t.deleteExistingCounter.succeeded++
			t.log.Infof("Removing existing document '%s' (%s) from '%s' succeeded", key, ifMatchStatus, collectionName)
			return nil
		}

		time.Sleep(backoff)
		backoff += backoff

	}

	// Overall timeout :(
	t.reportFailure(
		test.NewFailure("Timed out while trying to create(%i) document %s in %s.", i, key, collectionName))
	return maskAny(fmt.Errorf("Timed out while trying to create(%i) document %s in %s.", i, key, collectionName))

}

// removeExistingDocumentWrongRevision removes an existing document with an explicit wrong revision.
// The operation is expected to fail.
func (t *simpleTest) removeExistingDocumentWrongRevision(collectionName string, key, rev string) error {
	operationTimeout := t.OperationTimeout
	q := url.Values{}
	q.Set("waitForSync", "true")
	hdr := ifMatchHeader(nil, rev)
	t.log.Infof("Removing existing document '%s' wrong revision from '%s'...", key, collectionName)
	if _, err := t.client.Delete(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, hdr, []int{412}, []int{200, 201, 202, 400, 404, 307}, operationTimeout, 1); err[0] != nil {
		// This is a failure
		t.deleteExistingWrongRevisionCounter.failed++
		t.reportFailure(test.NewFailure("Failed to delete existing document '%s' wrong revision in collection '%s': %v", key, collectionName, err[0]))
		return maskAny(err[0])
	}
	t.deleteExistingWrongRevisionCounter.succeeded++
	t.log.Infof("Removing existing document '%s' wrong revision from '%s' succeeded", key, collectionName)
	return nil
}

// removeNonExistingDocument removes a non-existing document.
// The operation is expected to fail.
func (t *simpleTest) removeNonExistingDocument(collectionName string, key string) error {
	operationTimeout := t.OperationTimeout
	q := url.Values{}
	q.Set("waitForSync", "true")
	t.log.Infof("Removing non-existing document '%s' from '%s'...", key, collectionName)
	if _, err := t.client.Delete(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, nil, []int{404}, []int{200, 201, 202, 400, 412, 307}, operationTimeout, 1); err[0] != nil {
		// This is a failure
		t.deleteNonExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to delete non-existing document '%s' in collection '%s': %v", key, collectionName, err[0]))
		return maskAny(err[0])
	}
	t.deleteNonExistingCounter.succeeded++
	t.log.Infof("Removing non-existing document '%s' from '%s' succeeded", key, collectionName)
	return nil
}
