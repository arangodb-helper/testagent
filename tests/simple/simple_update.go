package simple

import (
	"fmt"
	"net/url"
	"time"

	"github.com/arangodb-helper/testagent/service/test"
)

// updateExistingDocument updates an existing document with an optional explicit revision.
// The operation is expected to succeed.


/**
 *  20x, if document was replaced
 *  400, if body bad
 *  404, if document did not exist
 *  409, if unique constraint would be violated
 *  412, if if-match was given and document already there
 *  timeout, in which case the document might or might not have been written
 *    after this, either one of the 5 things might have happened,
 *    or nothing might have happened at all,
 *    or any of this might still happen in the future
 *  connection refused with coordinator ==> simply try again with another
 *  connection error ("broken pipe") with coordinator, to be treated like 
 *    a timeout
 *  503, cluster internal mishap, all bets off
 *  If first request gives correct result: OK
 *  if wrong result: ERROR  (include 503 in this case)
 *  if connection refused to coordinator: simply retry other
 *  if either timeout (or broken pipe with coordinator):
 *    try to read the document repeatedly for up to 15s:
 *      if new document there: treat as if op had worked
 *    else (new version not appeared within 15s):
 *      treat as if op has not worked and go to retry loop
 */

func (t *simpleTest) updateExistingDocument(c *collection, key, rev string) (string, error) {
	operationTimeout, retryTimeout := t.OperationTimeout, t.RetryTimeout
	q := url.Values{}
	q.Set("waitForSync", "true")
	newName := fmt.Sprintf("Updated name %s", time.Now())
	hdr, ifMatchStatus, explicitRev := createRandomIfMatchHeader(nil, rev)
	t.log.Infof("Updating existing document '%s' (%s) in '%s' (name -> '%s')...", key, ifMatchStatus, c.name, newName)
	delta := map[string]interface{}{
		"name": newName,
	}
	doc := c.existingDocs[key]
	update, err := t.client.Patch(fmt.Sprintf("/_api/document/%s/%s", c.name, key), q,
		hdr, delta, "", nil, []int{200, 201, 202, 412}, []int{400, 404, 307}, operationTimeout, retryTimeout)
	if err != nil {
		// This is a failure
		t.lastRequestErr = true
		t.updateExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to update existing document '%s' (%s) in collection '%s': %v", key, ifMatchStatus, c.name, err))
		return "", maskAny(err)
	} else if update.StatusCode == 412 {
		if explicitRev || t.lastRequestErr {
		t.lastRequestErr = false
			// Expected revision did NOT match.
			// This may happen when a coordinator succeeds in the first attempt but we've already timed out.
			// Check document against what we expect after the update.
			expected := c.existingDocs[key]
			expected.Name = newName
			if match, rev, err := t.isDocumentEqualTo(c, key, expected); err != nil {
				// Failed to read document. This is a failure.
				t.updateExistingCounter.failed++
				t.reportFailure(test.NewFailure("Failed to read existing document '%s' (%s) in collection '%s' that should have been updated: %v", key, ifMatchStatus, c.name, err))
				return "", maskAny(err)
			} else if !match {
				// The document does not match what we expect after the update. This is a failure.
				t.updateExistingCounter.failed++
				t.reportFailure(test.NewFailure("Failed to update existing document '%s' (%s) in collection '%s': got 412 but has not been updated", key, ifMatchStatus, c.name))
				return "", maskAny(fmt.Errorf("Failed to update existing document '%s' (%s) in collection '%s': got 412 but has not been updated", key, ifMatchStatus, c.name))
			} else {
				// Match found, document has been updated after all
				update.Rev = rev
			}
		} else {
			// We got a 412 without asking for an explicit revision. This is a failure.
			t.lastRequestErr = false
			t.updateExistingCounter.failed++
			t.reportFailure(test.NewFailure("Failed to update existing document '%s' (%s) in collection '%s': got 412 but did not set If-Match", key, ifMatchStatus, c.name))
			return "", maskAny(fmt.Errorf("Failed to update existing document '%s' (%s) in collection '%s': got 412 but did not set If-Match", key, ifMatchStatus, c.name))
		}
	}
	// Update internal doc
	doc.Name = newName
	doc.rev = update.Rev
	c.existingDocs[key] = doc
	t.updateExistingCounter.succeeded++
	t.log.Infof("Updating existing document '%s' (%s) in '%s' (name -> '%s') succeeded", key, ifMatchStatus, c.name, newName)
	return update.Rev, nil
}

// updateExistingDocumentWrongRevision updates an existing document with an explicit wrong revision.
// The operation is expected to fail.
func (t *simpleTest) updateExistingDocumentWrongRevision(collectionName string, key, rev string) error {
	operationTimeout, retryTimeout := t.OperationTimeout, t.RetryTimeout
	q := url.Values{}
	q.Set("waitForSync", "true")
	newName := fmt.Sprintf("Updated name %s", time.Now())
	hdr := ifMatchHeader(nil, rev)
	t.log.Infof("Updating existing document '%s' wrong revision in '%s' (name -> '%s')...", key, collectionName, newName)
	delta := map[string]interface{}{
		"name": newName,
	}
	_, err := t.client.Patch(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, hdr, delta, "", nil, []int{412}, []int{200, 201, 202, 400, 404, 307}, operationTimeout, retryTimeout)
	if err != nil {
		// This is a failure
		t.updateExistingWrongRevisionCounter.failed++
		t.reportFailure(test.NewFailure("Failed to update existing document '%s' wrong revision in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	t.updateExistingWrongRevisionCounter.succeeded++
	t.log.Infof("Updating existing document '%s' wrong revision in '%s' (name -> '%s') succeeded", key, collectionName, newName)
	return nil
}

// updateNonExistingDocument updates a non-existing document.
// The operation is expected to fail.
func (t *simpleTest) updateNonExistingDocument(collectionName string, key string) error {
	operationTimeout, retryTimeout := t.OperationTimeout, t.RetryTimeout
	q := url.Values{}
	q.Set("waitForSync", "true")
	newName := fmt.Sprintf("Updated non-existing name %s", time.Now())
	t.log.Infof("Updating non-existing document '%s' in '%s' (name -> '%s')...", key, collectionName, newName)
	delta := map[string]interface{}{
		"name": newName,
	}
	if _, err := t.client.Patch(fmt.Sprintf("/_api/document/%s/%s", collectionName, key), q, nil, delta, "", nil, []int{404}, []int{200, 201, 202, 400, 412, 307}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.updateNonExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to update non-existing document '%s' in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	t.updateNonExistingCounter.succeeded++
	t.log.Infof("Updating non-existing document '%s' in '%s' (name -> '%s') succeeded", key, collectionName, newName)
	return nil
}
